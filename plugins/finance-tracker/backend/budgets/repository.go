package budgets

import (
	"database/sql"
	"fmt"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Repository handles database operations for budgets.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a Repository backed by the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List returns all budgets for a given month, including recurring (month IS NULL or empty).
func (r *Repository) List(month string) ([]Budget, error) {
	rows, err := r.db.Query(`
		SELECT id, COALESCE(name, ''), COALESCE(category, ''), amount,
		       COALESCE(month, ''), created_at
		FROM budgets
		WHERE month = ? OR month IS NULL OR month = ''
		ORDER BY created_at DESC
	`, month)
	if err != nil {
		return nil, fmt.Errorf("querying budgets: %w", err)
	}
	defer rows.Close()

	return scanBudgets(rows)
}

// GetByID returns a single budget by its ID.
func (r *Repository) GetByID(id int64) (*Budget, *shared.AppError) {
	var b Budget
	var name, category, month sql.NullString

	err := r.db.QueryRow(`
		SELECT id, name, category, amount, month, created_at
		FROM budgets WHERE id = ?
	`, id).Scan(&b.ID, &name, &category, &b.Amount, &month, &b.CreatedAt)
	if err == sql.ErrNoRows {
		return nil, shared.NewNotFoundError("budget", fmt.Sprintf("%d", id))
	}
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying budget: %v", err), 500)
	}

	if name.Valid {
		b.Name = name.String
	}
	if category.Valid {
		b.Category = category.String
	}
	if month.Valid {
		b.Month = month.String
	}

	return &b, nil
}

// Create inserts a new budget and returns the generated ID.
func (r *Repository) Create(input *CreateBudgetInput) (int64, error) {
	var nameVal, categoryVal, monthVal interface{}
	if input.Name != "" {
		nameVal = input.Name
	}
	if input.Category != "" {
		categoryVal = input.Category
	}
	if input.Month != "" {
		monthVal = input.Month
	}

	result, err := r.db.Exec(`
		INSERT INTO budgets (name, category, amount, month)
		VALUES (?, ?, ?, ?)
	`, nameVal, categoryVal, input.Amount, monthVal)
	if err != nil {
		return 0, fmt.Errorf("inserting budget: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}
	return id, nil
}

// Update modifies an existing budget's fields.
func (r *Repository) Update(id int64, input *UpdateBudgetInput) error {
	var nameVal, categoryVal, monthVal interface{}
	if input.Name != "" {
		nameVal = input.Name
	}
	if input.Category != "" {
		categoryVal = input.Category
	}
	if input.Month != "" {
		monthVal = input.Month
	}

	result, err := r.db.Exec(`
		UPDATE budgets SET name = ?, category = ?, amount = ?, month = ?
		WHERE id = ?
	`, nameVal, categoryVal, input.Amount, monthVal, id)
	if err != nil {
		return fmt.Errorf("updating budget: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("budget", fmt.Sprintf("%d", id))
	}
	return nil
}

// Delete removes a budget by its ID.
func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM budgets WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting budget: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("budget", fmt.Sprintf("%d", id))
	}
	return nil
}

// CalculateSpentGlobal returns the total amount of all expense transactions in a given month.
func (r *Repository) CalculateSpentGlobal(month string) (float64, error) {
	var spent float64
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE type = 'expense' AND date LIKE ?
	`, month+"%").Scan(&spent)
	if err != nil {
		return 0, fmt.Errorf("calculating global spent: %w", err)
	}
	return spent, nil
}

// CalculateSpentByCategory returns the total amount of expense transactions
// for a specific category in a given month.
func (r *Repository) CalculateSpentByCategory(month string, category string) (float64, error) {
	var spent float64
	err := r.db.QueryRow(`
		SELECT COALESCE(SUM(amount), 0)
		FROM transactions
		WHERE type = 'expense' AND date LIKE ? AND category = ?
	`, month+"%", category).Scan(&spent)
	if err != nil {
		return 0, fmt.Errorf("calculating category spent: %w", err)
	}
	return spent, nil
}

// scanBudgets reads all rows from the result set into a slice of Budget.
func scanBudgets(rows *sql.Rows) ([]Budget, error) {
	budgets := make([]Budget, 0)
	for rows.Next() {
		var b Budget
		if err := rows.Scan(&b.ID, &b.Name, &b.Category, &b.Amount, &b.Month, &b.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning budget row: %w", err)
		}
		budgets = append(budgets, b)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating budget rows: %w", err)
	}
	return budgets, nil
}
