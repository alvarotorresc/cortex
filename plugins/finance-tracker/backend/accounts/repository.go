package accounts

import (
	"database/sql"
	"fmt"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Repository handles database operations for accounts.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a Repository backed by the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// ListActive returns all non-archived accounts ordered by id.
func (r *Repository) ListActive() ([]Account, error) {
	rows, err := r.db.Query(
		`SELECT id, name, type, currency, interest_rate, icon, color, is_archived, created_at
		 FROM accounts
		 WHERE is_archived = 0
		 ORDER BY id`,
	)
	if err != nil {
		return nil, fmt.Errorf("querying active accounts: %w", err)
	}
	defer rows.Close()

	return scanAccounts(rows)
}

// GetByID returns a single account by its ID.
func (r *Repository) GetByID(id int64) (*Account, *shared.AppError) {
	var account Account
	var isArchived int
	var interestRate sql.NullFloat64
	var icon, color sql.NullString

	err := r.db.QueryRow(
		`SELECT id, name, type, currency, interest_rate, icon, color, is_archived, created_at
		 FROM accounts WHERE id = ?`, id,
	).Scan(
		&account.ID, &account.Name, &account.Type, &account.Currency,
		&interestRate, &icon, &color, &isArchived, &account.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, shared.NewNotFoundError("account", fmt.Sprintf("%d", id))
	}
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying account: %v", err), 500)
	}

	account.IsArchived = isArchived == 1
	if interestRate.Valid {
		account.InterestRate = &interestRate.Float64
	}
	if icon.Valid {
		account.Icon = icon.String
	}
	if color.Valid {
		account.Color = color.String
	}
	return &account, nil
}

// Create inserts a new account and returns the generated ID.
func (r *Repository) Create(input *CreateAccountInput) (int64, error) {
	var interestRate interface{}
	if input.InterestRate != nil {
		interestRate = *input.InterestRate
	}

	result, err := r.db.Exec(
		`INSERT INTO accounts (name, type, currency, interest_rate, icon, color)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		input.Name, input.Type, input.Currency, interestRate, input.Icon, input.Color,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting account: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}
	return id, nil
}

// Update modifies an existing account's fields.
func (r *Repository) Update(id int64, input *UpdateAccountInput) error {
	var interestRate interface{}
	if input.InterestRate != nil {
		interestRate = *input.InterestRate
	}

	result, err := r.db.Exec(
		`UPDATE accounts SET name = ?, type = ?, currency = ?, interest_rate = ?, icon = ?, color = ?
		 WHERE id = ?`,
		input.Name, input.Type, input.Currency, interestRate, input.Icon, input.Color, id,
	)
	if err != nil {
		return fmt.Errorf("updating account: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("account", fmt.Sprintf("%d", id))
	}
	return nil
}

// Archive soft-deletes an account by setting is_archived to 1.
func (r *Repository) Archive(id int64) error {
	result, err := r.db.Exec(
		"UPDATE accounts SET is_archived = 1 WHERE id = ? AND is_archived = 0", id,
	)
	if err != nil {
		return fmt.Errorf("archiving account: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("account", fmt.Sprintf("%d", id))
	}
	return nil
}

// CalculateBalance computes the net balance for an account from all its
// transactions, considering income, expense, and transfer types.
func (r *Repository) CalculateBalance(accountID int64) (float64, error) {
	var balance float64
	err := r.db.QueryRow(
		`SELECT
			COALESCE(SUM(
				CASE
					WHEN type='income' AND account_id=? THEN amount
					WHEN type='transfer' AND dest_account_id=? THEN amount
					ELSE 0
				END
			), 0) -
			COALESCE(SUM(
				CASE
					WHEN type='expense' AND account_id=? THEN amount
					WHEN type='transfer' AND account_id=? THEN amount
					ELSE 0
				END
			), 0)
		 FROM transactions
		 WHERE account_id = ? OR dest_account_id = ?`,
		accountID, accountID, accountID, accountID, accountID, accountID,
	).Scan(&balance)
	if err != nil {
		return 0, fmt.Errorf("calculating balance for account %d: %w", accountID, err)
	}

	return balance, nil
}

// scanAccounts reads all rows from the result set into a slice of Account.
func scanAccounts(rows *sql.Rows) ([]Account, error) {
	accounts := make([]Account, 0)
	for rows.Next() {
		var a Account
		var isArchived int
		var interestRate sql.NullFloat64
		var icon, color sql.NullString

		if err := rows.Scan(
			&a.ID, &a.Name, &a.Type, &a.Currency,
			&interestRate, &icon, &color, &isArchived, &a.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning account row: %w", err)
		}

		a.IsArchived = isArchived == 1
		if interestRate.Valid {
			a.InterestRate = &interestRate.Float64
		}
		if icon.Valid {
			a.Icon = icon.String
		}
		if color.Valid {
			a.Color = color.String
		}
		accounts = append(accounts, a)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating account rows: %w", err)
	}
	return accounts, nil
}
