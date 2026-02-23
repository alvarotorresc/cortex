package main

import (
	"database/sql"
	"embed"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/alvarotorresc/cortex/pkg/sdk"
	_ "modernc.org/sqlite"
)

//go:embed migrations/*.sql
var migrations embed.FS

// FinancePlugin implements sdk.CortexPlugin for personal finance tracking.
type FinancePlugin struct {
	db *sql.DB
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

// Migrate opens the SQLite database and runs embedded SQL migrations.
func (p *FinancePlugin) Migrate(databasePath string) error {
	database, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return fmt.Errorf("opening database: %w", err)
	}
	p.db = database

	// Enable WAL mode for better concurrent read performance.
	if _, err := p.db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		return fmt.Errorf("enabling WAL mode: %w", err)
	}

	migrationSQL, err := migrations.ReadFile("migrations/001_init.sql")
	if err != nil {
		return fmt.Errorf("reading migration: %w", err)
	}

	if _, err := p.db.Exec(string(migrationSQL)); err != nil {
		return fmt.Errorf("running migration: %w", err)
	}

	return nil
}

// HandleAPI routes incoming API requests to the appropriate handler.
func (p *FinancePlugin) HandleAPI(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	switch {
	case req.Method == "GET" && req.Path == "/transactions":
		return p.listTransactions(req)
	case req.Method == "POST" && req.Path == "/transactions":
		return p.createTransaction(req)
	case req.Method == "DELETE" && strings.HasPrefix(req.Path, "/transactions/"):
		return p.deleteTransaction(req)
	case req.Method == "GET" && req.Path == "/categories":
		return p.listCategories()
	case req.Method == "GET" && req.Path == "/summary":
		return p.getSummary(req)
	default:
		return jsonError(404, "NOT_FOUND", "route not found")
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

// --- Data types ---

// Transaction represents a financial transaction record.
type Transaction struct {
	ID          int64   `json:"id"`
	Amount      float64 `json:"amount"`
	Type        string  `json:"type"`
	Category    string  `json:"category"`
	Description string  `json:"description"`
	Date        string  `json:"date"`
	CreatedAt   string  `json:"created_at"`
}

// Category represents a transaction category.
type Category struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	Icon      string `json:"icon"`
	IsDefault bool   `json:"is_default"`
}

// CategorySummary represents aggregated spending per category.
type CategorySummary struct {
	Category string  `json:"category"`
	Total    float64 `json:"total"`
}

// --- Handlers ---

func (p *FinancePlugin) listTransactions(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	month := req.Query["month"]
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	rows, err := p.db.Query(
		`SELECT id, amount, type, category, description, date, created_at
		 FROM transactions
		 WHERE date LIKE ?
		 ORDER BY date DESC, id DESC`,
		month+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("querying transactions: %w", err)
	}
	defer rows.Close()

	transactions := make([]Transaction, 0)
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(&t.ID, &t.Amount, &t.Type, &t.Category, &t.Description, &t.Date, &t.CreatedAt); err != nil {
			return nil, fmt.Errorf("scanning transaction: %w", err)
		}
		transactions = append(transactions, t)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating transactions: %w", err)
	}

	return jsonSuccess(200, transactions)
}

func (p *FinancePlugin) createTransaction(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	var input struct {
		Amount      float64 `json:"amount"`
		Type        string  `json:"type"`
		Category    string  `json:"category"`
		Description string  `json:"description"`
		Date        string  `json:"date"`
	}

	if err := json.Unmarshal(req.Body, &input); err != nil {
		return jsonError(400, "VALIDATION_ERROR", "invalid JSON body")
	}

	// Validate amount
	if input.Amount <= 0 {
		return jsonError(400, "VALIDATION_ERROR", "amount must be greater than 0")
	}

	// Validate type enum
	if input.Type != "income" && input.Type != "expense" {
		return jsonError(400, "VALIDATION_ERROR", "type must be 'income' or 'expense'")
	}

	// Validate category is not empty
	if strings.TrimSpace(input.Category) == "" {
		return jsonError(400, "VALIDATION_ERROR", "category is required")
	}

	// Default date to today
	if input.Date == "" {
		input.Date = time.Now().Format("2006-01-02")
	}

	result, err := p.db.Exec(
		"INSERT INTO transactions (amount, type, category, description, date) VALUES (?, ?, ?, ?, ?)",
		input.Amount, input.Type, input.Category, input.Description, input.Date,
	)
	if err != nil {
		return nil, fmt.Errorf("inserting transaction: %w", err)
	}

	id, _ := result.LastInsertId()
	return jsonSuccess(201, map[string]interface{}{"id": id})
}

func (p *FinancePlugin) deleteTransaction(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	// Extract ID from path: /transactions/{id}
	parts := strings.Split(strings.TrimPrefix(req.Path, "/"), "/")
	if len(parts) < 2 || parts[1] == "" {
		return jsonError(400, "VALIDATION_ERROR", "missing transaction ID")
	}
	id := parts[1]

	result, err := p.db.Exec("DELETE FROM transactions WHERE id = ?", id)
	if err != nil {
		return nil, fmt.Errorf("deleting transaction: %w", err)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return jsonError(404, "NOT_FOUND", "transaction not found")
	}

	return jsonSuccess(200, map[string]interface{}{"deleted": id})
}

func (p *FinancePlugin) listCategories() (*sdk.APIResponse, error) {
	rows, err := p.db.Query("SELECT id, name, icon, is_default FROM categories ORDER BY is_default DESC, name")
	if err != nil {
		return nil, fmt.Errorf("querying categories: %w", err)
	}
	defer rows.Close()

	categories := make([]Category, 0)
	for rows.Next() {
		var c Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Icon, &c.IsDefault); err != nil {
			return nil, fmt.Errorf("scanning category: %w", err)
		}
		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating categories: %w", err)
	}

	return jsonSuccess(200, categories)
}

func (p *FinancePlugin) getSummary(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	month := req.Query["month"]
	if month == "" {
		month = time.Now().Format("2006-01")
	}

	rows, err := p.db.Query(
		`SELECT category, SUM(amount) as total
		 FROM transactions
		 WHERE type = 'expense' AND date LIKE ?
		 GROUP BY category
		 ORDER BY total DESC`,
		month+"%",
	)
	if err != nil {
		return nil, fmt.Errorf("querying summary: %w", err)
	}
	defer rows.Close()

	summaries := make([]CategorySummary, 0)
	for rows.Next() {
		var s CategorySummary
		if err := rows.Scan(&s.Category, &s.Total); err != nil {
			return nil, fmt.Errorf("scanning summary: %w", err)
		}
		summaries = append(summaries, s)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating summary: %w", err)
	}

	return jsonSuccess(200, summaries)
}

// --- JSON response helpers ---

// jsonSuccess wraps data in `{ "data": ... }` format per PATTERNS.md.
func jsonSuccess(status int, data interface{}) (*sdk.APIResponse, error) {
	body, err := json.Marshal(map[string]interface{}{"data": data})
	if err != nil {
		return nil, fmt.Errorf("marshaling response: %w", err)
	}
	return &sdk.APIResponse{
		StatusCode:  status,
		Body:        body,
		ContentType: "application/json",
	}, nil
}

// jsonError wraps errors in `{ "error": { "code": ..., "message": ... } }` format per PATTERNS.md.
func jsonError(status int, code string, message string) (*sdk.APIResponse, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    code,
			"message": message,
		},
	})
	return &sdk.APIResponse{
		StatusCode:  status,
		Body:        body,
		ContentType: "application/json",
	}, nil
}
