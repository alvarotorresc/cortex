package db

import (
	"database/sql"
	"fmt"
	"path/filepath"

	_ "modernc.org/sqlite"
)

// WidgetLayout represents a widget's position and size on the dashboard grid.
type WidgetLayout struct {
	ID        int64  `json:"id"`
	WidgetID  string `json:"widget_id"`
	PositionX int    `json:"position_x"`
	PositionY int    `json:"position_y"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// HostDB manages the host-level SQLite database.
type HostDB struct {
	db *sql.DB
}

// NewHostDB opens the host database and runs migrations.
func NewHostDB(dataDir string) (*HostDB, error) {
	databasePath := filepath.Join(dataDir, "cortex.db")

	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("opening host database: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := database.Exec("PRAGMA journal_mode=WAL"); err != nil {
		database.Close()
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	hostDB := &HostDB{db: database}

	if err := hostDB.migrate(); err != nil {
		database.Close()
		return nil, fmt.Errorf("running host migrations: %w", err)
	}

	return hostDB, nil
}

// migrate creates host-level tables if they do not exist.
func (h *HostDB) migrate() error {
	query := `
		CREATE TABLE IF NOT EXISTS dashboard_layouts (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			widget_id TEXT NOT NULL UNIQUE,
			position_x INTEGER NOT NULL DEFAULT 0,
			position_y INTEGER NOT NULL DEFAULT 0,
			width INTEGER NOT NULL DEFAULT 4,
			height INTEGER NOT NULL DEFAULT 2,
			created_at TEXT NOT NULL DEFAULT (datetime('now')),
			updated_at TEXT NOT NULL DEFAULT (datetime('now'))
		);

		CREATE INDEX IF NOT EXISTS idx_dashboard_layouts_widget_id
			ON dashboard_layouts(widget_id);
	`
	_, err := h.db.Exec(query)
	return err
}

// GetDashboardLayouts returns all widget layouts from the dashboard.
func (h *HostDB) GetDashboardLayouts() ([]WidgetLayout, error) {
	query := `
		SELECT id, widget_id, position_x, position_y, width, height, created_at, updated_at
		FROM dashboard_layouts
		ORDER BY position_y, position_x
	`

	rows, err := h.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("querying dashboard layouts: %w", err)
	}
	defer rows.Close()

	var layouts []WidgetLayout
	for rows.Next() {
		var layout WidgetLayout
		if err := rows.Scan(
			&layout.ID,
			&layout.WidgetID,
			&layout.PositionX,
			&layout.PositionY,
			&layout.Width,
			&layout.Height,
			&layout.CreatedAt,
			&layout.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning dashboard layout: %w", err)
		}
		layouts = append(layouts, layout)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating dashboard layouts: %w", err)
	}

	return layouts, nil
}

// SaveDashboardLayouts replaces all widget layouts with the provided set.
// This is a full-replace operation within a transaction.
func (h *HostDB) SaveDashboardLayouts(layouts []WidgetLayout) error {
	transaction, err := h.db.Begin()
	if err != nil {
		return fmt.Errorf("starting transaction: %w", err)
	}
	defer transaction.Rollback()

	// Delete all existing layouts
	if _, err := transaction.Exec("DELETE FROM dashboard_layouts"); err != nil {
		return fmt.Errorf("clearing dashboard layouts: %w", err)
	}

	// Insert new layouts with parametrized queries
	insertQuery := `
		INSERT INTO dashboard_layouts (widget_id, position_x, position_y, width, height, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`

	statement, err := transaction.Prepare(insertQuery)
	if err != nil {
		return fmt.Errorf("preparing insert statement: %w", err)
	}
	defer statement.Close()

	for _, layout := range layouts {
		if _, err := statement.Exec(
			layout.WidgetID,
			layout.PositionX,
			layout.PositionY,
			layout.Width,
			layout.Height,
			layout.CreatedAt,
			layout.UpdatedAt,
		); err != nil {
			return fmt.Errorf("inserting layout for widget %s: %w", layout.WidgetID, err)
		}
	}

	if err := transaction.Commit(); err != nil {
		return fmt.Errorf("committing dashboard layouts: %w", err)
	}

	return nil
}

// Close closes the host database connection.
func (h *HostDB) Close() error {
	return h.db.Close()
}
