package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/alvarotorresc/cortex/pkg/sdk"
)

// newTestPlugin creates a FinancePlugin with a migrated SQLite database in a temp directory.
// It returns the plugin ready for testing and calls t.Cleanup to close the database.
func newTestPlugin(t *testing.T) *FinancePlugin {
	t.Helper()

	p := &FinancePlugin{}
	dbPath := filepath.Join(t.TempDir(), "finance_test.db")

	if err := p.Migrate(dbPath); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	t.Cleanup(func() { p.Teardown() })
	return p
}

// parseDataArray parses an APIResponse body and returns the "data" field as raw JSON.
func parseDataArray(t *testing.T, resp *sdk.APIResponse) []json.RawMessage {
	t.Helper()

	var body struct {
		Data []json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	return body.Data
}

// parseErrorResponse parses an error APIResponse body and returns the code and message.
func parseErrorResponse(t *testing.T, resp *sdk.APIResponse) (code string, message string) {
	t.Helper()

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to parse error body: %v", err)
	}
	return body.Error.Code, body.Error.Message
}

// --- Tests ---

func TestMigrate_CreatesTables(t *testing.T) {
	p := &FinancePlugin{}
	dbPath := filepath.Join(t.TempDir(), "test_migrate.db")

	if err := p.Migrate(dbPath); err != nil {
		t.Fatalf("Migrate failed: %v", err)
	}
	defer p.Teardown()

	// Verify the transactions table exists by querying it.
	rows, err := p.db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='transactions'")
	if err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	defer rows.Close()

	if !rows.Next() {
		t.Fatal("transactions table does not exist after migration")
	}

	// Verify the categories table exists.
	rows2, err := p.db.Query("SELECT name FROM sqlite_master WHERE type='table' AND name='categories'")
	if err != nil {
		t.Fatalf("failed to query sqlite_master: %v", err)
	}
	defer rows2.Close()

	if !rows2.Next() {
		t.Fatal("categories table does not exist after migration")
	}
}

func TestCreateTransaction_Valid(t *testing.T) {
	p := newTestPlugin(t)

	createBody := `{"amount": 1500.50, "type": "income", "category": "salary", "description": "Monthly salary", "date": "2026-02-01"}`

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/transactions",
		Body:   []byte(createBody),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 201 {
		t.Fatalf("expected status 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify the transaction appears in the list.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"month": "2026-02"},
	})
	if err != nil {
		t.Fatalf("list transactions returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	if len(items) != 1 {
		t.Fatalf("expected 1 transaction, got %d", len(items))
	}

	var tx Transaction
	if err := json.Unmarshal(items[0], &tx); err != nil {
		t.Fatalf("failed to unmarshal transaction: %v", err)
	}
	if tx.Amount != 1500.50 {
		t.Errorf("expected amount 1500.50, got %f", tx.Amount)
	}
	if tx.Type != "income" {
		t.Errorf("expected type 'income', got '%s'", tx.Type)
	}
}

func TestCreateTransaction_NegativeAmount(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"amount": -100, "type": "expense", "category": "groceries", "date": "2026-02-01"}`

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/transactions",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected error code 'VALIDATION_ERROR', got '%s'", code)
	}
}

func TestCreateTransaction_InvalidType(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"amount": 100, "type": "other", "category": "groceries", "date": "2026-02-01"}`

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/transactions",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 400 {
		t.Fatalf("expected status 400, got %d", resp.StatusCode)
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected error code 'VALIDATION_ERROR', got '%s'", code)
	}
}

func TestListTransactions_FilterByMonth(t *testing.T) {
	p := newTestPlugin(t)

	// Create transactions in two different months.
	months := []struct {
		date  string
		month string
	}{
		{"2026-01-15", "2026-01"},
		{"2026-01-20", "2026-01"},
		{"2026-02-10", "2026-02"},
	}

	for _, m := range months {
		body := fmt.Sprintf(`{"amount": 50, "type": "expense", "category": "groceries", "date": "%s"}`, m.date)
		resp, err := p.HandleAPI(&sdk.APIRequest{
			Method: "POST",
			Path:   "/transactions",
			Body:   []byte(body),
		})
		if err != nil {
			t.Fatalf("failed to create transaction: %v", err)
		}
		if resp.StatusCode != 201 {
			t.Fatalf("expected 201, got %d", resp.StatusCode)
		}
	}

	// Filter by January: should get 2 transactions.
	janResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"month": "2026-01"},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	janItems := parseDataArray(t, janResp)
	if len(janItems) != 2 {
		t.Errorf("expected 2 transactions for January, got %d", len(janItems))
	}

	// Filter by February: should get 1 transaction.
	febResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"month": "2026-02"},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	febItems := parseDataArray(t, febResp)
	if len(febItems) != 1 {
		t.Errorf("expected 1 transaction for February, got %d", len(febItems))
	}
}

