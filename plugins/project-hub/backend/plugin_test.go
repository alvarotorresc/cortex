package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"

	"github.com/alvarotorresc/cortex/pkg/sdk"
)

// newTestPlugin creates a ProjectHubPlugin with a migrated SQLite database in a temp directory.
// It returns the plugin ready for testing and calls t.Cleanup to close the database.
func newTestPlugin(t *testing.T) *ProjectHubPlugin {
	t.Helper()

	p := &ProjectHubPlugin{}
	dbPath := filepath.Join(t.TempDir(), "project_hub_test.db")

	if err := p.Migrate(dbPath); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	t.Cleanup(func() { p.Teardown() })
	return p
}

// parseDataObject parses an APIResponse body and returns the "data" field as raw JSON.
func parseDataObject(t *testing.T, resp *sdk.APIResponse) json.RawMessage {
	t.Helper()

	var body struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	return body.Data
}

// parseDataArray parses an APIResponse body and returns the "data" field as a JSON array.
func parseDataArray(t *testing.T, resp *sdk.APIResponse) []json.RawMessage {
	t.Helper()

	var body struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	return body.Data
}

// parseErrorResponse parses an error APIResponse body and returns the code and message.
func parseErrorResponse(t *testing.T, resp *sdk.APIResponse) (code string, message string) {
	t.Helper()

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to parse error body: %v", err)
	}
	return body.Error.Code, body.Error.Message
}

// --- Migration & Seed tests ---

func TestMigrate_CreatesTables(t *testing.T) {
	p := &ProjectHubPlugin{}
	dbPath := filepath.Join(t.TempDir(), "test_migrate.db")

	if err := p.Migrate(dbPath); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	defer p.Teardown()

	// Verify the projects table exists.
	rows, err := p.db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='projects'")
	if err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("projects table does not exist after migration")
	}

	// Verify the project_links table exists.
	rows2, err := p.db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='project_links'")
	if err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	defer rows2.Close()

	if !rows2.Next() {
		t.Fatal("project_links table does not exist after migration")
	}
}

func TestSeedData_LoadsProjects(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	items := parseDataArray(t, resp)
	// 7 flagship + 7 lab + 2 absorbed = 16
	if len(items) != 16 {
		t.Fatalf("expected 16 seeded projects, got %d", len(items))
	}
}

// --- CRUD Project tests ---

func TestCreateProject_Valid(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name": "Test Project", "tagline": "A test project", "status": "concept", "category": "lab", "stack": "Go, TypeScript"}`

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/projects",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse response to check slug.
	data := parseDataObject(t, resp)
	var result struct {
		ID   int64  `json:"id"`
		Slug string `json:"slug"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}

	if result.Slug != "test-project" {
		t.Errorf("expected slug 'test-project', got '%s'", result.Slug)
	}
}

func TestCreateProject_MissingName(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"tagline": "A test", "status": "concept", "category": "lab", "stack": "Go"}`

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/projects",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestCreateProject_InvalidStatus(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name": "Bad Status", "tagline": "Test", "status": "unknown", "category": "lab", "stack": "Go"}`

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/projects",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestCreateProject_DuplicateName(t *testing.T) {
	p := newTestPlugin(t)

	// "Cortex" already exists in seed data.
	body := `{"name": "Cortex", "tagline": "Duplicate", "status": "concept", "category": "lab", "stack": "Go"}`

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/projects",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 409 {
		t.Fatalf("expected status 409, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}
}

func TestGetProject_BySlug(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects/cortex",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	data := parseDataObject(t, resp)
	var proj ProjectWithLinksAndTags
	if err := json.Unmarshal(data, &proj); err != nil {
		t.Fatalf("failed to unmarshal project: %v", err)
	}

	if proj.Name != "Cortex" {
		t.Errorf("expected name 'Cortex', got '%s'", proj.Name)
	}
	if proj.Status != "development" {
		t.Errorf("expected status 'development', got '%s'", proj.Status)
	}
	if proj.Links == nil {
		t.Error("expected links to be non-nil (empty slice)")
	}
}

func TestGetProject_NotFound(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects/nonexistent",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 404 {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestUpdateProject(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"status": "active", "version": "v1.0.0"}`

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "PUT",
		Path:   "/projects/cortex",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify the update.
	getResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects/cortex",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("get after update returned error: %v", err)
	}

	data := parseDataObject(t, getResp)
	var proj Project
	if err := json.Unmarshal(data, &proj); err != nil {
		t.Fatalf("failed to unmarshal project: %v", err)
	}

	if proj.Status != "active" {
		t.Errorf("expected status 'active', got '%s'", proj.Status)
	}
	if proj.Version == nil || *proj.Version != "v1.0.0" {
		t.Errorf("expected version 'v1.0.0', got '%v'", proj.Version)
	}
}

