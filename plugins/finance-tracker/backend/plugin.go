package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/alvarotorresc/cortex/pkg/sdk"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/accounts"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/budgets"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/categories"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/goals"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/investments"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/recurring"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/reports"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/shared"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/tags"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/transactions"
)

//go:embed migrations/*.sql
var migrations embed.FS

// FinancePlugin implements sdk.CortexPlugin for personal finance tracking.
type FinancePlugin struct {
	db                  *sql.DB
	accountsHandler     *accounts.Handler
	budgetsHandler      *budgets.Handler
	categoriesHandler   *categories.Handler
	goalsHandler        *goals.Handler
	investmentsHandler  *investments.Handler
	tagsHandler         *tags.Handler
	transactionsHandler *transactions.Handler
	recurringHandler    *recurring.Handler
	reportsHandler      *reports.Handler
}

// GetManifest returns the plugin's metadata.
func (p *FinancePlugin) GetManifest() (*sdk.Manifest, error) {
	return &sdk.Manifest{
		ID:          "finance-tracker",
		Name:        "Finance Tracker",
		Version:     "0.1.0",
		Description: "Track income and expenses, local and private",
		Icon:        "wallet",
		Color:       "#10B981",
		Permissions: []string{"db:read", "db:write"},
	}, nil
}

// Migrate opens the SQLite database and runs all embedded SQL migrations in
// order, tracking which ones have been applied to ensure idempotency.
func (p *FinancePlugin) Migrate(databasePath string) error {
	db, err := shared.OpenDatabase(databasePath)
	if err != nil {
		return err
	}
	p.db = db

	// Create migrations tracking table.
	if _, err := p.db.Exec(`
		CREATE TABLE IF NOT EXISTS _migrations (
			filename TEXT PRIMARY KEY,
			applied_at TEXT NOT NULL DEFAULT (datetime('now'))
		)
	`); err != nil {
		return fmt.Errorf("creating migrations table: %w", err)
	}

	entries, err := migrations.ReadDir("migrations")
	if err != nil {
		return fmt.Errorf("reading migrations dir: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Skip if already applied.
		var count int
		if err := p.db.QueryRow(
			"SELECT COUNT(*) FROM _migrations WHERE filename = ?", entry.Name(),
		).Scan(&count); err != nil {
			return fmt.Errorf("checking migration %s: %w", entry.Name(), err)
		}
		if count > 0 {
			continue
		}

		migrationSQL, err := migrations.ReadFile("migrations/" + entry.Name())
		if err != nil {
			return fmt.Errorf("reading migration %s: %w", entry.Name(), err)
		}

		if _, err := p.db.Exec(string(migrationSQL)); err != nil {
			return fmt.Errorf("running migration %s: %w", entry.Name(), err)
		}

		if _, err := p.db.Exec(
			"INSERT INTO _migrations (filename) VALUES (?)", entry.Name(),
		); err != nil {
			return fmt.Errorf("recording migration %s: %w", entry.Name(), err)
		}
	}

	p.accountsHandler = accounts.NewHandler(p.db)
	p.budgetsHandler = budgets.NewHandler(p.db)
	p.categoriesHandler = categories.NewHandler(p.db)
	p.goalsHandler = goals.NewHandler(p.db)
	p.investmentsHandler = investments.NewHandler(p.db)
	p.tagsHandler = tags.NewHandler(p.db)
	p.transactionsHandler = transactions.NewHandler(p.db)
	p.recurringHandler = recurring.NewHandler(p.db)
	p.reportsHandler = reports.NewHandler(p.db)

	return nil
}

// HandleAPI routes incoming API requests to the appropriate handler.
func (p *FinancePlugin) HandleAPI(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	switch {
	case strings.HasPrefix(req.Path, "/transactions"):
		return p.transactionsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/categories"):
		return p.categoriesHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/reports"):
		return p.reportsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/accounts"):
		return p.accountsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/budgets"):
		return p.budgetsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/goals"):
		return p.goalsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/investments"):
		return p.investmentsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/tags"):
		return p.tagsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/recurring"):
		return p.recurringHandler.Handle(req)
	default:
		return shared.JSONError(shared.NewAppError("NOT_FOUND", "route not found", 404))
	}
}

// widgetSparklineEntry represents a single month in the sparkline trend data.
type widgetSparklineEntry struct {
	Month   string  `json:"month"`
	Balance float64 `json:"balance"`
}

// widgetBudgetProgress represents the global budget progress for the current month.
type widgetBudgetProgress struct {
	Amount     float64 `json:"amount"`
	Spent      float64 `json:"spent"`
	Remaining  float64 `json:"remaining"`
	Percentage float64 `json:"percentage"`
}

// widgetData represents the full dashboard widget response payload.
type widgetData struct {
	Income    float64                `json:"income"`
	Expense   float64                `json:"expense"`
	Balance   float64                `json:"balance"`
	Month     string                 `json:"month"`
	Sparkline []widgetSparklineEntry `json:"sparkline"`
	Budget    *widgetBudgetProgress  `json:"budget"`
}

// GetWidgetData returns dashboard widget data for the requested slot.
func (p *FinancePlugin) GetWidgetData(slot string) ([]byte, error) {
	if slot != "dashboard-widget" {
		return json.Marshal(map[string]interface{}{"data": nil})
	}

	now := time.Now()
	month := now.Format("2006-01")

	// Current month income/expense.
	var income, expense float64
	row := p.db.QueryRow(
		`SELECT COALESCE(SUM(CASE WHEN type='income' THEN amount ELSE 0 END), 0),
		        COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END), 0)
		 FROM transactions WHERE date LIKE ?`,
		month+"%",
	)
	if err := row.Scan(&income, &expense); err != nil {
		return nil, fmt.Errorf("querying monthly totals: %w", err)
	}

	// Sparkline: last 6 months balance trend.
	sparkline, err := p.getSparkline(now)
	if err != nil {
		return nil, fmt.Errorf("querying sparkline: %w", err)
	}

	// Budget progress: global budget for current month.
	budget, err := p.getGlobalBudgetProgress(month)
	if err != nil {
		return nil, fmt.Errorf("querying budget progress: %w", err)
	}

	data := widgetData{
		Income:    income,
		Expense:   expense,
		Balance:   income - expense,
		Month:     month,
		Sparkline: sparkline,
		Budget:    budget,
	}

	return json.Marshal(map[string]interface{}{"data": data})
}

// getSparkline returns the balance for each of the last 6 months (including current).
// Months with no transactions are filled with zero balance.
func (p *FinancePlugin) getSparkline(now time.Time) ([]widgetSparklineEntry, error) {
	startMonth := now.AddDate(0, -5, 0).Format("2006-01")
	endMonth := now.Format("2006-01")

	rows, err := p.db.Query(
		`SELECT substr(date, 1, 7) as month,
		        COALESCE(SUM(CASE WHEN type='income' THEN amount ELSE 0 END), 0) -
		        COALESCE(SUM(CASE WHEN type='expense' THEN amount ELSE 0 END), 0) as balance
		 FROM transactions
		 WHERE substr(date, 1, 7) >= ? AND substr(date, 1, 7) <= ?
		 GROUP BY substr(date, 1, 7)
		 ORDER BY month`,
		startMonth, endMonth,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	// Collect query results into a map for gap-filling.
	balanceByMonth := make(map[string]float64)
	for rows.Next() {
		var m string
		var b float64
		if err := rows.Scan(&m, &b); err != nil {
			return nil, err
		}
		balanceByMonth[m] = b
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Build 6-month array, filling gaps with zero.
	sparkline := make([]widgetSparklineEntry, 6)
	for i := 0; i < 6; i++ {
		m := now.AddDate(0, -(5 - i), 0).Format("2006-01")
		sparkline[i] = widgetSparklineEntry{
			Month:   m,
			Balance: balanceByMonth[m], // defaults to 0 if not in map
		}
	}

	return sparkline, nil
}

// getGlobalBudgetProgress returns the budget progress for the global budget
// (category IS NULL or empty) that matches the current month or is a recurring
// budget (month IS NULL or empty). Returns nil if no global budget exists.
func (p *FinancePlugin) getGlobalBudgetProgress(month string) (*widgetBudgetProgress, error) {
	var budgetAmount float64
	err := p.db.QueryRow(
		`SELECT amount FROM budgets
		 WHERE (category IS NULL OR category = '')
		   AND (month = ? OR month = '' OR month IS NULL)
		 ORDER BY CASE WHEN month = ? THEN 0 ELSE 1 END
		 LIMIT 1`,
		month, month,
	).Scan(&budgetAmount)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No global budget exists.
		}
		return nil, err
	}

	// Sum expenses for the current month.
	var spent float64
	if err := p.db.QueryRow(
		`SELECT COALESCE(SUM(amount), 0) FROM transactions
		 WHERE type = 'expense' AND date LIKE ?`,
		month+"%",
	).Scan(&spent); err != nil {
		return nil, err
	}

	remaining := budgetAmount - spent
	var percentage float64
	if budgetAmount > 0 {
		percentage = (spent / budgetAmount) * 100
	}

	return &widgetBudgetProgress{
		Amount:     budgetAmount,
		Spent:      spent,
		Remaining:  remaining,
		Percentage: percentage,
	}, nil
}

// Teardown closes the database connection when the plugin is unloaded.
func (p *FinancePlugin) Teardown() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