func TestDeleteTransaction(t *testing.T) {
	p := newTestPlugin(t)

	// Create a transaction.
	createBody := `{"amount": 200, "type": "expense", "category": "transport", "date": "2026-02-15"}`
	createResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/transactions",
		Body:   []byte(createBody),
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	var createData struct {
		Data struct {
			ID int64 `json:"id"`
		} `json:"data"`
	}
	if err := json.Unmarshal(createResp.Body, &createData); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}

	txID := fmt.Sprintf("%d", createData.Data.ID)

	// Delete the transaction.
	delResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   "/transactions/" + txID,
	})
	if err != nil {
		t.Fatalf("delete failed: %v", err)
	}

	if delResp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d. Body: %s", delResp.StatusCode, string(delResp.Body))
	}

	// Verify the transaction no longer appears in the list.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"month": "2026-02"},
	})
	if err != nil {
		t.Fatalf("list failed: %v", err)
	}

	items := parseDataArray(t, listResp)
	if len(items) != 0 {
		t.Errorf("expected 0 transactions after delete, got %d", len(items))
	}
}

func TestGetCategories(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/categories",
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d", resp.StatusCode)
	}

	var body struct {
		Data []Category `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to parse categories: %v", err)
	}

	// The migration seeds 8 default categories.
	if len(body.Data) != 8 {
		t.Fatalf("expected 8 default categories, got %d", len(body.Data))
	}

	// Verify all are marked as default.
	for _, c := range body.Data {
		if !c.IsDefault {
			t.Errorf("expected category '%s' to be default", c.Name)
		}
	}

	// Verify expected category names.
	expectedNames := map[string]bool{
		"salary": true, "groceries": true, "transport": true, "entertainment": true,
		"restaurants": true, "bills": true, "health": true, "other": true,
	}
	for _, c := range body.Data {
		if !expectedNames[c.Name] {
			t.Errorf("unexpected category name: '%s'", c.Name)
		}
	}
}

func TestWidgetData_MonthlyBalance(t *testing.T) {
	p := newTestPlugin(t)

	// Use the current month so GetWidgetData picks them up (it uses time.Now).
	currentMonth := time.Now().Format("2006-01")

	// Create income transaction.
	incomeBody := fmt.Sprintf(`{"amount": 3000, "type": "income", "category": "salary", "date": "%s-15"}`, currentMonth)
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/transactions",
		Body:   []byte(incomeBody),
	})
	if err != nil || resp.StatusCode != 201 {
		t.Fatalf("failed to create income: err=%v, status=%d", err, resp.StatusCode)
	}

	// Create expense transaction.
	expenseBody := fmt.Sprintf(`{"amount": 750.50, "type": "expense", "category": "groceries", "date": "%s-20"}`, currentMonth)
	resp, err = p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/transactions",
		Body:   []byte(expenseBody),
	})
	if err != nil || resp.StatusCode != 201 {
		t.Fatalf("failed to create expense: err=%v, status=%d", err, resp.StatusCode)
	}

	// Get widget data.
	widgetData, err := p.GetWidgetData("dashboard-widget")
	if err != nil {
		t.Fatalf("GetWidgetData returned error: %v", err)
	}

	var widget struct {
		Data struct {
			Income  float64 `json:"income"`
			Expense float64 `json:"expense"`
			Balance float64 `json:"balance"`
			Month   string  `json:"month"`
		} `json:"data"`
	}
	if err := json.Unmarshal(widgetData, &widget); err != nil {
		t.Fatalf("failed to parse widget data: %v", err)
	}

	if widget.Data.Income != 3000 {
		t.Errorf("expected income 3000, got %f", widget.Data.Income)
	}
	if widget.Data.Expense != 750.50 {
		t.Errorf("expected expense 750.50, got %f", widget.Data.Expense)
	}

	expectedBalance := 3000 - 750.50
	if widget.Data.Balance != expectedBalance {
		t.Errorf("expected balance %f, got %f", expectedBalance, widget.Data.Balance)
	}
	if widget.Data.Month != currentMonth {
		t.Errorf("expected month '%s', got '%s'", currentMonth, widget.Data.Month)
	}
}

// --- Migration v2 tests ---

func TestMigrate_V2TablesExist(t *testing.T) {
	p := newTestPlugin(t)

	expectedTables := []string{
		"transactions", "categories", "accounts", "tags",
		"transaction_tags", "recurring_rules", "budgets",
		"savings_goals", "investments",
	}
	for _, table := range expectedTables {
		var name string
		err := p.db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		).Scan(&name)
		if err != nil {
			t.Errorf("table %s does not exist after migration: %v", table, err)
		}
	}
}

func TestMigrate_DefaultAccountCreated(t *testing.T) {
	p := newTestPlugin(t)

	var name, accountType string
	err := p.db.QueryRow("SELECT name, type FROM accounts WHERE id = 1").Scan(&name, &accountType)
	if err != nil {
		t.Fatalf("default account not found: %v", err)
	}
	if name != "Main Account" || accountType != "checking" {
		t.Errorf("unexpected default account: name=%s type=%s", name, accountType)
	}
}

func TestMigrate_IdempotentExecution(t *testing.T) {
	p := &FinancePlugin{}
	dbPath := filepath.Join(t.TempDir(), "test_idempotent.db")

	if err := p.Migrate(dbPath); err != nil {
		t.Fatalf("first Migrate failed: %v", err)
	}
	p.Teardown()

	p2 := &FinancePlugin{}
	if err := p2.Migrate(dbPath); err != nil {
		t.Fatalf("second Migrate failed: %v", err)
	}
	defer p2.Teardown()

	var count int
	if err := p2.db.QueryRow("SELECT COUNT(*) FROM _migrations").Scan(&count); err != nil {
		t.Fatalf("query migrations count failed: %v", err)
	}
	if count != 2 {
		t.Errorf("expected 2 migrations recorded, got %d", count)
	}
}

func TestMigrate_MigrationTrackingTable(t *testing.T) {
	p := newTestPlugin(t)

	rows, err := p.db.Query("SELECT filename FROM _migrations ORDER BY filename")
	if err != nil {
		t.Fatalf("failed to query _migrations: %v", err)
	}
	defer rows.Close()

	var filenames []string
	for rows.Next() {
		var f string
		if err := rows.Scan(&f); err != nil {
			t.Fatalf("failed to scan filename: %v", err)
		}
		filenames = append(filenames, f)
	}

	if len(filenames) != 2 {
		t.Fatalf("expected 2 migration records, got %d: %v", len(filenames), filenames)
	}
	if filenames[0] != "001_init.sql" || filenames[1] != "002_enhanced.sql" {
		t.Errorf("unexpected migration filenames: %v", filenames)
	}
}

func TestMigrate_TransactionsHaveAccountIDDefault(t *testing.T) {
	p := newTestPlugin(t)

	// Create a transaction with the old format (no account_id specified).
	_, err := p.db.Exec(
		"INSERT INTO transactions (amount, type, category, description, date) VALUES (?, ?, ?, ?, ?)",
		100.0, "expense", "groceries", "test", "2026-02-01",
	)
	if err != nil {
		t.Fatalf("insert without account_id failed: %v", err)
	}

	// Verify the default account_id is 1.
	var accountID int
	if err := p.db.QueryRow("SELECT account_id FROM transactions WHERE id = 1").Scan(&accountID); err != nil {
		t.Fatalf("failed to read account_id: %v", err)
	}
	if accountID != 1 {
		t.Errorf("expected default account_id=1, got %d", accountID)
	}
}

func TestMigrate_V2Indexes(t *testing.T) {
	p := newTestPlugin(t)

	expectedIndexes := []string{
		"idx_transactions_account",
		"idx_transactions_category",
		"idx_recurring_rules_active",
		"idx_budgets_month",
		"idx_investments_type",
	}
	for _, idx := range expectedIndexes {
		var name string
		err := p.db.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='index' AND name=?", idx,
		).Scan(&name)
		if err != nil {
			t.Errorf("index %s does not exist after migration: %v", idx, err)
		}
	}
}