func TestUpdateProject_NotFound(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"status": "active"}`

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "PUT",
		Path:   "/projects/nonexistent",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 404 {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestDeleteProject(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   "/projects/cortex",
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify it's gone.
	getResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects/cortex",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("get after delete returned error: %v", err)
	}

	if getResp.StatusCode != 404 {
		t.Fatalf("expected 404 after delete, got %d", getResp.StatusCode)
	}
}

func TestDeleteProject_NotFound(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   "/projects/nonexistent",
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 404 {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

// --- Filter tests ---

func TestListProjects_FilterByStatus(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects",
		Query:  map[string]string{"status": "active"},
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	items := parseDataArray(t, resp)
	// Active projects in seed: Sinherencia, create-astro-blog, PokeUtils, DevTools, Swiss Knife = 5
	if len(items) != 5 {
		t.Errorf("expected 5 active projects, got %d", len(items))
	}

	// Verify all are active.
	for _, raw := range items {
		var proj Project
		if err := json.Unmarshal(raw, &proj); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if proj.Status != "active" {
			t.Errorf("expected status 'active', got '%s' for project '%s'", proj.Status, proj.Name)
		}
	}
}

func TestListProjects_FilterByCategory(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects",
		Query:  map[string]string{"category": "flagship"},
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	items := parseDataArray(t, resp)
	if len(items) != 7 {
		t.Errorf("expected 7 flagship projects, got %d", len(items))
	}
}

func TestListProjects_SearchByName(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects",
		Query:  map[string]string{"search": "Swiss"},
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	items := parseDataArray(t, resp)
	if len(items) != 1 {
		t.Errorf("expected 1 project matching 'Swiss', got %d", len(items))
	}
}

func TestListProjects_CombinedFilters(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects",
		Query:  map[string]string{"status": "concept", "category": "flagship"},
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	items := parseDataArray(t, resp)
	// Concept + flagship: Guitar App, Libroteca = 2
	if len(items) != 2 {
		t.Errorf("expected 2 concept+flagship projects, got %d", len(items))
	}
}

// --- Link tests ---

func TestCreateLink(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"label": "Play Store", "url": "https://play.google.com/store/apps/details?id=com.example"}`

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/projects/cortex/links",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify link appears in project detail.
	getResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects/cortex",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("get project returned error: %v", err)
	}

	data := parseDataObject(t, getResp)
	var proj ProjectWithLinksAndTags
	if err := json.Unmarshal(data, &proj); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(proj.Links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(proj.Links))
	}
	if proj.Links[0].Label != "Play Store" {
		t.Errorf("expected label 'Play Store', got '%s'", proj.Links[0].Label)
	}
}

