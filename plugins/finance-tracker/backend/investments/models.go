package investments

// Investment represents an investment position (crypto, ETF, fund, stock, etc.).
type Investment struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	AccountID    *int64   `json:"account_id,omitempty"`
	Units        *float64 `json:"units,omitempty"`
	AvgBuyPrice  *float64 `json:"avg_buy_price,omitempty"`
	CurrentPrice *float64 `json:"current_price,omitempty"`
	Currency     string   `json:"currency"`
	Notes        string   `json:"notes"`
	LastUpdated  *string  `json:"last_updated,omitempty"`
	CreatedAt    string   `json:"created_at"`
}

// InvestmentWithPnL extends Investment with computed profit & loss fields.
type InvestmentWithPnL struct {
	Investment
	TotalInvested *float64 `json:"total_invested,omitempty"`
	CurrentValue  *float64 `json:"current_value,omitempty"`
	PnL           *float64 `json:"pnl,omitempty"`
	PnLPercentage *float64 `json:"pnl_percentage,omitempty"`
}

// CreateInvestmentInput holds validated input for creating an investment.
type CreateInvestmentInput struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	AccountID    *int64   `json:"account_id"`
	Units        *float64 `json:"units"`
	AvgBuyPrice  *float64 `json:"avg_buy_price"`
	CurrentPrice *float64 `json:"current_price"`
	Currency     string   `json:"currency"`
	Notes        string   `json:"notes"`
	LastUpdated  *string  `json:"last_updated"`
}

// UpdateInvestmentInput holds validated input for updating an investment.
type UpdateInvestmentInput struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	AccountID    *int64   `json:"account_id"`
	Units        *float64 `json:"units"`
	AvgBuyPrice  *float64 `json:"avg_buy_price"`
	CurrentPrice *float64 `json:"current_price"`
	Currency     string   `json:"currency"`
	Notes        string   `json:"notes"`
	LastUpdated  *string  `json:"last_updated"`
}

var validInvestmentTypes = map[string]bool{
	"crypto": true, "etf": true, "fund": true, "stock": true, "other": true,
}
