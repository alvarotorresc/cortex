package shared

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// ExtractIDFromPath parses a numeric resource ID from the second segment of a
// URL path. For example, "/transactions/42" returns 42.
func ExtractIDFromPath(path string) (int64, *AppError) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 2 || parts[1] == "" {
		return 0, NewValidationError("missing resource ID in path")
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, NewValidationError("invalid resource ID: must be a number")
	}
	return id, nil
}

// OpenDatabase opens a SQLite database at the given path with WAL mode and
// foreign keys enabled. The caller is responsible for closing the database.
func OpenDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}
	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}
	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}
	return db, nil
}