func TestCreateLink_MissingLabel(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"url": "https://example.com"}`

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/projects/cortex/links",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestUpdateLink(t *testing.T) {
	p := newTestPlugin(t)

	// Create a link first.
	createBody := `{"label": "Old Label", "url": "https://old.com"}`
	createResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/projects/cortex/links",
		Body:   []byte(createBody),
	})
	if err != nil {
		t.Fatalf("create link failed: %v", err)
	}

	data := parseDataObject(t, createResp)
	var createResult struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &createResult); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}

	// Update the link.
	updateBody := `{"label": "New Label", "url": "https://new.com"}`
	updateResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "PUT",
		Path:   fmt.Sprintf("/links/%d", createResult.ID),
		Body:   []byte(updateBody),
	})
	if err != nil {
		t.Fatalf("update link failed: %v", err)
	}

	if updateResp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d. Body: %s", updateResp.StatusCode, string(updateResp.Body))
	}

	// Verify the update took effect.
	getResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects/cortex",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("get project after link update failed: %v", err)
	}

	getData := parseDataObject(t, getResp)
	var proj ProjectWithLinksAndTags
	if err := json.Unmarshal(getData, &proj); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(proj.Links) != 1 {
		t.Fatalf("expected 1 link, got %d", len(proj.Links))
	}
	if proj.Links[0].Label != "New Label" {
		t.Errorf("expected label 'New Label', got '%s'", proj.Links[0].Label)
	}
	if proj.Links[0].URL != "https://new.com" {
		t.Errorf("expected url 'https://new.com', got '%s'", proj.Links[0].URL)
	}
}

func TestDeleteLink(t *testing.T) {
	p := newTestPlugin(t)

	// Create a link first.
	createBody := `{"label": "To Delete", "url": "https://delete.me"}`
	createResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/projects/cortex/links",
		Body:   []byte(createBody),
	})
	if err != nil {
		t.Fatalf("create link failed: %v", err)
	}

	data := parseDataObject(t, createResp)
	var createResult struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &createResult); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}

	// Delete the link.
	delResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   fmt.Sprintf("/links/%d", createResult.ID),
	})
	if err != nil {
		t.Fatalf("delete link failed: %v", err)
	}

	if delResp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", delResp.StatusCode)
	}

	// Verify the link is actually gone.
	getResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects/cortex",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("get project after link delete failed: %v", err)
	}

	getData := parseDataObject(t, getResp)
	var proj ProjectWithLinksAndTags
	if err := json.Unmarshal(getData, &proj); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(proj.Links) != 0 {
		t.Errorf("expected 0 links after delete, got %d", len(proj.Links))
	}
}

func TestDeleteProject_CascadesLinks(t *testing.T) {
	p := newTestPlugin(t)

	// Get project ID before deletion so we can verify cascade.
	var projectID int64
	err := p.db.QueryRow("SELECT id FROM projects WHERE slug = 'cortex'").Scan(&projectID)
	if err != nil {
		t.Fatalf("failed to get project ID: %v", err)
	}

	// Create a link on Cortex.
	linkBody := `{"label": "Test Link", "url": "https://test.com"}`
	_, err = p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/projects/cortex/links",
		Body:   []byte(linkBody),
	})
	if err != nil {
		t.Fatalf("create link failed: %v", err)
	}

	// Verify link exists before deletion.
	var linkCountBefore int
	err = p.db.QueryRow("SELECT COUNT(*) FROM project_links WHERE project_id = ?", projectID).Scan(&linkCountBefore)
	if err != nil {
		t.Fatalf("failed to count links before delete: %v", err)
	}
	if linkCountBefore != 1 {
		t.Fatalf("expected 1 link before delete, got %d", linkCountBefore)
	}

	// Delete the project.
	delResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   "/projects/cortex",
	})
	if err != nil {
		t.Fatalf("delete project failed: %v", err)
	}

	if delResp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", delResp.StatusCode)
	}

	// Verify links were cascade-deleted using the saved project ID.
	var linkCountAfter int
	err = p.db.QueryRow("SELECT COUNT(*) FROM project_links WHERE project_id = ?", projectID).Scan(&linkCountAfter)
	if err != nil {
		t.Fatalf("failed to count links after delete: %v", err)
	}
	if linkCountAfter != 0 {
		t.Errorf("expected 0 links after cascade delete, got %d", linkCountAfter)
	}
}

// --- Widget tests ---

func TestWidgetData_CountsByStatus(t *testing.T) {
	p := newTestPlugin(t)

	widgetData, err := p.GetWidgetData("dashboard-widget")
	if err != nil {
		t.Fatalf("GetWidgetData returned error: %v", err)
	}

	var widget struct {
		Data struct {
			Total    int            `json:"total"`
			ByStatus map[string]int `json:"by_status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(widgetData, &widget); err != nil {
		t.Fatalf("failed to parse widget data: %v", err)
	}

	if widget.Data.Total != 16 {
		t.Errorf("expected total 16, got %d", widget.Data.Total)
	}

	// Check specific counts from seed data.
	if widget.Data.ByStatus["active"] != 5 {
		t.Errorf("expected 5 active, got %d", widget.Data.ByStatus["active"])
	}
	if widget.Data.ByStatus["development"] != 2 {
		t.Errorf("expected 2 development, got %d", widget.Data.ByStatus["development"])
	}
	if widget.Data.ByStatus["design"] != 3 {
		t.Errorf("expected 3 design, got %d", widget.Data.ByStatus["design"])
	}
	if widget.Data.ByStatus["concept"] != 4 {
		t.Errorf("expected 4 concept, got %d", widget.Data.ByStatus["concept"])
	}
	if widget.Data.ByStatus["absorbed"] != 2 {
		t.Errorf("expected 2 absorbed, got %d", widget.Data.ByStatus["absorbed"])
	}
}

