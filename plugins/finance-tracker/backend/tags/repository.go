package tags

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Repository handles database operations for tags.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a Repository backed by the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List returns all tags ordered by name.
func (r *Repository) List() ([]Tag, error) {
	rows, err := r.db.Query(
		`SELECT id, name, COALESCE(color, '') FROM tags ORDER BY name`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying tags: %w", err)
	}
	defer rows.Close()

	tags := make([]Tag, 0)
	for rows.Next() {
		var t Tag
		if err := rows.Scan(&t.ID, &t.Name, &t.Color); err != nil {
			return nil, fmt.Errorf("scanning tag row: %w", err)
		}
		tags = append(tags, t)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating tag rows: %w", err)
	}
	return tags, nil
}

// Create inserts a new tag and returns the generated ID.
func (r *Repository) Create(input *CreateTagInput) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO tags (name, color) VALUES (?, ?)`,
		input.Name, input.Color,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, shared.NewConflictError(
				fmt.Sprintf("tag '%s' already exists", input.Name),
			)
		}
		return 0, fmt.Errorf("inserting tag: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}
	return id, nil
}

// Update modifies an existing tag's fields.
func (r *Repository) Update(id int64, input *UpdateTagInput) error {
	result, err := r.db.Exec(
		`UPDATE tags SET name = ?, color = ? WHERE id = ?`,
		input.Name, input.Color, id,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return shared.NewConflictError(
				fmt.Sprintf("tag '%s' already exists", input.Name),
			)
		}
		return fmt.Errorf("updating tag: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("tag", fmt.Sprintf("%d", id))
	}
	return nil
}

// Delete removes a tag by its ID. The CASCADE on transaction_tags handles cleanup.
func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM tags WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting tag: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("tag", fmt.Sprintf("%d", id))
	}
	return nil
}
