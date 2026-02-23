package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/alvarotorresc/cortex/internal/plugin"
)

// newPluginRouter creates a minimal chi router with only plugin routes registered.
// This avoids needing a full config, host DB, or middleware for focused tests.
func newPluginRouter(registry *plugin.Registry) *chi.Mux {
	router := chi.NewRouter()
	pluginAPIRoutes(router, registry)
	return router
}

func TestListPlugins_Empty(t *testing.T) {
	registry := plugin.NewRegistry()
	router := newPluginRouter(registry)

	req := httptest.NewRequest(http.MethodGet, "/api/plugins", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	dataRaw, ok := body["data"]
	if !ok {
		t.Fatal("response missing 'data' key")
	}

	// An empty registry should return an empty array or null.
	var items []json.RawMessage
	if err := json.Unmarshal(dataRaw, &items); err != nil {
		// data could be null for an empty list; that is also acceptable.
		if string(dataRaw) != "null" {
			t.Fatalf("expected data to be an array or null, got: %s", string(dataRaw))
		}
		return
	}

	if len(items) != 0 {
		t.Errorf("expected 0 plugins, got %d", len(items))
	}
}

func TestListPlugins_WithPlugins(t *testing.T) {
	registry := plugin.NewRegistry()
	registry.Register("alpha", nil, &plugin.Manifest{ID: "alpha", Name: "Alpha", Version: "1.0.0"})
	registry.Register("beta", nil, &plugin.Manifest{ID: "beta", Name: "Beta", Version: "2.0.0"})

	router := newPluginRouter(registry)

	req := httptest.NewRequest(http.MethodGet, "/api/plugins", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", rec.Code)
	}

	var body struct {
		Data []plugin.Manifest `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if len(body.Data) != 2 {
		t.Fatalf("expected 2 plugins, got %d", len(body.Data))
	}

	// Verify both IDs are present (order may vary due to map iteration).
	ids := make(map[string]bool)
	for _, m := range body.Data {
		ids[m.ID] = true
	}
	if !ids["alpha"] || !ids["beta"] {
		t.Errorf("expected plugins 'alpha' and 'beta', got IDs: %v", ids)
	}
}

func TestPluginWidget_NotFound(t *testing.T) {
	registry := plugin.NewRegistry()
	router := newPluginRouter(registry)

	req := httptest.NewRequest(http.MethodGet, "/api/plugins/unknown/widget/dashboard-widget", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse error body: %v", err)
	}

	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("expected error code 'NOT_FOUND', got '%s'", body.Error.Code)
	}
}

func TestPluginProxy_NotFound(t *testing.T) {
	registry := plugin.NewRegistry()
	router := newPluginRouter(registry)

	req := httptest.NewRequest(http.MethodGet, "/api/plugins/unknown/anything", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected status 404, got %d", rec.Code)
	}

	contentType := rec.Header().Get("Content-Type")
	if contentType != "application/json" {
		t.Errorf("expected Content-Type 'application/json', got '%s'", contentType)
	}

	var body struct {
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse error body: %v", err)
	}

	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("expected error code 'NOT_FOUND', got '%s'", body.Error.Code)
	}
}
