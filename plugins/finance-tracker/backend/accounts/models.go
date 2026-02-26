package accounts

// Account represents a financial account record.
type Account struct {
	ID             int64   `json:"id"`
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Currency       string  `json:"currency"`
	InitialBalance float64 `json:"initial_balance"`
	IsActive       bool    `json:"is_active"`
	SortOrder      int     `json:"sort_order"`
	CreatedAt      string  `json:"created_at"`
}

// AccountWithBalance extends Account with a computed balance from transactions.
type AccountWithBalance struct {
	Account
	Balance float64 `json:"balance"`
}

// CreateAccountInput holds the validated input for creating an account.
type CreateAccountInput struct {
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Currency       string  `json:"currency"`
	InitialBalance float64 `json:"initial_balance"`
	SortOrder      int     `json:"sort_order"`
}

// UpdateAccountInput holds the validated input for updating an account.
type UpdateAccountInput struct {
	Name           string  `json:"name"`
	Type           string  `json:"type"`
	Currency       string  `json:"currency"`
	InitialBalance float64 `json:"initial_balance"`
	SortOrder      int     `json:"sort_order"`
}

// validAccountTypes defines the allowed account type values.
var validAccountTypes = map[string]bool{
	"checking":   true,
	"savings":    true,
	"cash":       true,
	"investment": true,
}

// IsValidAccountType checks whether a given type string is allowed.
func IsValidAccountType(accountType string) bool {
	return validAccountTypes[accountType]
}
