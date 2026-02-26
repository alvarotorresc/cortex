package main

import (
	"database/sql"
	"embed"
	"encoding/json"
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

// GetWidgetData returns dashboard widget data for the requested slot.
func (p *FinancePlugin) GetWidgetData(slot string) ([]byte, error) {
	if slot != "dashboard-widget" {
		return json.Marshal(map[string]interface{}{"data": nil})
	}

	month := time.Now().Format("2006-01")

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

	return json.Marshal(map[string]interface{}{
		"data": map[string]interface{}{
			"income":  income,
			"expense": expense,
			"balance": income - expense,
			"month":   month,
		},
	})
}

// Teardown closes the database connection when the plugin is unloaded.
func (p *FinancePlugin) Teardown() error {
	if p.db != nil {
		return p.db.Close()
	}
	return nil
}
