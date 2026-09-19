package transport

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"betor-search/internal/health/application"
)

func TestHandler_GetHealthReturnsComponentStatus(t *testing.T) {
	handler := NewHandler(application.NewService())
	handler.now = func() time.Time {
		return time.Date(2026, time.September, 19, 12, 0, 0, 0, time.UTC)
	}

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Content-Type"); got != "application/json" {
		t.Fatalf("content-type = %q, want application/json", got)
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Fatalf("allow = %q, want GET, HEAD", got)
	}

	var response map[string]json.RawMessage
	if err := json.NewDecoder(bytes.NewReader(rec.Body.Bytes())).Decode(&response); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if _, ok := response["status"]; !ok {
		t.Fatal("missing status field")
	}
	if _, ok := response["components"]; !ok {
		t.Fatal("missing components field")
	}
	if _, ok := response["timestamp"]; !ok {
		t.Fatal("missing timestamp field")
	}
	if rec.Body.String() == "" {
		t.Fatal("empty body")
	}
}

func TestHandler_HeadHealthReturns200WithoutBody(t *testing.T) {
	handler := NewHandler(application.NewService())
	handler.now = func() time.Time {
		return time.Date(2026, time.September, 19, 12, 0, 0, 0, time.UTC)
	}

	req := httptest.NewRequest(http.MethodHead, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if rec.Body.Len() != 0 {
		t.Fatalf("body len = %d, want 0", rec.Body.Len())
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Fatalf("allow = %q, want GET, HEAD", got)
	}
}

func TestHandler_UnsupportedMethodReturns405(t *testing.T) {
	handler := NewHandler(application.NewService())

	req := httptest.NewRequest(http.MethodPost, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusMethodNotAllowed {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusMethodNotAllowed)
	}
	if got := rec.Header().Get("Allow"); got != "GET, HEAD" {
		t.Fatalf("allow = %q, want GET, HEAD", got)
	}
}
