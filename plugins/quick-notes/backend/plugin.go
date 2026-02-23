package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alvarotorresc/cortex/pkg/sdk"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

// QuickNotesPlugin implements sdk.CortexPlugin for note-taking.
type QuickNotesPlugin struct {
	db *sql.DB
}

// GetManifest returns the plugin's metadata.
func (p *QuickNotesPlugin) GetManifest() (*sdk.Manifest, error) {
	return &sdk.Manifest{
		ID:          "quick-notes",
		Name:        "Quick Notes",
		Version:     "0.1.0",
		Description: "Capture ideas and notes quickly, local and private",
		Icon:        "notebook-pen",
		Color:       "#6366F1",
		Permissions: []string{"db:read", "db:write"},
	}, nil
}

// Migrate opens the SQLite database and runs embedded SQL migrations.
func (p *QuickNotesPlugin) Migrate(databasePath string) error {
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	p.db = database

	// Enable WAL mode for better concurrent read performance.
	if _, err := p.db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("enabling WAL mode: %w", err)
	}

	migrationSQL, err := migrations.ReadFile("migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("reading migration: %w", err)
	}

	if _, err := p.db.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("running migration: %w", err)
	}

	return nil
}

// HandleAPI routes incoming API requests to the appropriate handler.
func (p *QuickNotesPlugin) HandleAPI(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	switch {
	case req.Method == "GET" && req.Path == "/notes":
		return p.listNotes()
	case req.Method == "POST" && req.Path == "/notes":
		return p.createNote(req)
	case req.Method == "PUT" && matchPath(req.Path, "/notes/", "/pin"):
		return p.togglePin(req)
	case req.Method == "PUT" && strings.HasPrefix(req.Path, "/notes/"):
		return p.updateNote(req)
	case req.Method == "DELETE" && strings.HasPrefix(req.Path, "/notes/"):
		return p.deleteNote(req)
	default:
		return jsonError(404, "NOT_FOUND", "route not found")
	}
}

// GetWidgetData returns dashboard widget data for the requested slot.
func (p *QuickNotesPlugin) GetWidgetData(slot string) ([]byte, error) {
	if slot != "dashboard-widget" {
		return json.Marshal(map[string]interface{}{"data": nil})
	}

	// Get latest 3 notes
	rows, err := p.db.Query(
		`SELECT id, title, content, pinned, created_at, updated_at
		 FROM notes
		 ORDER BY pinned DESC, updated_at DESC
		 LIMIT 3`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying latest notes: %w", err)
	}
	defer rows.Close()

	latestNotes := make([]Note, 0, 3)
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Pinned, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning note: %w", err)
		}
		latestNotes = append(latestNotes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating notes: %w", err)
	}

	// Count pinned notes
	var pinnedCount int
	row := p.db.QueryRow("SELECT COUNT(*) FROM notes WHERE pinned = 1")
	if err := row.Scan(&pinnedCount); err != nil {
		return nil, fmt.Errorf("counting pinned notes: %w", err)
	}

	return json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"latest":       latestNotes,
			"pinned_count": pinnedCount,
		},
	})
}

// Teardown closes the database connection when the plugin is unloaded.
func (p *QuickNotesPlugin) Teardown() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}

// --- Data types ---

// Note represents a user note.
type Note struct {
	ID        int64  `json:"id"`
	Title     string `json:"title"`
	Content   string `json:"content"`
	Pinned    bool   `json:"pinned"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// --- Handlers ---

func (p *QuickNotesPlugin) listNotes() (*sdk.APIResponse, error) {
	rows, err := p.db.Query(
		`SELECT id, title, content, pinned, created_at, updated_at
		 FROM notes
		 ORDER BY pinned DESC, updated_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying notes: %w", err)
	}
	defer rows.Close()

	notes := make([]Note, 0)
	for rows.Next() {
		var n Note
		if err := rows.Scan(&n.ID, &n.Title, &n.Content, &n.Pinned, &n.CreatedAt, &n.UpdatedAt); err != nil {
			return nil, fmt.Errorf("scanning note: %w", err)
		}
		notes = append(notes, n)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating notes: %w", err)
	}

	return jsonSuccess(200, notes)
}

