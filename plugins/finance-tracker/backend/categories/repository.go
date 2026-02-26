package categories

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Repository handles database operations for categories.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a Repository backed by the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List returns all categories, optionally filtered by type.
// When typeFilter is "income" or "expense", categories with that type or "both" are returned.
func (r *Repository) List(typeFilter string) ([]Category, error) {
	var rows *sql.Rows
	var err error

	if typeFilter != "" {
		rows, err = r.db.Query(
			`SELECT id, name, type, icon, color, is_default, sort_order
			 FROM categories
			 WHERE type = ? OR type = 'both'
			 ORDER BY sort_order, is_default DESC, name`,
			typeFilter,
		)
	} else {
		rows, err = r.db.Query(
			`SELECT id, name, type, icon, color, is_default, sort_order
			 FROM categories
			 ORDER BY sort_order, is_default DESC, name`,
		)
	}
	if err != nil {
		return nil, fmt.Errorf("querying categories: %w", err)
	}
	defer rows.Close()

	return scanCategories(rows)
}

// ExistsByName checks if a category with the given name already exists (case-insensitive).
func (r *Repository) ExistsByName(name string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM categories WHERE LOWER(name) = LOWER(?)", name,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking category existence: %w", err)
	}
	return count > 0, nil
}

// ExistsByNameExcluding checks if a category with the given name exists,
// excluding the category with the specified ID (for updates).
func (r *Repository) ExistsByNameExcluding(name string, excludeID int64) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM categories WHERE LOWER(name) = LOWER(?) AND id != ?",
		name, excludeID,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking category existence: %w", err)
	}
	return count > 0, nil
}

// GetByID returns a single category by its ID.
func (r *Repository) GetByID(id int64) (*Category, *shared.AppError) {
	var c Category
	var isDefault int
	var icon, color, categoryType sql.NullString

	err := r.db.QueryRow(
		`SELECT id, name, type, icon, color, is_default, sort_order
		 FROM categories WHERE id = ?`, id,
	).Scan(&c.ID, &c.Name, &categoryType, &icon, &color, &isDefault, &c.SortOrder)
	if err == sql.ErrNoRows {
		return nil, shared.NewNotFoundError("category", fmt.Sprintf("%d", id))
	}
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying category: %v", err), 500)
	}

	c.IsDefault = isDefault == 1
	if categoryType.Valid {
		c.Type = categoryType.String
	}
	if icon.Valid {
		c.Icon = icon.String
	}
	if color.Valid {
		c.Color = color.String
	}
	return &c, nil
}

// Create inserts a new category and returns the generated ID.
func (r *Repository) Create(input *CreateCategoryInput) (int64, error) {
	result, err := r.db.Exec(
		`INSERT INTO categories (name, type, icon, color)
		 VALUES (?, ?, ?, ?)`,
		input.Name, input.Type, input.Icon, input.Color,
	)
	if err != nil {
		// Check for UNIQUE constraint violation on name.
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return 0, shared.NewConflictError(
				fmt.Sprintf("category '%s' already exists", input.Name),
			)
		}
		return 0, fmt.Errorf("inserting category: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}
	return id, nil
}

// Update modifies an existing category's fields.
func (r *Repository) Update(id int64, input *UpdateCategoryInput) error {
	result, err := r.db.Exec(
		`UPDATE categories SET name = ?, type = ?, icon = ?, color = ?
		 WHERE id = ?`,
		input.Name, input.Type, input.Icon, input.Color, id,
	)
	if err != nil {
		if strings.Contains(err.Error(), "UNIQUE constraint failed") {
			return shared.NewConflictError(
				fmt.Sprintf("category '%s' already exists", input.Name),
			)
		}
		return fmt.Errorf("updating category: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("category", fmt.Sprintf("%d", id))
	}
	return nil
}

// HasTransactions checks if any transactions reference the given category name.
func (r *Repository) HasTransactions(categoryName string) (bool, error) {
	var count int
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM transactions WHERE category = ?", categoryName,
	).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking transactions for category: %w", err)
	}
	return count > 0, nil
}

// Delete removes a category by its ID.
func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM categories WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting category: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("category", fmt.Sprintf("%d", id))
	}
	return nil
}

// Reorder updates the sort_order for each category in the provided list.
// Uses a transaction to ensure atomicity.
func (r *Repository) Reorder(items []ReorderItem) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	stmt, err := tx.Prepare("UPDATE categories SET sort_order = ? WHERE id = ?")
	if err != nil {
		return fmt.Errorf("preparing reorder statement: %w", err)
	}
	defer stmt.Close()

	for _, item := range items {
		if _, err := stmt.Exec(item.SortOrder, item.ID); err != nil {
			return fmt.Errorf("updating sort_order for category %d: %w", item.ID, err)
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("committing reorder transaction: %w", err)
	}
	return nil
}

// scanCategories reads all rows from the result set into a slice of Category.
func scanCategories(rows *sql.Rows) ([]Category, error) {
	categories := make([]Category, 0)
	for rows.Next() {
		var c Category
		var isDefault int
		var icon, color, categoryType sql.NullString

		if err := rows.Scan(
			&c.ID, &c.Name, &categoryType, &icon, &color, &isDefault, &c.SortOrder,
		); err != nil {
			return nil, fmt.Errorf("scanning category row: %w", err)
		}

		c.IsDefault = isDefault == 1
		if categoryType.Valid {
			c.Type = categoryType.String
		}
		if icon.Valid {
			c.Icon = icon.String
		}
		if color.Valid {
			c.Color = color.String
		}
		categories = append(categories, c)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating category rows: %w", err)
	}
	return categories, nil
}
