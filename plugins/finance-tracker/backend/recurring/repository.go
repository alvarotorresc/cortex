package recurring

import (
	"database/sql"
	"fmt"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Repository handles database operations for recurring rules.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a Repository backed by the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List returns all recurring rules ordered by creation date descending.
func (r *Repository) List() ([]Rule, error) {
	rows, err := r.db.Query(`
		SELECT id, amount, type, account_id, dest_account_id, category,
		       COALESCE(description, ''), frequency, day_of_month, day_of_week,
		       month_of_year, start_date, end_date, last_generated, is_active, created_at
		FROM recurring_rules
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying recurring rules: %w", err)
	}
	defer rows.Close()

	return scanRules(rows)
}

// GetByID returns a single recurring rule by its ID.
func (r *Repository) GetByID(id int64) (*Rule, *shared.AppError) {
	var rule Rule
	var destAccountID sql.NullInt64
	var dayOfMonth, dayOfWeek, monthOfYear sql.NullInt64
	var endDate, lastGenerated sql.NullString
	var isActive int
	var description sql.NullString

	err := r.db.QueryRow(`
		SELECT id, amount, type, account_id, dest_account_id, category, description,
		       frequency, day_of_month, day_of_week, month_of_year,
		       start_date, end_date, last_generated, is_active, created_at
		FROM recurring_rules WHERE id = ?
	`, id).Scan(
		&rule.ID, &rule.Amount, &rule.Type, &rule.AccountID, &destAccountID,
		&rule.Category, &description, &rule.Frequency, &dayOfMonth, &dayOfWeek,
		&monthOfYear, &rule.StartDate, &endDate, &lastGenerated, &isActive, &rule.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, shared.NewNotFoundError("recurring rule", fmt.Sprintf("%d", id))
	}
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying recurring rule: %v", err), 500)
	}

	rule.IsActive = isActive == 1
	if description.Valid {
		rule.Description = description.String
	}
	if destAccountID.Valid {
		v := destAccountID.Int64
		rule.DestAccountID = &v
	}
	if dayOfMonth.Valid {
		v := int(dayOfMonth.Int64)
		rule.DayOfMonth = &v
	}
	if dayOfWeek.Valid {
		v := int(dayOfWeek.Int64)
		rule.DayOfWeek = &v
	}
	if monthOfYear.Valid {
		v := int(monthOfYear.Int64)
		rule.MonthOfYear = &v
	}
	if endDate.Valid {
		rule.EndDate = endDate.String
	}
	if lastGenerated.Valid {
		rule.LastGenerated = lastGenerated.String
	}

	return &rule, nil
}

// Create inserts a new recurring rule and returns the generated ID.
func (r *Repository) Create(input *CreateRuleInput) (int64, error) {
	var destAccountID interface{}
	if input.DestAccountID != nil {
		destAccountID = *input.DestAccountID
	}

	var dayOfMonth, dayOfWeek, monthOfYear interface{}
	if input.DayOfMonth != nil {
		dayOfMonth = *input.DayOfMonth
	}
	if input.DayOfWeek != nil {
		dayOfWeek = *input.DayOfWeek
	}
	if input.MonthOfYear != nil {
		monthOfYear = *input.MonthOfYear
	}

	var endDate interface{}
	if input.EndDate != "" {
		endDate = input.EndDate
	}

	result, err := r.db.Exec(`
		INSERT INTO recurring_rules
			(amount, type, account_id, dest_account_id, category, description,
			 frequency, day_of_month, day_of_week, month_of_year, start_date, end_date)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		input.Amount, input.Type, *input.AccountID, destAccountID,
		input.Category, input.Description, input.Frequency,
		dayOfMonth, dayOfWeek, monthOfYear, input.StartDate, endDate,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting recurring rule: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}
	return id, nil
}

// Update modifies an existing recurring rule's fields.
func (r *Repository) Update(id int64, input *UpdateRuleInput) error {
	var destAccountID interface{}
	if input.DestAccountID != nil {
		destAccountID = *input.DestAccountID
	}

	var dayOfMonth, dayOfWeek, monthOfYear interface{}
	if input.DayOfMonth != nil {
		dayOfMonth = *input.DayOfMonth
	}
	if input.DayOfWeek != nil {
		dayOfWeek = *input.DayOfWeek
	}
	if input.MonthOfYear != nil {
		monthOfYear = *input.MonthOfYear
	}

	var endDate interface{}
	if input.EndDate != "" {
		endDate = input.EndDate
	}

	var accountID int64 = 1
	if input.AccountID != nil {
		accountID = *input.AccountID
	}

	result, err := r.db.Exec(`
		UPDATE recurring_rules SET
			amount = ?, type = ?, account_id = ?, dest_account_id = ?,
			category = ?, description = ?, frequency = ?,
			day_of_month = ?, day_of_week = ?, month_of_year = ?,
			start_date = ?, end_date = ?
		WHERE id = ?
	`,
		input.Amount, input.Type, accountID, destAccountID,
		input.Category, input.Description, input.Frequency,
		dayOfMonth, dayOfWeek, monthOfYear, input.StartDate, endDate, id,
	)
	if err != nil {
		return fmt.Errorf("updating recurring rule: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("recurring rule", fmt.Sprintf("%d", id))
	}
	return nil
}

// Deactivate sets is_active=0 for a rule (soft delete).
func (r *Repository) Deactivate(id int64) error {
	result, err := r.db.Exec("UPDATE recurring_rules SET is_active = 0 WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deactivating recurring rule: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("recurring rule", fmt.Sprintf("%d", id))
	}
	return nil
}

// ListActiveRules returns all active rules where last_generated is NULL or before the given date.
func (r *Repository) ListActiveRules(today string) ([]Rule, error) {
	rows, err := r.db.Query(`
		SELECT id, amount, type, account_id, dest_account_id, category,
		       COALESCE(description, ''), frequency, day_of_month, day_of_week,
		       month_of_year, start_date, end_date, last_generated, is_active, created_at
		FROM recurring_rules
		WHERE is_active = 1
		  AND (last_generated IS NULL OR last_generated < ?)
	`, today)
	if err != nil {
		return nil, fmt.Errorf("querying active rules: %w", err)
	}
	defer rows.Close()

	return scanRules(rows)
}

// InsertGeneratedTransaction inserts a transaction marked as a recurring instance.
func (r *Repository) InsertGeneratedTransaction(ruleID int64, amount float64, txType string,
	accountID int64, destAccountID *int64, category string, description string, date string) error {

	var destAcct interface{}
	if destAccountID != nil {
		destAcct = *destAccountID
	}

	_, err := r.db.Exec(`
		INSERT INTO transactions
			(amount, type, account_id, dest_account_id, category, description, date,
			 is_recurring_instance, recurring_rule_id)
		VALUES (?, ?, ?, ?, ?, ?, ?, 1, ?)
	`, amount, txType, accountID, destAcct, category, description, date, ruleID)
	if err != nil {
		return fmt.Errorf("inserting generated transaction: %w", err)
	}
	return nil
}

// UpdateLastGenerated sets the last_generated date for a rule.
func (r *Repository) UpdateLastGenerated(id int64, date string) error {
	_, err := r.db.Exec("UPDATE recurring_rules SET last_generated = ? WHERE id = ?", date, id)
	if err != nil {
		return fmt.Errorf("updating last_generated: %w", err)
	}
	return nil
}

// DeactivateRule sets is_active=0 for a rule (used when end_date has passed).
func (r *Repository) DeactivateRule(id int64) error {
	_, err := r.db.Exec("UPDATE recurring_rules SET is_active = 0 WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deactivating rule: %w", err)
	}
	return nil
}

// TransactionExistsForDate checks if a generated transaction already exists for the rule on the given date.
func (r *Repository) TransactionExistsForDate(ruleID int64, date string) (bool, error) {
	var count int
	err := r.db.QueryRow(`
		SELECT COUNT(*) FROM transactions
		WHERE recurring_rule_id = ? AND date = ? AND is_recurring_instance = 1
	`, ruleID, date).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking existing transaction: %w", err)
	}
	return count > 0, nil
}

