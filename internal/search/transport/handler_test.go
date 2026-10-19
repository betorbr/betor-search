package transport

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"betor-search/internal/search/application"
)

func TestHandler_SearchReturnsResults(t *testing.T) {
	service := application.NewServiceWithItems([]application.Item{
		{ID: "1", ItemType: "movie", TorrentName: "Dune Part Two", ProviderSlug: "test", ProviderURL: "https://example.com/1", MagnetURI: "magnet:?xt=urn:btih:1", MagnetXT: "urn:btih:1", Languages: []string{"pt-BR"}, InsertedAt: time.Now()},
		{ID: "2", ItemType: "tv", TorrentName: "The Office", ProviderSlug: "test", ProviderURL: "https://example.com/2", MagnetURI: "magnet:?xt=urn:btih:2", MagnetXT: "urn:btih:2", Languages: []string{"en"}, InsertedAt: time.Now()},
	}, "", 30*time.Minute)
	handler := NewHandler(&service)

	req := httptest.NewRequest(http.MethodGet, "/v1/search/?q=Dune&page=1&size=10", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var payload struct {
		Items []application.Item `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(payload.Items))
	}
	if payload.Items[0].TorrentName != "Dune Part Two" {
		t.Fatalf("title = %q, want Dune Part Two", payload.Items[0].TorrentName)
	}
}

func TestHandler_SearchReadsLiveServiceState(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"items_total": 1,
			"generated_at": "2026-10-19T22:00:27.586036Z",
			"items": [{
				"id": "2",
				"provider_slug": "test",
				"provider_url": "https://example.com/2",
				"item_type": "movie",
				"magnet_uri": "magnet:?xt=urn:btih:2",
				"magnet_xt": "urn:btih:2",
				"torrent_name": "Fresh Title",
				"inserted_at": "2026-10-19T22:00:27.586036Z",
				"updated_at": "2026-10-19T22:00:27.586036Z"
			}]
		}`))
	}))
	defer server.Close()

	service := application.NewService(server.URL, 30*time.Minute)
	handler := NewHandler(&service)

	if err := service.Sync(); err != nil {
		t.Fatalf("sync failed: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/search/?q=Fresh&page=1&size=10", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	var payload struct {
		Items []application.Item `json:"items"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(payload.Items) != 1 {
		t.Fatalf("len(items) = %d, want 1", len(payload.Items))
	}
	if payload.Items[0].TorrentName != "Fresh Title" {
		t.Fatalf("title = %q, want Fresh Title", payload.Items[0].TorrentName)
	}
}
