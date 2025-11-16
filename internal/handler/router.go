package handler

import (
	"net/http"
	"path/filepath"
)

func SetupRouter(h *Handler) *http.ServeMux {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /api/v1/notify", h.CreateNotification)
	mux.HandleFunc("GET /api/v1/notify/", h.GetNotificationStatus)
	mux.HandleFunc("DELETE /api/v1/notify/", h.CancelNotification)
	mux.HandleFunc("GET /api/v1/notifications", h.ListNotifications)

	staticDir := "./static"
	mux.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir(staticDir))))

	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		http.ServeFile(w, r, filepath.Join(staticDir, "index.html"))
	})

	return mux
}
