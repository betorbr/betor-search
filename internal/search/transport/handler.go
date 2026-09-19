package transport

import (
	"encoding/json"
	"net/http"
	"time"

	"betor-search/internal/search/application"
)

type Handler struct {
	service application.Service
	now     func() time.Time
}

func NewHandler(service application.Service) Handler {
	return Handler{service: service, now: time.Now}
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	filter := h.service.ParseHTTPQuery(r.URL.RawQuery)
	filter.Q = r.URL.Query().Get("q")
	results, err := h.service.Search(filter.Q, filter)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