func TestWidgetData_UnknownSlot(t *testing.T) {
	p := newTestPlugin(t)

	widgetData, err := p.GetWidgetData("unknown-slot")
	if err != nil {
		t.Fatalf("GetWidgetData returned error: %v", err)
	}

	var result struct {
		Data interface{} `json:"data"`
	}
	if err := json.Unmarshal(widgetData, &result); err != nil {
		t.Fatalf("failed to parse widget data: %v", err)
	}

	if result.Data != nil {
		t.Errorf("expected nil data for unknown slot, got %v", result.Data)
	}
}

// --- Route tests ---

func TestRouteNotFound(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/nonexistent",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 404 {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got '%s'", code)
	}
}

// --- Additional validation tests ---

func TestCreateProject_MissingTagline(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name": "No Tagline", "status": "concept", "category": "lab", "stack": "Go"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/projects", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
	code, msg := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
	if msg != "tagline is required" {
		t.Errorf("expected 'tagline is required', got '%s'", msg)
	}
}

func TestCreateProject_InvalidCategory(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name": "Bad Cat", "tagline": "Test", "status": "concept", "category": "unknown", "stack": "Go"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/projects", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestCreateProject_MissingStack(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name": "No Stack", "tagline": "Test", "status": "concept", "category": "lab"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/projects", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestCreateProject_NameTooLong(t *testing.T) {
	p := newTestPlugin(t)

	longName := make([]byte, 101)
	for i := range longName {
		longName[i] = 'a'
	}
	body := fmt.Sprintf(`{"name": "%s", "tagline": "Test", "status": "concept", "category": "lab", "stack": "Go"}`, string(longName))
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/projects", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestCreateProject_InvalidColor(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name": "Bad Color", "tagline": "Test", "status": "concept", "category": "lab", "stack": "Go", "color": "not-hex"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/projects", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestCreateProject_InvalidJSON(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/projects", Body: []byte(`not json`)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestUpdateProject_NoFields(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "PUT", Path: "/projects/cortex", Body: []byte(`{}`)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}
}

func TestUpdateProject_InvalidStatus(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "PUT", Path: "/projects/cortex", Body: []byte(`{"status": "banana"}`)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestCreateLink_NonexistentProject(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"label": "Test", "url": "https://test.com"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/projects/nonexistent/links", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestCreateLink_MissingURL(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"label": "Missing URL"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/projects/cortex/links", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestDeleteLink_NotFound(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "DELETE", Path: "/links/99999"})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestUpdateLink_NotFound(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "PUT", Path: "/links/99999", Body: []byte(`{"label": "test"}`)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestGetManifest(t *testing.T) {
	p := newTestPlugin(t)

	manifest, err := p.GetManifest()
	if err != nil {
		t.Fatalf("GetManifest returned error: %v", err)
	}
	if manifest.ID != "project-hub" {
		t.Errorf("expected ID 'project-hub', got '%s'", manifest.ID)
	}
	if manifest.Version != "0.1.0" {
		t.Errorf("expected Version '0.1.0', got '%s'", manifest.Version)
	}
}

func TestWidgetData_AfterMutation(t *testing.T) {
	p := newTestPlugin(t)

	// Delete one active project to change counts.
	_, err := p.HandleAPI(&sdk.APIRequest{Method: "DELETE", Path: "/projects/sinherencia"})
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	widgetData, err := p.GetWidgetData("dashboard-widget")
	if err != nil {
		t.Fatalf("GetWidgetData returned error: %v", err)
	}

	var widget struct {
		Data struct {
			Total    int            `json:"total"`
			ByStatus map[string]int `json:"by_status"`
		} `json:"data"`
	}
	if err := json.Unmarshal(widgetData, &widget); err != nil {
		t.Fatalf("failed to parse widget data: %v", err)
	}

	if widget.Data.Total != 15 {
		t.Errorf("expected total 15 after deletion, got %d", widget.Data.Total)
	}
	if widget.Data.ByStatus["active"] != 4 {
		t.Errorf("expected 4 active after deletion, got %d", widget.Data.ByStatus["active"])
	}
}

func TestListProjects_EmptyResultReturnsArray(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects",
		Query:  map[string]string{"status": "archived"},
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	items := parseDataArray(t, resp)
	if items == nil {
		t.Error("expected non-nil empty array, got nil")
	}
	if len(items) != 0 {
		t.Errorf("expected 0 archived projects, got %d", len(items))
	}
}

func TestSearchWithLikeWildcards(t *testing.T) {
	p := newTestPlugin(t)

	// Search with % should not match everything.
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects",
		Query:  map[string]string{"search": "%"},
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	items := parseDataArray(t, resp)
	if len(items) != 0 {
		t.Errorf("expected 0 results when searching for literal '%%', got %d", len(items))
	}
}

// --- Slug generation tests ---

func TestToSlug(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"Test Project", "test-project"},
		{"Guitar App", "guitar-app"},
		{"create-astro-blog", "create-astro-blog"},
		{"Swiss Knife", "swiss-knife"},
		{"My  App!!!", "my-app"},
		{"  Spaces  ", "spaces"},
	}

	for _, tt := range tests {
		result := toSlug(tt.input)
		if result != tt.expected {
			t.Errorf("toSlug(%q) = %q, expected %q", tt.input, result, tt.expected)
		}
	}
}

// --- Tag tests ---

func TestSeedData_CreatesTags(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "GET", Path: "/tags", Query: map[string]string{}})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	items := parseDataArray(t, resp)
	// 33 seeded tags
	if len(items) < 30 {
		t.Errorf("expected at least 30 seeded tags, got %d", len(items))
	}
}

