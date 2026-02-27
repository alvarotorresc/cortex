package accounts

// Account represents a financial account record.
type Account struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Currency     string   `json:"currency"`
	InterestRate *float64 `json:"interest_rate,omitempty"`
	Icon         string   `json:"icon"`
	Color        string   `json:"color"`
	IsArchived   bool     `json:"is_archived"`
	CreatedAt    string   `json:"created_at"`
}

// AccountWithBalance extends Account with a computed balance from transactions.
type AccountWithBalance struct {
	Account
	Balance           float64  `json:"balance"`
	EstimatedInterest *float64 `json:"estimated_interest,omitempty"`
}

// CreateAccountInput holds the validated input for creating an account.
type CreateAccountInput struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Currency     string   `json:"currency"`
	InterestRate *float64 `json:"interest_rate"`
	Icon         string   `json:"icon"`
	Color        string   `json:"color"`
}

// UpdateAccountInput holds the validated input for updating an account.
type UpdateAccountInput struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Currency     string   `json:"currency"`
	InterestRate *float64 `json:"interest_rate"`
	Icon         string   `json:"icon"`
	Color        string   `json:"color"`
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
