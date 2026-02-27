package transactions

// Transaction represents a financial transaction record with v2 fields.
type Transaction struct {
	ID                  int64   `json:"id"`
	Amount              float64 `json:"amount"`
	Type                string  `json:"type"`
	AccountID           int64   `json:"account_id"`
	DestAccountID       *int64  `json:"dest_account_id,omitempty"`
	Category            string  `json:"category"`
	Description         string  `json:"description"`
	Date                string  `json:"date"`
	IsRecurringInstance bool    `json:"is_recurring_instance"`
	RecurringRuleID     *int64  `json:"recurring_rule_id,omitempty"`
	Tags                []Tag   `json:"tags"`
	CreatedAt           string  `json:"created_at"`
}

// Tag is a lightweight tag representation embedded in transaction responses.
type Tag struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Color string `json:"color"`
}

// TransactionFilter holds combinable query parameters for listing transactions.
type TransactionFilter struct {
	Month    string
	Account  string
	Category string
	Tag      string
	Type     string
	Search   string
}

// CreateTransactionInput holds the validated input for creating a transaction.
type CreateTransactionInput struct {
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"`
	AccountID     *int64  `json:"account_id"`
	DestAccountID *int64  `json:"dest_account_id"`
	Category      string  `json:"category"`
	Description   string  `json:"description"`
	Date          string  `json:"date"`
	TagIDs        []int64 `json:"tag_ids"`
}

// UpdateTransactionInput holds the validated input for updating a transaction.
type UpdateTransactionInput struct {
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"`
	AccountID     *int64  `json:"account_id"`
	DestAccountID *int64  `json:"dest_account_id"`
	Category      string  `json:"category"`
	Description   string  `json:"description"`
	Date          string  `json:"date"`
	TagIDs        []int64 `json:"tag_ids"`
}

// validTransactionTypes defines the allowed transaction type values.
var validTransactionTypes = map[string]bool{
	"income":   true,
	"expense":  true,
	"transfer": true,
}

// IsValidTransactionType checks whether a given type string is allowed.
func IsValidTransactionType(txType string) bool {
	return validTransactionTypes[txType]
}
