package application

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	searchApplication "betor-search/internal/search/application"
)

type fakeCatalog struct {
	status        string
	lastExecution time.Time
	lastSuccess   *time.Time
	lastError     string
}

func (f fakeCatalog) Status() string           { return f.status }
func (f fakeCatalog) LastExecution() time.Time { return f.lastExecution }
func (f fakeCatalog) LastSuccess() *time.Time  { return f.lastSuccess }
func (f fakeCatalog) LastError() string        { return f.lastError }

func TestService_ResponseReflectsCatalogComponentStatus(t *testing.T) {
	lastSuccess := time.Date(2026, time.September, 19, 16, 0, 0, 0, time.UTC)
	service := NewServiceWithCatalog("betor-search-catalog", fakeCatalog{
		status:        "DEGRADED",
		lastExecution: time.Date(2026, time.September, 19, 16, 1, 0, 0, time.UTC),
		lastSuccess:   &lastSuccess,
		lastError:     "temporary upstream timeout",
	})

	response := service.Response(time.Date(2026, time.September, 19, 16, 2, 0, 0, time.UTC))
	if response.Status != "DEGRADED" {
		t.Fatalf("status = %q, want DEGRADED", response.Status)
	}
	if len(response.Components) != 1 {
		t.Fatalf("components len = %d, want 1", len(response.Components))
	}
	if response.Components[0].Name != "betor-search-catalog" {
		t.Fatalf("component name = %q, want betor-search-catalog", response.Components[0].Name)
	}
	if response.Components[0].Status != "DEGRADED" {
		t.Fatalf("component status = %q, want DEGRADED", response.Components[0].Status)
	}
	serviceWithStartedAt := NewServiceWithCatalog("betor-search-catalog", fakeCatalog{
		status:        "UP",
		lastExecution: time.Date(2026, time.September, 19, 16, 1, 0, 0, time.UTC),
		lastSuccess:   &lastSuccess,
		lastError:     "",
	})
	responseWithStartedAt := serviceWithStartedAt.Response(time.Date(2026, time.September, 19, 16, 2, 0, 0, time.UTC))
	if responseWithStartedAt.StartedAt.IsZero() {
		t.Fatal("started_at should reflect application start time")
	}
}

func TestService_ResponseTracksLiveCatalogInstance(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"items":[{"id":"1","provider_slug":"test","provider_url":"https://example.com/1","item_type":"movie","magnet_uri":"magnet:?xt=urn:btih:1","magnet_xt":"urn:btih:1","torrent_name":"Movie A","inserted_at":"2024-01-01T00:00:00Z"}]}`))
	}))
	defer server.Close()

	catalog := searchApplication.NewService(server.URL, time.Minute)
	health := NewServiceWithCatalog("betor-search-catalog", &catalog)

	if err := catalog.Sync(); err != nil {
		t.Fatalf("catalog.Sync() returned error: %v", err)
	}

	response := health.Response(time.Now().UTC())
	if len(response.Components) != 1 {
		t.Fatalf("components len = %d, want 1", len(response.Components))
	}
	if response.Components[0].LastRun != catalog.LastExecution().UTC() {
		t.Fatalf("last_run_at = %s, want %s", response.Components[0].LastRun.Format(time.RFC3339Nano), catalog.LastExecution().UTC().Format(time.RFC3339Nano))
	}
	if response.Components[0].Updated != catalog.LastSuccess().UTC() {
		t.Fatalf("updated_at = %s, want %s", response.Components[0].Updated.Format(time.RFC3339Nano), catalog.LastSuccess().UTC().Format(time.RFC3339Nano))
	}
}