func (p *QuickNotesPlugin) createNote(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.Unmarshal(req.Body, &input); err != nil {
		return jsonError(400, "VALIDATION_ERROR", "invalid JSON body")
	}

	if strings.TrimSpace(input.Title) == "" {
		return jsonError(400, "VALIDATION_ERROR", "title is required")
	}

	result, err := p.db.Exec(
		"INSERT INTO notes (title, content) VALUES (?, ?)",
		input.Title, input.Content,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting note: %w", err)
	}

	id, _ := result.LastInsertId()
	return jsonSuccess(201, map[string]interface{}{"id": id})
}

func (p *QuickNotesPlugin) updateNote(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id := extractID(req.Path, "/notes/")
	if id == "" {
		return jsonError(400, "VALIDATION_ERROR", "missing note ID")
	}

	var input struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}

	if err := json.Unmarshal(req.Body, &input); err != nil {
		return jsonError(400, "VALIDATION_ERROR", "invalid JSON body")
	}

	if strings.TrimSpace(input.Title) == "" {
		return jsonError(400, "VALIDATION_ERROR", "title is required")
	}

	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	result, err := p.db.Exec(
		"UPDATE notes SET title = ?, content = ?, updated_at = ? WHERE id = ?",
		input.Title, input.Content, now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("updating note: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return jsonError(404, "NOT_FOUND", "note not found")
	}

	return jsonSuccess(200, map[string]interface{}{"id": id, "updated_at": now})
}

func (p *QuickNotesPlugin) deleteNote(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	id := extractID(req.Path, "/notes/")
	if id == "" {
		return jsonError(400, "VALIDATION_ERROR", "missing note ID")
	}

	result, err := p.db.Exec("DELETE FROM notes WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("deleting note: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return jsonError(404, "NOT_FOUND", "note not found")
	}

	return jsonSuccess(200, map[string]interface{}{"deleted": id})
}

func (p *QuickNotesPlugin) togglePin(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	// Path: /notes/{id}/pin
	parts := strings.Split(strings.TrimPrefix(req.Path, "/"), "/")
	if len(parts) < 3 || parts[1] == "" {
		return jsonError(400, "VALIDATION_ERROR", "missing note ID")
	}
	id := parts[1]

	// Toggle the pinned state: 0 -> 1, 1 -> 0
	now := time.Now().UTC().Format("2006-01-02 15:04:05")
	result, err := p.db.Exec(
		"UPDATE notes SET pinned = CASE WHEN pinned = 0 THEN 1 ELSE 0 END, updated_at = ? WHERE id = ?",
		now, id,
	)
	if err != nil {
		return nil, fmt.Errorf("toggling pin: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return jsonError(404, "NOT_FOUND", "note not found")
	}

	// Read the new state
	var pinned bool
	row := p.db.QueryRow("SELECT pinned FROM notes WHERE id = ?", id)
	if err := row.Scan(&pinned); err != nil {
		return nil, fmt.Errorf("reading pin state: %w", err)
	}

	return jsonSuccess(200, map[string]interface{}{"id": id, "pinned": pinned})
}

// --- Helpers ---

// matchPath checks if path matches a pattern like "/notes/{id}/pin".
func matchPath(path string, prefix string, suffix string) bool {
	if !strings.HasPrefix(path, prefix) {
		return false
	}
	if !strings.HasSuffix(path, suffix) {
		return false
	}
	// Ensure there is content between prefix and suffix
	middle := strings.TrimSuffix(strings.TrimPrefix(path, prefix), suffix)
	return middle != ""
}

// extractID extracts the ID segment from a path like "/notes/{id}".
func extractID(path string, prefix string) string {
	trimmed := strings.TrimPrefix(path, prefix)
	// Handle paths like "/notes/123" and "/notes/123/"
	parts := strings.SplitN(trimmed, "/", 2)
	if len(parts) == 0 || parts[0] == "" {
		return ""
	}
	return parts[0]
}

// --- JSON response helpers ---

// jsonSuccess wraps data in `{ "data": ... }` format per PATTERNS.md.
func jsonSuccess(status int, data interface{}) (*sdk.APIResponse, error) {
	body, err := json.Marshal(map[string]interface{}{"data": data})
	if err != nil {
		return nil, fmt.Errorf("marshaling response: %w", err)
	}
	return &sdk.APIResponse{
		StatusCode:  status,
		Body:        body,
		ContentType: "application/json",
	}, nil
}

// jsonError wraps errors in `{ "error": { "code": ..., "message": ... } }` format per PATTERNS.md.
func jsonError(status int, code string, message string) (*sdk.APIResponse, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
	return &sdk.APIResponse{
		StatusCode:  status,
		Body:        body,
		ContentType: "application/json",
	}, nil
}
