package transactions

import (
	"database/sql"
	"fmt"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Repository handles database operations for transactions.
type Repository struct {
	db *sql.DB
}

// NewRepository creates a Repository backed by the given database connection.
func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

// List returns transactions matching the given filter, ordered by date descending.
// Builds WHERE clauses dynamically based on non-empty filter fields.
func (r *Repository) List(filter *TransactionFilter) ([]Transaction, error) {
	query := `SELECT id, amount, type, account_id, dest_account_id, category, description, date,
	          is_recurring_instance, recurring_rule_id, created_at
	          FROM transactions WHERE 1=1`
	args := []interface{}{}

	if filter.Month != "" {
		query += " AND date LIKE ?"
		args = append(args, filter.Month+"%")
	}
	if filter.Account != "" {
		query += " AND account_id = ?"
		args = append(args, filter.Account)
	}
	if filter.Category != "" {
		query += " AND category = ?"
		args = append(args, filter.Category)
	}
	if filter.Type != "" {
		query += " AND type = ?"
		args = append(args, filter.Type)
	}
	if filter.Search != "" {
		query += " AND (description LIKE ? OR category LIKE ?)"
		args = append(args, "%"+filter.Search+"%", "%"+filter.Search+"%")
	}
	if filter.Tag != "" {
		query += " AND id IN (SELECT transaction_id FROM transaction_tags WHERE tag_id = ?)"
		args = append(args, filter.Tag)
	}

	query += " ORDER BY date DESC, id DESC"

	rows, err := r.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("querying transactions: %w", err)
	}
	defer rows.Close()

	transactions, err := scanTransactions(rows)
	if err != nil {
		return nil, err
	}

	// Load tags for each transaction.
	for i := range transactions {
		tags, err := r.loadTags(transactions[i].ID)
		if err != nil {
			return nil, err
		}
		transactions[i].Tags = tags
	}

	return transactions, nil
}

// GetByID returns a single transaction by its ID.
func (r *Repository) GetByID(id int64) (*Transaction, *shared.AppError) {
	var tx Transaction
	var destAccountID sql.NullInt64
	var recurringRuleID sql.NullInt64
	var isRecurring int

	err := r.db.QueryRow(
		`SELECT id, amount, type, account_id, dest_account_id, category, description, date,
		 is_recurring_instance, recurring_rule_id, created_at
		 FROM transactions WHERE id = ?`, id,
	).Scan(
		&tx.ID, &tx.Amount, &tx.Type, &tx.AccountID, &destAccountID,
		&tx.Category, &tx.Description, &tx.Date,
		&isRecurring, &recurringRuleID, &tx.CreatedAt,
	)
	if err == sql.ErrNoRows {
		return nil, shared.NewNotFoundError("transaction", fmt.Sprintf("%d", id))
	}
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying transaction: %v", err), 500)
	}

	tx.IsRecurringInstance = isRecurring == 1
	if destAccountID.Valid {
		tx.DestAccountID = &destAccountID.Int64
	}
	if recurringRuleID.Valid {
		tx.RecurringRuleID = &recurringRuleID.Int64
	}

	tags, err2 := r.loadTags(tx.ID)
	if err2 != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("loading tags: %v", err2), 500)
	}
	tx.Tags = tags

	return &tx, nil
}

