package reports

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
)

// Service executes read-only aggregation queries for financial reports.
// It takes *sql.DB directly (no repository layer) since these are complex
// aggregation queries, not standard CRUD operations.
type Service struct {
	db *sql.DB
}

// NewService creates a new reports Service.
func NewService(db *sql.DB) *Service {
	return &Service{db: db}
}

// Summary returns income, expense, balance, and breakdowns for a given month.
func (s *Service) Summary(month string) (*MonthlySummary, *shared.AppError) {
	prefix := month + "%"

	// Total income and expense for the month.
	var income, expense float64
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(CASE WHEN type='income' THEN amount ELSE 0 END), 0),
		        COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END), 0)
		 FROM transactions WHERE date LIKE ?`,
		prefix,
	).Scan(&income, &expense)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying monthly totals: %v", err), 500)
	}

	// By category (expenses only).
	catRows, err := s.db.Query(
		`SELECT category, SUM(amount) as total
		 FROM transactions
		 WHERE type = 'expense' AND date LIKE ?
		 GROUP BY category ORDER BY total DESC`,
		prefix,
	)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying categories: %v", err), 500)
	}
	defer catRows.Close()

	categories := make([]CategoryTotal, 0)
	for catRows.Next() {
		var ct CategoryTotal
		if err := catRows.Scan(&ct.Category, &ct.Total); err != nil {
			return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("scanning category: %v", err), 500)
		}
		categories = append(categories, ct)
	}
	if err := catRows.Err(); err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("iterating categories: %v", err), 500)
	}

	// By account (net balance per account for the month).
	acctRows, err := s.db.Query(
		`SELECT a.id, a.name,
		        COALESCE(SUM(CASE WHEN t.type='income' THEN t.amount
		                          WHEN t.type='expense' THEN -t.amount
		                          ELSE 0 END), 0) as total
		 FROM accounts a
		 LEFT JOIN transactions t ON t.account_id = a.id AND t.date LIKE ?
		 WHERE a.is_archived = 0
		 GROUP BY a.id, a.name`,
		prefix,
	)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying accounts: %v", err), 500)
	}
	defer acctRows.Close()

	accounts := make([]AccountTotal, 0)
	for acctRows.Next() {
		var at AccountTotal
		if err := acctRows.Scan(&at.AccountID, &at.AccountName, &at.Total); err != nil {
			return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("scanning account: %v", err), 500)
		}
		accounts = append(accounts, at)
	}
	if err := acctRows.Err(); err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("iterating accounts: %v", err), 500)
	}

	return &MonthlySummary{
		Month:      month,
		Income:     income,
		Expense:    expense,
		Balance:    income - expense,
		ByCategory: categories,
		ByAccount:  accounts,
	}, nil
}

// Trends returns monthly income/expense/balance totals between two months (inclusive).
func (s *Service) Trends(from, to string) ([]TrendPoint, *shared.AppError) {
	// Generate all months in range.
	months, appErr := generateMonths(from, to)
	if appErr != nil {
		return nil, appErr
	}

	// Query all transaction totals grouped by month within range.
	rows, err := s.db.Query(
		`SELECT substr(date, 1, 7) as month,
		        COALESCE(SUM(CASE WHEN type='income' THEN amount ELSE 0 END), 0) as income,
		        COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END), 0) as expense
		 FROM transactions
		 WHERE substr(date, 1, 7) >= ? AND substr(date, 1, 7) <= ?
		 GROUP BY substr(date, 1, 7)
		 ORDER BY month`,
		from, to,
	)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying trends: %v", err), 500)
	}
	defer rows.Close()

	// Build lookup of data we got from the DB.
	dataMap := make(map[string]TrendPoint)
	for rows.Next() {
		var tp TrendPoint
		if err := rows.Scan(&tp.Month, &tp.Income, &tp.Expense); err != nil {
			return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("scanning trend: %v", err), 500)
		}
		tp.Balance = tp.Income - tp.Expense
		dataMap[tp.Month] = tp
	}
	if err := rows.Err(); err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("iterating trends: %v", err), 500)
	}

	// Fill in all months, using zero values for months with no data.
	result := make([]TrendPoint, 0, len(months))
	for _, m := range months {
		if tp, ok := dataMap[m]; ok {
			result = append(result, tp)
		} else {
			result = append(result, TrendPoint{Month: m})
		}
	}

	return result, nil
}

// Categories returns expense totals by category for the given month compared
// to the previous month, including percentage change.
func (s *Service) Categories(month string) ([]CategoryComparison, *shared.AppError) {
	prevMonth, appErr := previousMonth(month)
	if appErr != nil {
		return nil, appErr
	}

	currentPrefix := month + "%"
	prevPrefix := prevMonth + "%"

	// Current month expenses by category.
	currentMap := make(map[string]float64)
	currentRows, err := s.db.Query(
		`SELECT category, SUM(amount) as total
		 FROM transactions
		 WHERE type = 'expense' AND date LIKE ?
		 GROUP BY category`,
		currentPrefix,
	)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying current categories: %v", err), 500)
	}
	defer currentRows.Close()

	var allCategories []string
	for currentRows.Next() {
		var cat string
		var total float64
		if err := currentRows.Scan(&cat, &total); err != nil {
			return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("scanning current category: %v", err), 500)
		}
		currentMap[cat] = total
		allCategories = append(allCategories, cat)
	}
	if err := currentRows.Err(); err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("iterating current categories: %v", err), 500)
	}

	// Previous month expenses by category.
	prevMap := make(map[string]float64)
	prevRows, err := s.db.Query(
		`SELECT category, SUM(amount) as total
		 FROM transactions
		 WHERE type = 'expense' AND date LIKE ?
		 GROUP BY category`,
		prevPrefix,
	)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying previous categories: %v", err), 500)
	}
	defer prevRows.Close()

	for prevRows.Next() {
		var cat string
		var total float64
		if err := prevRows.Scan(&cat, &total); err != nil {
			return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("scanning previous category: %v", err), 500)
		}
		prevMap[cat] = total
		// Add categories that only appear in previous month.
		if _, exists := currentMap[cat]; !exists {
			allCategories = append(allCategories, cat)
		}
	}
	if err := prevRows.Err(); err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("iterating previous categories: %v", err), 500)
	}

	// Build comparisons.
	result := make([]CategoryComparison, 0, len(allCategories))
	for _, cat := range allCategories {
		current := currentMap[cat]
		previous := prevMap[cat]

		var change float64
		if previous > 0 {
			change = ((current - previous) / previous) * 100
		}

		result = append(result, CategoryComparison{
			Category:      cat,
			CurrentMonth:  current,
			PreviousMonth: previous,
			Change:        change,
		})
	}

	return result, nil
}

// NetWorthReport returns total net worth computed from account transaction
// totals and investment positions.
func (s *Service) NetWorthReport() (*NetWorth, *shared.AppError) {
	// Sum of all income - expense across non-archived accounts.
	var accountsTotal float64
	err := s.db.QueryRow(
		`SELECT COALESCE(SUM(
		    CASE WHEN type='income' THEN amount
		         WHEN type='expense' THEN -amount
		         ELSE 0 END
		), 0) FROM transactions
		WHERE account_id IN (SELECT id FROM accounts WHERE is_archived = 0)`,
	).Scan(&accountsTotal)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying accounts total: %v", err), 500)
	}

	// Sum of units * current_price for all investments with both values set.
	var investmentsTotal float64
	err = s.db.QueryRow(
		`SELECT COALESCE(SUM(units * current_price), 0) FROM investments
		 WHERE units IS NOT NULL AND current_price IS NOT NULL`,
	).Scan(&investmentsTotal)
	if err != nil {
		return nil, shared.NewAppError("INTERNAL", fmt.Sprintf("querying investments total: %v", err), 500)
	}

	return &NetWorth{
		AccountsTotal:    accountsTotal,
		InvestmentsTotal: investmentsTotal,
		NetWorth:         accountsTotal + investmentsTotal,
	}, nil
}

// generateMonths returns a sorted slice of "YYYY-MM" strings from from to to (inclusive).
func generateMonths(from, to string) ([]string, *shared.AppError) {
	fromTime, err := time.Parse("2006-01", from)
	if err != nil {
		return nil, shared.NewValidationError(fmt.Sprintf("invalid 'from' month format: %s", from))
	}
	toTime, err := time.Parse("2006-01", to)
	if err != nil {
		return nil, shared.NewValidationError(fmt.Sprintf("invalid 'to' month format: %s", to))
	}

	if fromTime.After(toTime) {
		return nil, shared.NewValidationError("'from' must be before or equal to 'to'")
	}

	var months []string
	current := fromTime
	for !current.After(toTime) {
		months = append(months, current.Format("2006-01"))
		current = current.AddDate(0, 1, 0)
	}
	return months, nil
}

// previousMonth returns the month before the given "YYYY-MM" string.
func previousMonth(month string) (string, *shared.AppError) {
	t, err := time.Parse("2006-01", month)
	if err != nil {
		return "", shared.NewValidationError(fmt.Sprintf("invalid month format: %s", month))
	}
	prev := t.AddDate(0, -1, 0)
	return prev.Format("2006-01"), nil
}