func TestSeedData_AssignsTagsToProjects(t *testing.T) {
	p := newTestPlugin(t)

	// Cortex should have Go, SvelteKit, gRPC, SQLite tags
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "GET", Path: "/projects/cortex", Query: map[string]string{}})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	data := parseDataObject(t, resp)
	var proj ProjectWithLinksAndTags
	if err := json.Unmarshal(data, &proj); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(proj.Tags) != 4 {
		t.Fatalf("expected 4 tags for Cortex, got %d", len(proj.Tags))
	}

	tagNames := make(map[string]bool)
	for _, tag := range proj.Tags {
		tagNames[tag.Name] = true
	}

	for _, expected := range []string{"Go", "SvelteKit", "gRPC", "SQLite"} {
		if !tagNames[expected] {
			t.Errorf("expected tag '%s' for Cortex, not found", expected)
		}
	}
}

func TestListProjects_IncludesTags(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "GET", Path: "/projects", Query: map[string]string{}})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	items := parseDataArray(t, resp)

	// Check that at least one project has tags.
	foundTags := false
	for _, raw := range items {
		var proj ProjectWithTags
		if err := json.Unmarshal(raw, &proj); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if len(proj.Tags) > 0 {
			foundTags = true
			break
		}
	}

	if !foundTags {
		t.Error("expected at least one project with tags in list response")
	}
}

func TestListProjects_FilterByTag(t *testing.T) {
	p := newTestPlugin(t)

	// Filter by "Go" tag should return Cortex and Clipboard Manager.
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/projects",
		Query:  map[string]string{"tag": "Go"},
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	items := parseDataArray(t, resp)
	if len(items) != 2 {
		t.Errorf("expected 2 projects with Go tag, got %d", len(items))
	}

	names := make(map[string]bool)
	for _, raw := range items {
		var proj Project
		if err := json.Unmarshal(raw, &proj); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		names[proj.Name] = true
	}

	if !names["Cortex"] {
		t.Error("expected Cortex in Go-tagged projects")
	}
	if !names["Clipboard Manager"] {
		t.Error("expected Clipboard Manager in Go-tagged projects")
	}
}

func TestCreateTag(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name": "Vue.js", "color": "#4FC08D"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/tags", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	data := parseDataObject(t, resp)
	var result struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if result.Name != "Vue.js" {
		t.Errorf("expected name 'Vue.js', got '%s'", result.Name)
	}
	if result.Color != "#4FC08D" {
		t.Errorf("expected color '#4FC08D', got '%s'", result.Color)
	}
}

