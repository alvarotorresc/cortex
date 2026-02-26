package budgets

import "regexp"

// Budget represents a spending budget for a category or globally.
type Budget struct {
	ID        int64   `json:"id"`
	Name      string  `json:"name"`
	Category  string  `json:"category"`
	Amount    float64 `json:"amount"`
	Month     string  `json:"month"`
	CreatedAt string  `json:"created_at"`
}

// BudgetWithProgress extends Budget with calculated spending metrics.
type BudgetWithProgress struct {
	Budget
	Spent      float64 `json:"spent"`
	Remaining  float64 `json:"remaining"`
	Percentage float64 `json:"percentage"`
}

// CreateBudgetInput holds validated input for creating a budget.
type CreateBudgetInput struct {
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Month    string  `json:"month"`
}

// UpdateBudgetInput holds validated input for updating a budget.
type UpdateBudgetInput struct {
	Name     string  `json:"name"`
	Category string  `json:"category"`
	Amount   float64 `json:"amount"`
	Month    string  `json:"month"`
}

// monthPattern validates YYYY-MM format.
var monthPattern = regexp.MustCompile(`^\d{4}-(0[1-9]|1[0-2])$`)

// IsValidMonth checks whether a string matches YYYY-MM format.
func IsValidMonth(month string) bool {
	return monthPattern.MatchString(month)
}
