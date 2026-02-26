package recurring

// Rule represents a recurring transaction rule.
type Rule struct {
	ID            int64   `json:"id"`
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"`
	AccountID     int64   `json:"account_id"`
	DestAccountID *int64  `json:"dest_account_id,omitempty"`
	Category      string  `json:"category"`
	Description   string  `json:"description"`
	Frequency     string  `json:"frequency"`
	DayOfMonth    *int    `json:"day_of_month,omitempty"`
	DayOfWeek     *int    `json:"day_of_week,omitempty"`
	MonthOfYear   *int    `json:"month_of_year,omitempty"`
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date,omitempty"`
	LastGenerated string  `json:"last_generated,omitempty"`
	IsActive      bool    `json:"is_active"`
	CreatedAt     string  `json:"created_at"`
}

// CreateRuleInput holds validated input for creating a recurring rule.
type CreateRuleInput struct {
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"`
	AccountID     *int64  `json:"account_id"`
	DestAccountID *int64  `json:"dest_account_id"`
	Category      string  `json:"category"`
	Description   string  `json:"description"`
	Frequency     string  `json:"frequency"`
	DayOfMonth    *int    `json:"day_of_month"`
	DayOfWeek     *int    `json:"day_of_week"`
	MonthOfYear   *int    `json:"month_of_year"`
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date"`
}

// UpdateRuleInput holds validated input for updating a recurring rule.
type UpdateRuleInput struct {
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"`
	AccountID     *int64  `json:"account_id"`
	DestAccountID *int64  `json:"dest_account_id"`
	Category      string  `json:"category"`
	Description   string  `json:"description"`
	Frequency     string  `json:"frequency"`
	DayOfMonth    *int    `json:"day_of_month"`
	DayOfWeek     *int    `json:"day_of_week"`
	MonthOfYear   *int    `json:"month_of_year"`
	StartDate     string  `json:"start_date"`
	EndDate       string  `json:"end_date"`
}

// GenerateResult holds the result of a generation run.
type GenerateResult struct {
	Generated int `json:"generated"`
}

// validFrequencies defines the allowed frequency values.
var validFrequencies = map[string]bool{
	"weekly":   true,
	"biweekly": true,
	"monthly":  true,
	"yearly":   true,
}

// IsValidFrequency checks whether a given frequency string is allowed.
func IsValidFrequency(freq string) bool {
	return validFrequencies[freq]
}

// validTypes defines the allowed transaction type values.
var validTypes = map[string]bool{
	"income":   true,
	"expense":  true,
	"transfer": true,
}

// IsValidType checks whether a given type string is allowed.
func IsValidType(txType string) bool {
	return validTypes[txType]
}