func TestCreateTag_DuplicateName(t *testing.T) {
	p := newTestPlugin(t)

	// "Go" already exists in seed data.
	body := `{"name": "Go", "color": "#00ADD8"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/tags", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 409 {
		t.Fatalf("expected status 409, got %d", resp.StatusCode)
	}
}

func TestCreateTag_MissingName(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"color": "#FF0000"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/tags", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}
}

func TestDeleteTag(t *testing.T) {
	p := newTestPlugin(t)

	// Create a tag then delete it.
	createBody := `{"name": "ToDelete", "color": "#FF0000"}`
	createResp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/tags", Body: []byte(createBody)})
	if err != nil {
		t.Fatalf("create tag failed: %v", err)
	}

	data := parseDataObject(t, createResp)
	var result struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	delResp, err := p.HandleAPI(&sdk.APIRequest{Method: "DELETE", Path: fmt.Sprintf("/tags/%d", result.ID)})
	if err != nil {
		t.Fatalf("delete tag failed: %v", err)
	}

	if delResp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", delResp.StatusCode)
	}
}

func TestSetProjectTags(t *testing.T) {
	p := newTestPlugin(t)

	// Get tag IDs for React and TypeScript.
	var reactID, tsID int64
	err := p.db.QueryRow("SELECT id FROM tags WHERE name = 'React'").Scan(&reactID)
	if err != nil {
		t.Fatalf("failed to get React tag: %v", err)
	}
	err = p.db.QueryRow("SELECT id FROM tags WHERE name = 'TypeScript'").Scan(&tsID)
	if err != nil {
		t.Fatalf("failed to get TypeScript tag: %v", err)
	}

	// Set tags for Cortex to React + TypeScript (replacing Go, SvelteKit, gRPC, SQLite).
	body := fmt.Sprintf(`{"tag_ids": [%d, %d]}`, reactID, tsID)
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/projects/cortex/tags", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify tags changed.
	getResp, err := p.HandleAPI(&sdk.APIRequest{Method: "GET", Path: "/projects/cortex", Query: map[string]string{}})
	if err != nil {
		t.Fatalf("get project failed: %v", err)
	}

	getData := parseDataObject(t, getResp)
	var proj ProjectWithLinksAndTags
	if err := json.Unmarshal(getData, &proj); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(proj.Tags) != 2 {
		t.Fatalf("expected 2 tags after set, got %d", len(proj.Tags))
	}

	tagNames := make(map[string]bool)
	for _, tag := range proj.Tags {
		tagNames[tag.Name] = true
	}

	if !tagNames["React"] || !tagNames["TypeScript"] {
		t.Errorf("expected React and TypeScript tags, got %v", tagNames)
	}
}

func TestSetProjectTags_NonexistentProject(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"tag_ids": [1]}`
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/projects/nonexistent/tags", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 404 {
		t.Fatalf("expected status 404, got %d", resp.StatusCode)
	}
}

func TestCreateProject_WithTagIDs(t *testing.T) {
	p := newTestPlugin(t)

	// Get a tag ID.
	var goID int64
	err := p.db.QueryRow("SELECT id FROM tags WHERE name = 'Go'").Scan(&goID)
	if err != nil {
		t.Fatalf("failed to get Go tag: %v", err)
	}

	body := fmt.Sprintf(`{"name": "Tag Test", "tagline": "Test tags", "status": "concept", "category": "lab", "stack": "Go", "tag_ids": [%d]}`, goID)
	resp, err := p.HandleAPI(&sdk.APIRequest{Method: "POST", Path: "/projects", Body: []byte(body)})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify tags assigned.
	getResp, err := p.HandleAPI(&sdk.APIRequest{Method: "GET", Path: "/projects/tag-test", Query: map[string]string{}})
	if err != nil {
		t.Fatalf("get project failed: %v", err)
	}

	data := parseDataObject(t, getResp)
	var proj ProjectWithLinksAndTags
	if err := json.Unmarshal(data, &proj); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(proj.Tags) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(proj.Tags))
	}
	if proj.Tags[0].Name != "Go" {
		t.Errorf("expected tag 'Go', got '%s'", proj.Tags[0].Name)
	}
}
