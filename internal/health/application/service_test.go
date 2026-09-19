package application

import (
	"testing"
	"time"
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