// AccountExists checks whether an account with the given ID exists.
func (r *Repository) AccountExists(id int64) (bool, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM accounts WHERE id = ?", id).Scan(&count)
	if err != nil {
		return false, fmt.Errorf("checking account existence: %w", err)
	}
	return count > 0, nil
}

// scanRules reads all rows from the result set into a slice of Rule.
func scanRules(rows *sql.Rows) ([]Rule, error) {
	rules := make([]Rule, 0)
	for rows.Next() {
		var rule Rule
		var destAccountID sql.NullInt64
		var dayOfMonth, dayOfWeek, monthOfYear sql.NullInt64
		var endDate, lastGenerated sql.NullString
		var isActive int

		if err := rows.Scan(
			&rule.ID, &rule.Amount, &rule.Type, &rule.AccountID, &destAccountID,
			&rule.Category, &rule.Description, &rule.Frequency, &dayOfMonth, &dayOfWeek,
			&monthOfYear, &rule.StartDate, &endDate, &lastGenerated, &isActive, &rule.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning recurring rule row: %w", err)
		}

		rule.IsActive = isActive == 1
		if destAccountID.Valid {
			v := destAccountID.Int64
			rule.DestAccountID = &v
		}
		if dayOfMonth.Valid {
			v := int(dayOfMonth.Int64)
			rule.DayOfMonth = &v
		}
		if dayOfWeek.Valid {
			v := int(dayOfWeek.Int64)
			rule.DayOfWeek = &v
		}
		if monthOfYear.Valid {
			v := int(monthOfYear.Int64)
			rule.MonthOfYear = &v
		}
		if endDate.Valid {
			rule.EndDate = endDate.String
		}
		if lastGenerated.Valid {
			rule.LastGenerated = lastGenerated.String
		}

		rules = append(rules, rule)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating recurring rule rows: %w", err)
	}
	return rules, nil
}
