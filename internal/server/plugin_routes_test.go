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
func newPluginRouter(t *testing.T, registry *plugin.Registry) *chi.Mux {
	t.Helper()

	tempDir := t.TempDir()
	loader := plugin.NewLoader(tempDir, tempDir, registry)

	router := chi.NewRouter()
	pluginAPIRoutes(router, registry, loader)
	return router
}

func TestListPlugins_Empty(t *testing.T) {
	registry := plugin.NewRegistry()
	router := newPluginRouter(t, registry)

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

	router := newPluginRouter(t, registry)

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
	router := newPluginRouter(t, registry)

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
	router := newPluginRouter(t, registry)

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

func TestUninstallPlugin_Success(t *testing.T) {
	registry := plugin.NewRegistry()
	registry.Register("alpha", nil, &plugin.Manifest{ID: "alpha", Name: "Alpha", Version: "1.0.0"})

	router := newPluginRouter(t, registry)

	req := httptest.NewRequest(http.MethodDelete, "/api/plugins/alpha", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d. Body: %s", rec.Code, rec.Body.String())
	}

	var body struct {
		Data struct {
			ID     string `json:"id"`
			Status string `json:"status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}

	if body.Data.ID != "alpha" {
		t.Errorf("expected data.id 'alpha', got '%s'", body.Data.ID)
	}
	if body.Data.Status != "uninstalled" {
		t.Errorf("expected data.status 'uninstalled', got '%s'", body.Data.Status)
	}

	// Verify plugin is no longer in the registry.
	if _, ok := registry.Get("alpha"); ok {
		t.Error("expected plugin 'alpha' to be removed from registry after uninstall")
	}
}

func TestUninstallPlugin_NotFound(t *testing.T) {
	registry := plugin.NewRegistry()
	router := newPluginRouter(t, registry)

	req := httptest.NewRequest(http.MethodDelete, "/api/plugins/nonexistent", nil)
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

func TestInstallPlugin_AlreadyInstalled(t *testing.T) {
	registry := plugin.NewRegistry()
	registry.Register("alpha", nil, &plugin.Manifest{ID: "alpha", Name: "Alpha", Version: "1.0.0"})

	router := newPluginRouter(t, registry)

	req := httptest.NewRequest(http.MethodPost, "/api/plugins/alpha/install", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected status 409, got %d. Body: %s", rec.Code, rec.Body.String())
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

	if body.Error.Code != "ALREADY_INSTALLED" {
		t.Errorf("expected error code 'ALREADY_INSTALLED', got '%s'", body.Error.Code)
	}
}

func TestInstallPlugin_NotOnDisk(t *testing.T) {
	registry := plugin.NewRegistry()
	router := newPluginRouter(t, registry)

	// Plugin "ghost" does not exist on disk, so LoadPlugin should fail.
	req := httptest.NewRequest(http.MethodPost, "/api/plugins/ghost/install", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d. Body: %s", rec.Code, rec.Body.String())
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

	if body.Error.Code != "INSTALL_ERROR" {
		t.Errorf("expected error code 'INSTALL_ERROR', got '%s'", body.Error.Code)
	}
}

func TestReloadPlugin_NotFound(t *testing.T) {
	registry := plugin.NewRegistry()
	router := newPluginRouter(t, registry)

	req := httptest.NewRequest(http.MethodPost, "/api/plugins/nonexistent/reload", nil)
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

func TestReloadPlugin_UnloadSucceedsLoadFails(t *testing.T) {
	registry := plugin.NewRegistry()
	// Register a plugin with nil client/plugin so UnloadPlugin succeeds (no teardown, no kill)
	// but LoadPlugin will fail because there is no plugin binary on disk.
	registry.Register("alpha", nil, &plugin.Manifest{ID: "alpha", Name: "Alpha", Version: "1.0.0"})

	router := newPluginRouter(t, registry)

	req := httptest.NewRequest(http.MethodPost, "/api/plugins/alpha/reload", nil)
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	// Unload succeeds, but Load fails (no binary on disk) -> 500 LOAD_ERROR
	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected status 500, got %d. Body: %s", rec.Code, rec.Body.String())
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

	if body.Error.Code != "LOAD_ERROR" {
		t.Errorf("expected error code 'LOAD_ERROR', got '%s'", body.Error.Code)
	}

	// Plugin should have been unloaded (removed from registry) even though reload failed.
	if _, ok := registry.Get("alpha"); ok {
		t.Error("expected plugin 'alpha' to be removed from registry after failed reload")
	}
}
