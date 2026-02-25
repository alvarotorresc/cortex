package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/alvarotorresc/cortex/internal/db"
)

// newDashboardRouter creates a minimal chi router with only dashboard routes registered.
func newDashboardRouter(t *testing.T) (*chi.Mux, *db.HostDB) {
	t.Helper()

	tempDir := t.TempDir()
	hostDB, err := db.NewHostDB(tempDir)
	if err != nil {
		t.Fatalf("failed to create host DB: %v", err)
	}
	t.Cleanup(func() { hostDB.Close() })

	router := chi.NewRouter()
	dashboardRoutes(router, hostDB)
	return router, hostDB
}

func TestGetDashboardLayout_Empty(t *testing.T) {
	router, _ := newDashboardRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/api/dashboard/layout", nil)
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

	// Empty dashboard should return an empty array or null.
	var items []json.RawMessage
	if err := json.Unmarshal(dataRaw, &items); err != nil {
		if string(dataRaw) != "null" {
			t.Fatalf("expected data to be an array or null, got: %s", string(dataRaw))
		}
		return
	}

	if len(items) != 0 {
		t.Errorf("expected 0 layouts, got %d", len(items))
	}
}

func TestSaveDashboardLayout(t *testing.T) {
	router, _ := newDashboardRouter(t)

	// PUT two widgets.
	putBody := `{
		"widgets": [
			{"widget_id": "finance-tracker:balance", "position_x": 0, "position_y": 0, "width": 4, "height": 2},
			{"widget_id": "quick-notes:notes", "position_x": 4, "position_y": 0, "width": 4, "height": 3}
		]
	}`

	putReq := httptest.NewRequest(http.MethodPut, "/api/dashboard/layout", strings.NewReader(putBody))
	putReq.Header.Set("Content-Type", "application/json")
	putRec := httptest.NewRecorder()

	router.ServeHTTP(putRec, putReq)

	if putRec.Code != http.StatusOK {
		t.Fatalf("PUT expected status 200, got %d. Body: %s", putRec.Code, putRec.Body.String())
	}

	// Verify the PUT response contains the saved widgets.
	var putResponse struct {
		Data []db.WidgetLayout `json:"data"`
		Meta map[string]string `json:"meta"`
	}
	if err := json.Unmarshal(putRec.Body.Bytes(), &putResponse); err != nil {
		t.Fatalf("failed to parse PUT response: %v", err)
	}

	if len(putResponse.Data) != 2 {
		t.Fatalf("expected 2 widgets in PUT response, got %d", len(putResponse.Data))
	}

	if putResponse.Meta["saved_at"] == "" {
		t.Error("expected meta.saved_at to be set")
	}

	// GET to verify persistence.
	getReq := httptest.NewRequest(http.MethodGet, "/api/dashboard/layout", nil)
	getRec := httptest.NewRecorder()

	router.ServeHTTP(getRec, getReq)

	if getRec.Code != http.StatusOK {
		t.Fatalf("GET expected status 200, got %d", getRec.Code)
	}

	var getResponse struct {
		Data []db.WidgetLayout `json:"data"`
	}
	if err := json.Unmarshal(getRec.Body.Bytes(), &getResponse); err != nil {
		t.Fatalf("failed to parse GET response: %v", err)
	}

	if len(getResponse.Data) != 2 {
		t.Fatalf("expected 2 widgets from GET, got %d", len(getResponse.Data))
	}

	// Verify widget IDs are present.
	widgetIDs := make(map[string]bool)
	for _, w := range getResponse.Data {
		widgetIDs[w.WidgetID] = true
	}
	if !widgetIDs["finance-tracker:balance"] || !widgetIDs["quick-notes:notes"] {
		t.Errorf("expected widget IDs 'finance-tracker:balance' and 'quick-notes:notes', got: %v", widgetIDs)
	}
}

func TestSaveDashboardLayout_InvalidJSON(t *testing.T) {
	router, _ := newDashboardRouter(t)

	req := httptest.NewRequest(http.MethodPut, "/api/dashboard/layout", strings.NewReader("{invalid json"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()

	router.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected status 400, got %d", rec.Code)
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

	if body.Error.Code != "BAD_REQUEST" {
		t.Errorf("expected error code 'BAD_REQUEST', got '%s'", body.Error.Code)
	}
}
