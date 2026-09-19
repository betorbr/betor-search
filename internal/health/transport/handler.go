package transport

import (
	"encoding/json"
	"net/http"
	"time"

	"betor-search/internal/health/application"
)

type Handler struct {
	service application.Service
	now     func() time.Time
}

func NewHandler(service application.Service) Handler {
	return Handler{
		service: service,
		now:     time.Now,
	}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		response := h.service.Response(h.now())
		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Allow", "GET, HEAD")
		w.WriteHeader(http.StatusOK)
		if r.Method == http.MethodHead {
			return
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		}
	default:
		w.Header().Set("Allow", "GET, HEAD")
		w.WriteHeader(http.StatusMethodNotAllowed)
	}
}
