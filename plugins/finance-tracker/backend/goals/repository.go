package goals

import (
	"database/sql"
	"fmt"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Repository handles database operations for savings goals.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a Repository backed by the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List returns all savings goals ordered by creation date (newest first).
func (r *Repository) List() ([]SavingsGoal, error) {
	rows, err := r.db.Query(`
		SELECT id, name, target_amount, current_amount,
		       target_date, COALESCE(icon, ''), COALESCE(color, ''),
		       is_completed, created_at
		FROM savings_goals
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying savings goals: %w", err)
	}
	defer rows.Close()

	return scanGoals(rows)
}

// GetByID returns a single savings goal by its ID.
func (r *Repository) GetByID(id int64) (*SavingsGoal, *shared.AppError) {
	var g SavingsGoal
	var targetDate sql.NullString
	var isCompleted int

	err := r.db.QueryRow(`
		SELECT id, name, target_amount, current_amount,
		       target_date, COALESCE(icon, ''), COALESCE(color, ''),
		       is_completed, created_at
		FROM savings_goals WHERE id = ?
	`, id).Scan(
		&g.ID, &g.Name, &g.TargetAmount, &g.CurrentAmount,
		&targetDate, &g.Icon, &g.Color,
		&isCompleted, &g.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, shared.NewNotFoundError("goal", fmt.Sprintf("%d", id))
	}
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying goal: %v", err), 500)
	}

	if targetDate.Valid {
		g.TargetDate = &targetDate.String
	}
	g.IsCompleted = isCompleted != 0

	return &g, nil
}

// Create inserts a new savings goal and returns the generated ID.
func (r *Repository) Create(input *CreateGoalInput) (int64, error) {
	var targetDate interface{}
	if input.TargetDate != nil {
		targetDate = *input.TargetDate
	}

	result, err := r.db.Exec(`
		INSERT INTO savings_goals (name, target_amount, icon, color, target_date)
		VALUES (?, ?, ?, ?, ?)
	`, input.Name, input.TargetAmount, input.Icon, input.Color, targetDate)
	if err != nil {
		return 0, fmt.Errorf("inserting savings goal: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}
	return id, nil
}

// Update modifies an existing savings goal's fields.
func (r *Repository) Update(id int64, input *UpdateGoalInput) error {
	var targetDate interface{}
	if input.TargetDate != nil {
		targetDate = *input.TargetDate
	}

	result, err := r.db.Exec(`
		UPDATE savings_goals
		SET name = ?, target_amount = ?, icon = ?, color = ?, target_date = ?
		WHERE id = ?
	`, input.Name, input.TargetAmount, input.Icon, input.Color, targetDate, id)
	if err != nil {
		return fmt.Errorf("updating savings goal: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("goal", fmt.Sprintf("%d", id))
	}
	return nil
}

// Delete removes a savings goal by its ID.
func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM savings_goals WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting savings goal: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("goal", fmt.Sprintf("%d", id))
	}
	return nil
}

// UpdateAmountAndCompletion atomically updates current_amount and is_completed.
func (r *Repository) UpdateAmountAndCompletion(id int64, newAmount float64, isCompleted bool) error {
	completed := 0
	if isCompleted {
		completed = 1
	}

	result, err := r.db.Exec(`
		UPDATE savings_goals
		SET current_amount = ?, is_completed = ?
		WHERE id = ?
	`, newAmount, completed, id)
	if err != nil {
		return fmt.Errorf("updating goal amount: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("goal", fmt.Sprintf("%d", id))
	}
	return nil
}

// scanGoals reads all rows from the result set into a slice of SavingsGoal.
func scanGoals(rows *sql.Rows) ([]SavingsGoal, error) {
	goals := make([]SavingsGoal, 0)
	for rows.Next() {
		var g SavingsGoal
		var targetDate sql.NullString
		var isCompleted int

		if err := rows.Scan(
			&g.ID, &g.Name, &g.TargetAmount, &g.CurrentAmount,
			&targetDate, &g.Icon, &g.Color,
			&isCompleted, &g.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning goal row: %w", err)
		}

		if targetDate.Valid {
			g.TargetDate = &targetDate.String
		}
		g.IsCompleted = isCompleted != 0

		goals = append(goals, g)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating goal rows: %w", err)
	}
	return goals, nil
}
