package investments

import (
	"database/sql"
	"fmt"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Repository handles database operations for investments.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a Repository backed by the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List returns all investments ordered by creation date (newest first).
func (r *Repository) List() ([]Investment, error) {
	rows, err := r.db.Query(`
		SELECT id, name, type, account_id, units, avg_buy_price, current_price,
		       currency, COALESCE(notes, ''), last_updated, created_at
		FROM investments
		ORDER BY created_at DESC
	`)
	if err != nil {
		return nil, fmt.Errorf("querying investments: %w", err)
	}
	defer rows.Close()

	return scanInvestments(rows)
}

// GetByID returns a single investment by its ID.
func (r *Repository) GetByID(id int64) (*Investment, *shared.AppError) {
	var inv Investment
	var accountID sql.NullInt64
	var units, avgBuyPrice, currentPrice sql.NullFloat64
	var lastUpdated sql.NullString

	err := r.db.QueryRow(`
		SELECT id, name, type, account_id, units, avg_buy_price, current_price,
		       currency, COALESCE(notes, ''), last_updated, created_at
		FROM investments WHERE id = ?
	`, id).Scan(
		&inv.ID, &inv.Name, &inv.Type, &accountID, &units, &avgBuyPrice, &currentPrice,
		&inv.Currency, &inv.Notes, &lastUpdated, &inv.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, shared.NewNotFoundError("investment", fmt.Sprintf("%d", id))
	}
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying investment: %v", err), 500)
	}

	applyNullables(&inv, accountID, units, avgBuyPrice, currentPrice, lastUpdated)
	return &inv, nil
}

// Create inserts a new investment and returns the generated ID.
func (r *Repository) Create(input *CreateInvestmentInput) (int64, error) {
	result, err := r.db.Exec(`
		INSERT INTO investments (name, type, account_id, units, avg_buy_price, current_price,
		                         currency, notes, last_updated)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, input.Name, input.Type, ptrToNullInt64(input.AccountID),
		ptrToNullFloat64(input.Units), ptrToNullFloat64(input.AvgBuyPrice),
		ptrToNullFloat64(input.CurrentPrice), input.Currency, input.Notes,
		ptrToNullString(input.LastUpdated))
	if err != nil {
		return 0, fmt.Errorf("inserting investment: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}
	return id, nil
}

// Update modifies an existing investment's fields.
func (r *Repository) Update(id int64, input *UpdateInvestmentInput) error {
	result, err := r.db.Exec(`
		UPDATE investments
		SET name = ?, type = ?, account_id = ?, units = ?, avg_buy_price = ?,
		    current_price = ?, currency = ?, notes = ?, last_updated = ?
		WHERE id = ?
	`, input.Name, input.Type, ptrToNullInt64(input.AccountID),
		ptrToNullFloat64(input.Units), ptrToNullFloat64(input.AvgBuyPrice),
		ptrToNullFloat64(input.CurrentPrice), input.Currency, input.Notes,
		ptrToNullString(input.LastUpdated), id)
	if err != nil {
		return fmt.Errorf("updating investment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("investment", fmt.Sprintf("%d", id))
	}
	return nil
}

// Delete removes an investment by its ID.
func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM investments WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting investment: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("investment", fmt.Sprintf("%d", id))
	}
	return nil
}

// scanInvestments reads all rows from the result set into a slice of Investment.
func scanInvestments(rows *sql.Rows) ([]Investment, error) {
	investments := make([]Investment, 0)
	for rows.Next() {
		var inv Investment
		var accountID sql.NullInt64
		var units, avgBuyPrice, currentPrice sql.NullFloat64
		var lastUpdated sql.NullString

		if err := rows.Scan(
			&inv.ID, &inv.Name, &inv.Type, &accountID, &units, &avgBuyPrice, &currentPrice,
			&inv.Currency, &inv.Notes, &lastUpdated, &inv.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning investment row: %w", err)
		}

		applyNullables(&inv, accountID, units, avgBuyPrice, currentPrice, lastUpdated)
		investments = append(investments, inv)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating investment rows: %w", err)
	}
	return investments, nil
}

// applyNullables maps sql.Null* values to the Investment pointer fields.
func applyNullables(inv *Investment, accountID sql.NullInt64, units, avgBuyPrice, currentPrice sql.NullFloat64, lastUpdated sql.NullString) {
	if accountID.Valid {
		inv.AccountID = &accountID.Int64
	}
	if units.Valid {
		inv.Units = &units.Float64
	}
	if avgBuyPrice.Valid {
		inv.AvgBuyPrice = &avgBuyPrice.Float64
	}
	if currentPrice.Valid {
		inv.CurrentPrice = &currentPrice.Float64
	}
	if lastUpdated.Valid {
		inv.LastUpdated = &lastUpdated.String
	}
}

// --- Nullable converters ---

func ptrToNullInt64(p *int64) interface{} {
	if p == nil {
		return nil
	}
	return *p
}

func ptrToNullFloat64(p *float64) interface{} {
	if p == nil {
		return nil
	}
	return *p
}

func ptrToNullString(p *string) interface{} {
	if p == nil {
		return nil
	}
	return *p
}
