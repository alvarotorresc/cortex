package reports

// CategoryTotal represents total spending for a single category.
type CategoryTotal struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

// AccountTotal represents the net transaction total for a single account.
type AccountTotal struct {
	AccountID   int64   `json:"account_id"`
	AccountName string  `json:"account_name"`
	Total       float64 `json:"total"`
}

// MonthlySummary aggregates income, expense, and balance for a single month
// with breakdowns by category and account.
type MonthlySummary struct {
	Month      string          `json:"month"`
	Income     float64         `json:"income"`
	Expense    float64         `json:"expense"`
	Balance    float64         `json:"balance"`
	ByCategory []CategoryTotal `json:"by_category"`
	ByAccount  []AccountTotal  `json:"by_account"`
}

// TrendPoint represents a single data point in a monthly trend series.
type TrendPoint struct {
	Month   string  `json:"month"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Balance float64 `json:"balance"`
}

// CategoryComparison compares a category's expense between the current
// and previous month, including the percentage change.
type CategoryComparison struct {
	Category      string  `json:"category"`
	CurrentMonth  float64 `json:"current_month"`
	PreviousMonth float64 `json:"previous_month"`
	Change        float64 `json:"change"`
}

// NetWorth represents the total net worth computed from account transactions
// and investment positions.
type NetWorth struct {
	AccountsTotal    float64 `json:"accounts_total"`
	InvestmentsTotal float64 `json:"investments_total"`
	NetWorth         float64 `json:"net_worth"`
}
