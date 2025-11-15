package handler

import (
	"net/http"
)

func SetupRouter(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("POST /api/v1/notify", h.CreateNotification)
	mux.HandleFunc("GET /api/v1/notify/", h.GetNotificationStatus)
	mux.HandleFunc("DELETE /api/v1/notify/", h.CancelNotification)
	mux.HandleFunc("GET /api/v1/notifications", h.ListNotifications)
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("./static"))))
	return mux
}
