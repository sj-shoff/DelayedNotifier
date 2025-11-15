package handler

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"delayed-notifier/internal/domain"
	"delayed-notifier/internal/handler/dto"

	"github.com/go-playground/validator/v10"
	"github.com/wb-go/wbf/zlog"
)

type Handler struct {
	service  NotificationService
	validate *validator.Validate
}

func NewHandler(service NotificationService) *Handler {
	validate := validator.New()
	validate.RegisterValidation("datetime", func(fl validator.FieldLevel) bool {
		_, err := time.Parse(time.RFC3339, fl.Field().String())
		return err == nil
	})
	return &Handler{
		service:  service,
		validate: validate,
	}
}

func (h *Handler) CreateNotification(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	defer r.Body.Close()
	if err := h.validate.Struct(req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	notification, err := dto.ToDomain(req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	result, err := h.service.CreateNotification(ctx, notification)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to create notification")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := dto.FromDomain(result)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) GetNotificationStatus(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/notify/")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	status, err := h.service.GetNotificationStatus(ctx, id)
	if err != nil {
		if err == domain.ErrNotFound {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		zlog.Logger.Error().Err(err).Str("id", id).Msg("Failed to get notification status")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	resp := dto.StatusResponse{
		ID:     id,
		Status: string(status),
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func (h *Handler) CancelNotification(w http.ResponseWriter, r *http.Request) {
	id := strings.TrimPrefix(r.URL.Path, "/api/v1/notify/")
	if id == "" {
		http.Error(w, "ID is required", http.StatusBadRequest)
		return
	}
	ctx := r.Context()
	if err := h.service.CancelNotification(ctx, id); err != nil {
		if err == domain.ErrNotFound || err == domain.ErrCannotCancel {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		zlog.Logger.Error().Err(err).Str("id", id).Msg("Failed to cancel notification")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"message": "notification cancelled successfully"})
}

func (h *Handler) ListNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	notifications, err := h.service.ListNotifications(ctx)
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("Failed to list notifications")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	var resp []dto.NotificationResponse
	for _, n := range notifications {
		resp = append(resp, dto.FromDomain(n))
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