// Create inserts a new transaction and links tags atomically within a DB transaction.
func (r *Repository) Create(input *CreateTransactionInput) (int64, error) {
	tx, err := r.db.Begin()
	if err != nil {
		return 0, fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var destAccountID interface{}
	if input.DestAccountID != nil {
		destAccountID = *input.DestAccountID
	}

	result, err := tx.Exec(
		`INSERT INTO transactions (amount, type, account_id, dest_account_id, category, description, date)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		input.Amount, input.Type, *input.AccountID, destAccountID,
		input.Category, input.Description, input.Date,
	)
	if err != nil {
		return 0, fmt.Errorf("inserting transaction: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("getting last insert id: %w", err)
	}

	for _, tagID := range input.TagIDs {
		if _, err := tx.Exec(
			"INSERT INTO transaction_tags (transaction_id, tag_id) VALUES (?, ?)",
			id, tagID,
		); err != nil {
			return 0, fmt.Errorf("inserting transaction tag: %w", err)
		}
	}

	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("committing transaction: %w", err)
	}
	return id, nil
}

// Update modifies an existing transaction's fields and replaces tag links atomically.
func (r *Repository) Update(id int64, input *UpdateTransactionInput) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("beginning transaction: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var destAccountID interface{}
	if input.DestAccountID != nil {
		destAccountID = *input.DestAccountID
	}

	var accountID int64 = 1
	if input.AccountID != nil {
		accountID = *input.AccountID
	}

	result, err := tx.Exec(
		`UPDATE transactions SET amount = ?, type = ?, account_id = ?, dest_account_id = ?,
		 category = ?, description = ?, date = ?
		 WHERE id = ?`,
		input.Amount, input.Type, accountID, destAccountID,
		input.Category, input.Description, input.Date, id,
	)
	if err != nil {
		return fmt.Errorf("updating transaction: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("transaction", fmt.Sprintf("%d", id))
	}

	// Replace tag links within the same transaction.
	if _, err := tx.Exec("DELETE FROM transaction_tags WHERE transaction_id = ?", id); err != nil {
		return fmt.Errorf("clearing transaction tags: %w", err)
	}
	for _, tagID := range input.TagIDs {
		if _, err := tx.Exec(
			"INSERT INTO transaction_tags (transaction_id, tag_id) VALUES (?, ?)",
			id, tagID,
		); err != nil {
			return fmt.Errorf("inserting transaction tag: %w", err)
		}
	}

	return tx.Commit()
}

// Delete removes a transaction by its ID.
func (r *Repository) Delete(id int64) error {
	result, err := r.db.Exec("DELETE FROM transactions WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("deleting transaction: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return shared.NewNotFoundError("transaction", fmt.Sprintf("%d", id))
	}
	return nil
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

// SetTags replaces all tag associations for a transaction.
// Deletes existing links and inserts the new ones.
func (r *Repository) SetTags(transactionID int64, tagIDs []int64) error {
	if _, err := r.db.Exec("DELETE FROM transaction_tags WHERE transaction_id = ?", transactionID); err != nil {
		return fmt.Errorf("clearing transaction tags: %w", err)
	}

	for _, tagID := range tagIDs {
		if _, err := r.db.Exec(
			"INSERT INTO transaction_tags (transaction_id, tag_id) VALUES (?, ?)",
			transactionID, tagID,
		); err != nil {
			return fmt.Errorf("inserting transaction tag: %w", err)
		}
	}
	return nil
}

// loadTags returns all tags associated with a transaction via the join table.
func (r *Repository) loadTags(transactionID int64) ([]Tag, error) {
	rows, err := r.db.Query(
		`SELECT t.id, t.name, COALESCE(t.color, '') FROM tags t
		 INNER JOIN transaction_tags tt ON t.id = tt.tag_id
		 WHERE tt.transaction_id = ?`, transactionID,
	)
	if err != nil {
		return nil, fmt.Errorf("loading tags for transaction %d: %w", transactionID, err)
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

// scanTransactions reads all rows from the result set into a slice of Transaction.
func scanTransactions(rows *sql.Rows) ([]Transaction, error) {
	transactions := make([]Transaction, 0)
	for rows.Next() {
		var tx Transaction
		var destAccountID sql.NullInt64
		var recurringRuleID sql.NullInt64
		var isRecurring int

		if err := rows.Scan(
			&tx.ID, &tx.Amount, &tx.Type, &tx.AccountID, &destAccountID,
			&tx.Category, &tx.Description, &tx.Date,
			&isRecurring, &recurringRuleID, &tx.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scanning transaction row: %w", err)
		}

		tx.IsRecurringInstance = isRecurring == 1
		if destAccountID.Valid {
			tx.DestAccountID = &destAccountID.Int64
		}
		if recurringRuleID.Valid {
			tx.RecurringRuleID = &recurringRuleID.Int64
		}
		tx.Tags = make([]Tag, 0)
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating transaction rows: %w", err)
	}
	return transactions, nil
}
