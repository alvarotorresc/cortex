package main

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/alvarotorresc/cortex/pkg/sdk"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/recurring"
	"github.com/alvarotorresc/cortex/plugins/finance-tracker/backend/transactions"
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

// parseDataObject parses an APIResponse body and returns the "data" field as raw JSON.
func parseDataObject(t *testing.T, resp *sdk.APIResponse) json.RawMessage {
	t.Helper()

	var body struct {
		Data json.RawMessage `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to parse response body: %v", err)
	}
	return body.Data
}

// createTransaction is a test helper that creates a transaction via the API and returns the ID.
func createTransaction(t *testing.T, p *FinancePlugin, body string) int64 {
	t.Helper()

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/transactions",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("create transaction failed: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	data := parseDataObject(t, resp)
	var tx struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &tx); err != nil {
		t.Fatalf("failed to parse created transaction: %v", err)
	}
	return tx.ID
}

// createTag is a test helper that creates a tag via the API and returns the ID.
func createTag(t *testing.T, p *FinancePlugin, name string, color string) int64 {
	t.Helper()

	body := fmt.Sprintf(`{"name":"%s","color":"%s"}`, name, color)
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/tags",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("create tag failed: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	data := parseDataObject(t, resp)
	var tag struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &tag); err != nil {
		t.Fatalf("failed to parse created tag: %v", err)
	}
	return tag.ID
}

// createAccount is a test helper that creates an account via the API and returns the ID.
func createAccount(t *testing.T, p *FinancePlugin, body string) int64 {
	t.Helper()

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/accounts",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("create account failed: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	data := parseDataObject(t, resp)
	var acct struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &acct); err != nil {
		t.Fatalf("failed to parse created account: %v", err)
	}
	return acct.ID
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

	var tx transactions.Transaction
	if err := json.Unmarshal(items[0], &tx); err != nil {
		t.Fatalf("failed to unmarshal transaction: %v", err)
	}
	if tx.Amount != 1500.50 {
		t.Errorf("expected amount 1500.50, got %f", tx.Amount)
	}
	if tx.Type != "income" {
		t.Errorf("expected type 'income', got '%s'", tx.Type)
	}
	if tx.AccountID != 1 {
		t.Errorf("expected default account_id 1, got %d", tx.AccountID)
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
	txID := createTransaction(t, p, `{"amount": 200, "type": "expense", "category": "transport", "date": "2026-02-15"}`)

	// Delete the transaction.
	delResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   fmt.Sprintf("/transactions/%d", txID),
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

// --- Transaction v2 tests ---

func TestCreateTransaction_WithAccount(t *testing.T) {
	p := newTestPlugin(t)

	// Create a second account.
	acctID := createAccount(t, p, `{"name":"Savings","type":"savings","currency":"EUR"}`)

	// Create a transaction linked to that account.
	body := fmt.Sprintf(`{"amount":500,"type":"income","category":"salary","date":"2026-02-01","account_id":%d}`, acctID)
	txID := createTransaction(t, p, body)

	// Verify the transaction has the correct account_id.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"month": "2026-02"},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	var found bool
	for _, raw := range items {
		var tx transactions.Transaction
		if err := json.Unmarshal(raw, &tx); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if tx.ID == txID {
			found = true
			if tx.AccountID != acctID {
				t.Errorf("expected account_id %d, got %d", acctID, tx.AccountID)
			}
		}
	}
	if !found {
		t.Error("created transaction not found in list")
	}
}

func TestCreateTransaction_DefaultAccount(t *testing.T) {
	p := newTestPlugin(t)

	// Create without account_id — should default to 1.
	body := `{"amount":100,"type":"expense","category":"groceries","date":"2026-02-01"}`
	txID := createTransaction(t, p, body)

	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"month": "2026-02"},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	for _, raw := range items {
		var tx transactions.Transaction
		if err := json.Unmarshal(raw, &tx); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if tx.ID == txID && tx.AccountID != 1 {
			t.Errorf("expected default account_id 1, got %d", tx.AccountID)
		}
	}
}

func TestCreateTransaction_Transfer(t *testing.T) {
	p := newTestPlugin(t)

	// Create a second account for the transfer destination.
	destID := createAccount(t, p, `{"name":"Savings","type":"savings","currency":"EUR"}`)

	body := fmt.Sprintf(
		`{"amount":1000,"type":"transfer","category":"","description":"Monthly savings","date":"2026-02-01","account_id":1,"dest_account_id":%d}`,
		destID,
	)
	txID := createTransaction(t, p, body)

	// Verify the transaction.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"month": "2026-02"},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	for _, raw := range items {
		var tx transactions.Transaction
		if err := json.Unmarshal(raw, &tx); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if tx.ID == txID {
			if tx.Type != "transfer" {
				t.Errorf("expected type 'transfer', got '%s'", tx.Type)
			}
			if tx.DestAccountID == nil || *tx.DestAccountID != destID {
				t.Errorf("expected dest_account_id %d, got %v", destID, tx.DestAccountID)
			}
		}
	}
}

func TestCreateTransaction_TransferMissingDest(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"amount":500,"type":"transfer","category":"","date":"2026-02-01","account_id":1}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/transactions",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestUpdateTransaction(t *testing.T) {
	p := newTestPlugin(t)

	txID := createTransaction(t, p, `{"amount":100,"type":"expense","category":"groceries","description":"Weekly shop","date":"2026-02-01"}`)

	// Update amount, category, and description.
	updateBody := `{"amount":150,"type":"expense","category":"transport","description":"Taxi ride","date":"2026-02-01"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "PUT",
		Path:   fmt.Sprintf("/transactions/%d", txID),
		Body:   []byte(updateBody),
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify the update.
	data := parseDataObject(t, resp)
	var tx transactions.Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		t.Fatalf("failed to unmarshal updated transaction: %v", err)
	}
	if tx.Amount != 150 {
		t.Errorf("expected amount 150, got %f", tx.Amount)
	}
	if tx.Category != "transport" {
		t.Errorf("expected category 'transport', got '%s'", tx.Category)
	}
	if tx.Description != "Taxi ride" {
		t.Errorf("expected description 'Taxi ride', got '%s'", tx.Description)
	}
}

func TestDeleteTransaction_V2(t *testing.T) {
	p := newTestPlugin(t)

	txID := createTransaction(t, p, `{"amount":75,"type":"expense","category":"entertainment","date":"2026-02-10"}`)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   fmt.Sprintf("/transactions/%d", txID),
	})
	if err != nil {
		t.Fatalf("delete returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify deletion.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"month": "2026-02"},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	if len(items) != 0 {
		t.Errorf("expected 0 transactions, got %d", len(items))
	}
}

func TestDeleteTransaction_NotFound(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   "/transactions/999",
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got '%s'", code)
	}
}

func TestListTransactions_FilterByAccount(t *testing.T) {
	p := newTestPlugin(t)

	// Create a second account.
	acctID := createAccount(t, p, `{"name":"Cash","type":"cash","currency":"EUR"}`)

	// Create transactions on different accounts.
	createTransaction(t, p, `{"amount":100,"type":"expense","category":"groceries","date":"2026-02-01","account_id":1}`)
	createTransaction(t, p, fmt.Sprintf(`{"amount":50,"type":"expense","category":"transport","date":"2026-02-01","account_id":%d}`, acctID))

	// Filter by the second account.
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"account": fmt.Sprintf("%d", acctID)},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, resp)
	if len(items) != 1 {
		t.Fatalf("expected 1 transaction for account %d, got %d", acctID, len(items))
	}

	var tx transactions.Transaction
	if err := json.Unmarshal(items[0], &tx); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if tx.AccountID != acctID {
		t.Errorf("expected account_id %d, got %d", acctID, tx.AccountID)
	}
}

func TestListTransactions_FilterByCategory(t *testing.T) {
	p := newTestPlugin(t)

	createTransaction(t, p, `{"amount":100,"type":"expense","category":"groceries","date":"2026-02-01"}`)
	createTransaction(t, p, `{"amount":50,"type":"expense","category":"transport","date":"2026-02-01"}`)
	createTransaction(t, p, `{"amount":200,"type":"expense","category":"groceries","date":"2026-02-05"}`)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"category": "groceries"},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, resp)
	if len(items) != 2 {
		t.Fatalf("expected 2 groceries transactions, got %d", len(items))
	}

	for _, raw := range items {
		var tx transactions.Transaction
		if err := json.Unmarshal(raw, &tx); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if tx.Category != "groceries" {
			t.Errorf("expected category 'groceries', got '%s'", tx.Category)
		}
	}
}

func TestListTransactions_FilterByType(t *testing.T) {
	p := newTestPlugin(t)

	createTransaction(t, p, `{"amount":3000,"type":"income","category":"salary","date":"2026-02-01"}`)
	createTransaction(t, p, `{"amount":100,"type":"expense","category":"groceries","date":"2026-02-01"}`)
	createTransaction(t, p, `{"amount":50,"type":"expense","category":"transport","date":"2026-02-02"}`)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"type": "income"},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, resp)
	if len(items) != 1 {
		t.Fatalf("expected 1 income transaction, got %d", len(items))
	}

	var tx transactions.Transaction
	if err := json.Unmarshal(items[0], &tx); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if tx.Type != "income" {
		t.Errorf("expected type 'income', got '%s'", tx.Type)
	}
}

func TestListTransactions_SearchDescription(t *testing.T) {
	p := newTestPlugin(t)

	createTransaction(t, p, `{"amount":3000,"type":"income","category":"salary","description":"Monthly salary from ACME","date":"2026-02-01"}`)
	createTransaction(t, p, `{"amount":100,"type":"expense","category":"groceries","description":"Supermarket run","date":"2026-02-02"}`)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"search": "salary"},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, resp)
	if len(items) != 1 {
		t.Fatalf("expected 1 transaction matching 'salary', got %d", len(items))
	}

	var tx transactions.Transaction
	if err := json.Unmarshal(items[0], &tx); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if tx.Amount != 3000 {
		t.Errorf("expected the salary transaction (3000), got amount %f", tx.Amount)
	}
}

func TestListTransactions_FilterByTag(t *testing.T) {
	p := newTestPlugin(t)

	tagID := createTag(t, p, "vacation", "#3B82F6")

	// Create two transactions, only one with the tag.
	taggedBody := fmt.Sprintf(`{"amount":500,"type":"expense","category":"entertainment","date":"2026-02-01","tag_ids":[%d]}`, tagID)
	createTransaction(t, p, taggedBody)
	createTransaction(t, p, `{"amount":100,"type":"expense","category":"groceries","date":"2026-02-02"}`)

	// Filter by tag.
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"tag": fmt.Sprintf("%d", tagID)},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, resp)
	if len(items) != 1 {
		t.Fatalf("expected 1 tagged transaction, got %d", len(items))
	}

	var tx transactions.Transaction
	if err := json.Unmarshal(items[0], &tx); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if tx.Amount != 500 {
		t.Errorf("expected the tagged transaction (500), got amount %f", tx.Amount)
	}
}

func TestCreateTransaction_WithTags(t *testing.T) {
	p := newTestPlugin(t)

	tag1 := createTag(t, p, "recurring", "#EF4444")
	tag2 := createTag(t, p, "essential", "#10B981")

	body := fmt.Sprintf(
		`{"amount":1500,"type":"income","category":"salary","description":"Monthly salary","date":"2026-02-01","tag_ids":[%d,%d]}`,
		tag1, tag2,
	)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/transactions",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse the created transaction and verify tags.
	data := parseDataObject(t, resp)
	var tx transactions.Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(tx.Tags) != 2 {
		t.Fatalf("expected 2 tags, got %d", len(tx.Tags))
	}

	tagNames := map[string]bool{}
	for _, tag := range tx.Tags {
		tagNames[tag.Name] = true
	}
	if !tagNames["recurring"] {
		t.Error("expected tag 'recurring' in transaction tags")
	}
	if !tagNames["essential"] {
		t.Error("expected tag 'essential' in transaction tags")
	}
}

func TestUpdateTransaction_ChangeTags(t *testing.T) {
	p := newTestPlugin(t)

	tag1 := createTag(t, p, "old-tag", "#EF4444")
	tag2 := createTag(t, p, "new-tag", "#10B981")

	// Create with tag1.
	body := fmt.Sprintf(`{"amount":100,"type":"expense","category":"groceries","date":"2026-02-01","tag_ids":[%d]}`, tag1)
	txID := createTransaction(t, p, body)

	// Update to tag2.
	updateBody := fmt.Sprintf(`{"amount":100,"type":"expense","category":"groceries","date":"2026-02-01","tag_ids":[%d]}`, tag2)
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "PUT",
		Path:   fmt.Sprintf("/transactions/%d", txID),
		Body:   []byte(updateBody),
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify tags changed.
	data := parseDataObject(t, resp)
	var tx transactions.Transaction
	if err := json.Unmarshal(data, &tx); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(tx.Tags) != 1 {
		t.Fatalf("expected 1 tag after update, got %d", len(tx.Tags))
	}
	if tx.Tags[0].Name != "new-tag" {
		t.Errorf("expected tag 'new-tag', got '%s'", tx.Tags[0].Name)
	}
}

// --- Categories tests ---

func TestListCategories_DefaultsExist(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/categories",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}

	if resp.StatusCode != 200 {
		t.Fatalf("expected status 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	items := parseDataArray(t, resp)
	// The migration seeds 8 default categories.
	if len(items) != 8 {
		t.Fatalf("expected 8 default categories, got %d", len(items))
	}

	// Verify all are marked as default and have expected names.
	expectedNames := map[string]bool{
		"salary": true, "groceries": true, "transport": true, "entertainment": true,
		"restaurants": true, "bills": true, "health": true, "other": true,
	}
	for _, raw := range items {
		var c struct {
			Name      string `json:"name"`
			IsDefault bool   `json:"is_default"`
		}
		if err := json.Unmarshal(raw, &c); err != nil {
			t.Fatalf("failed to unmarshal category: %v", err)
		}
		if !c.IsDefault {
			t.Errorf("expected category '%s' to be default", c.Name)
		}
		if !expectedNames[c.Name] {
			t.Errorf("unexpected category name: '%s'", c.Name)
		}
	}
}

func TestListCategories_FilterByType(t *testing.T) {
	p := newTestPlugin(t)

	// Create an income-only category.
	createBody := `{"name":"freelance","type":"income","icon":"briefcase"}`
	createResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/categories",
		Body:   []byte(createBody),
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}
	if createResp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", createResp.StatusCode, string(createResp.Body))
	}

	// Create an expense-only category.
	createBody2 := `{"name":"gym","type":"expense","icon":"dumbbell"}`
	createResp2, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/categories",
		Body:   []byte(createBody2),
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}
	if createResp2.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", createResp2.StatusCode, string(createResp2.Body))
	}

	// Filter by income: should return income + both categories.
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/categories",
		Query:  map[string]string{"type": "income"},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	items := parseDataArray(t, resp)
	for _, raw := range items {
		var c struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &c); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if c.Type != "income" && c.Type != "both" {
			t.Errorf("expected type 'income' or 'both', got '%s' for category '%s'", c.Type, c.Name)
		}
		// The expense-only "gym" category should NOT appear.
		if c.Name == "gym" {
			t.Error("expense-only category 'gym' should not appear in income filter")
		}
	}

	// "freelance" (income) should be in the results.
	var foundFreelance bool
	for _, raw := range items {
		var c struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(raw, &c); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if c.Name == "freelance" {
			foundFreelance = true
		}
	}
	if !foundFreelance {
		t.Error("expected 'freelance' category in income filter results")
	}
}

func TestCreateCategory(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name":"subscriptions","type":"expense","icon":"credit-card","color":"#FF5733"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/categories",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify it appears in the list.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/categories",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	// 8 defaults + 1 new = 9.
	if len(items) != 9 {
		t.Fatalf("expected 9 categories, got %d", len(items))
	}

	var found bool
	for _, raw := range items {
		var c struct {
			Name  string `json:"name"`
			Type  string `json:"type"`
			Icon  string `json:"icon"`
			Color string `json:"color"`
		}
		if err := json.Unmarshal(raw, &c); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if c.Name == "subscriptions" {
			found = true
			if c.Type != "expense" {
				t.Errorf("expected type 'expense', got '%s'", c.Type)
			}
			if c.Icon != "credit-card" {
				t.Errorf("expected icon 'credit-card', got '%s'", c.Icon)
			}
			if c.Color != "#FF5733" {
				t.Errorf("expected color '#FF5733', got '%s'", c.Color)
			}
		}
	}
	if !found {
		t.Error("created category 'subscriptions' not found in list")
	}
}

func TestCreateCategory_DuplicateName(t *testing.T) {
	p := newTestPlugin(t)

	// "salary" already exists as a default category.
	body := `{"name":"salary","type":"income","icon":"banknote"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/categories",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 409 {
		t.Fatalf("expected 409, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "CONFLICT" {
		t.Errorf("expected CONFLICT, got '%s'", code)
	}
}

func TestCreateCategory_MissingName(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"type":"expense","icon":"box"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/categories",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestCreateCategory_InvalidType(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name":"bad","type":"invalid","icon":"x"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/categories",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestUpdateCategory(t *testing.T) {
	p := newTestPlugin(t)

	// Create a category to update.
	createBody := `{"name":"pets","type":"expense","icon":"dog"}`
	createResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/categories",
		Body:   []byte(createBody),
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}
	if createResp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", createResp.StatusCode, string(createResp.Body))
	}

	data := parseDataObject(t, createResp)
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}

	// Update name, type, icon, color.
	updateBody := `{"name":"animals","type":"both","icon":"cat","color":"#8B5CF6"}`
	updateResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "PUT",
		Path:   fmt.Sprintf("/categories/%d", created.ID),
		Body:   []byte(updateBody),
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}
	if updateResp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", updateResp.StatusCode, string(updateResp.Body))
	}

	// Verify the update in the list.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/categories",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	var found bool
	for _, raw := range items {
		var c struct {
			ID    int64  `json:"id"`
			Name  string `json:"name"`
			Type  string `json:"type"`
			Icon  string `json:"icon"`
			Color string `json:"color"`
		}
		if err := json.Unmarshal(raw, &c); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if c.ID == created.ID {
			found = true
			if c.Name != "animals" {
				t.Errorf("expected name 'animals', got '%s'", c.Name)
			}
			if c.Type != "both" {
				t.Errorf("expected type 'both', got '%s'", c.Type)
			}
			if c.Icon != "cat" {
				t.Errorf("expected icon 'cat', got '%s'", c.Icon)
			}
			if c.Color != "#8B5CF6" {
				t.Errorf("expected color '#8B5CF6', got '%s'", c.Color)
			}
		}
	}
	if !found {
		t.Error("updated category not found in list")
	}
}

func TestDeleteCategory_NoTransactions(t *testing.T) {
	p := newTestPlugin(t)

	// Create a category to delete.
	createBody := `{"name":"temporary","type":"expense","icon":"trash"}`
	createResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/categories",
		Body:   []byte(createBody),
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	data := parseDataObject(t, createResp)
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Delete it.
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   fmt.Sprintf("/categories/%d", created.ID),
	})
	if err != nil {
		t.Fatalf("delete returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify it no longer appears in the list.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/categories",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	for _, raw := range items {
		var c struct {
			ID int64 `json:"id"`
		}
		if err := json.Unmarshal(raw, &c); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if c.ID == created.ID {
			t.Error("deleted category still appears in list")
		}
	}
}

func TestDeleteCategory_HasTransactions(t *testing.T) {
	p := newTestPlugin(t)

	// "groceries" is a default category. Create a transaction referencing it.
	createTransaction(t, p, `{"amount":50,"type":"expense","category":"groceries","date":"2026-02-01"}`)

	// Find the "groceries" category ID.
	var groceriesID int64
	err := p.db.QueryRow("SELECT id FROM categories WHERE name = 'groceries'").Scan(&groceriesID)
	if err != nil {
		t.Fatalf("failed to find groceries category: %v", err)
	}

	// Attempt to delete it. Should fail with CONFLICT.
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   fmt.Sprintf("/categories/%d", groceriesID),
	})
	if err != nil {
		t.Fatalf("delete returned error: %v", err)
	}
	if resp.StatusCode != 409 {
		t.Fatalf("expected 409, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "CONFLICT" {
		t.Errorf("expected CONFLICT, got '%s'", code)
	}
}

func TestReorderCategories(t *testing.T) {
	p := newTestPlugin(t)

	// Get current categories to know their IDs.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/categories",
		Query:  map[string]string{},
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	if len(items) < 3 {
		t.Fatalf("expected at least 3 categories, got %d", len(items))
	}

	// Parse the first 3 category IDs.
	type catID struct {
		ID int64 `json:"id"`
	}
	var ids []int64
	for i := 0; i < 3; i++ {
		var c catID
		if err := json.Unmarshal(items[i], &c); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		ids = append(ids, c.ID)
	}

	// Reorder: reverse the sort_order of the first 3.
	reorderBody := fmt.Sprintf(
		`[{"id":%d,"sort_order":2},{"id":%d,"sort_order":1},{"id":%d,"sort_order":0}]`,
		ids[0], ids[1], ids[2],
	)
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "PUT",
		Path:   "/categories/reorder",
		Body:   []byte(reorderBody),
	})
	if err != nil {
		t.Fatalf("reorder returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify the sort_order was updated by querying the DB directly.
	var sortOrder int
	err = p.db.QueryRow("SELECT sort_order FROM categories WHERE id = ?", ids[0]).Scan(&sortOrder)
	if err != nil {
		t.Fatalf("failed to query sort_order: %v", err)
	}
	if sortOrder != 2 {
		t.Errorf("expected sort_order 2 for id %d, got %d", ids[0], sortOrder)
	}

	err = p.db.QueryRow("SELECT sort_order FROM categories WHERE id = ?", ids[2]).Scan(&sortOrder)
	if err != nil {
		t.Fatalf("failed to query sort_order: %v", err)
	}
	if sortOrder != 0 {
		t.Errorf("expected sort_order 0 for id %d, got %d", ids[2], sortOrder)
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

// --- Accounts tests ---

func TestCreateAccount_Valid(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name":"Savings EUR","type":"savings","currency":"EUR"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/accounts",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify it appears in the list.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/accounts",
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	// Default "Main Account" + the new one = 2.
	if len(items) != 2 {
		t.Fatalf("expected 2 accounts, got %d", len(items))
	}

	// Find the new account.
	var found bool
	for _, raw := range items {
		var a struct {
			Name string `json:"name"`
			Type string `json:"type"`
		}
		if err := json.Unmarshal(raw, &a); err != nil {
			t.Fatalf("failed to unmarshal account: %v", err)
		}
		if a.Name == "Savings EUR" && a.Type == "savings" {
			found = true
		}
	}
	if !found {
		t.Error("created account 'Savings EUR' not found in list")
	}
}

func TestCreateAccount_MissingName(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"type":"checking","currency":"EUR"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/accounts",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestCreateAccount_InvalidType(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name":"Credit Card","type":"credit","currency":"EUR"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/accounts",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestCreateAccount_DefaultCurrency(t *testing.T) {
	p := newTestPlugin(t)

	// Currency omitted, should default to EUR.
	body := `{"name":"Cash Wallet","type":"cash"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/accounts",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify the currency was set to EUR.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/accounts",
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	var found bool
	for _, raw := range items {
		var a struct {
			Name     string `json:"name"`
			Currency string `json:"currency"`
		}
		if err := json.Unmarshal(raw, &a); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if a.Name == "Cash Wallet" {
			found = true
			if a.Currency != "EUR" {
				t.Errorf("expected currency EUR, got '%s'", a.Currency)
			}
		}
	}
	if !found {
		t.Error("account 'Cash Wallet' not found in list")
	}
}

func TestCreateAccount_SavingsWithInterestRate(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name":"High Yield Savings","type":"savings","currency":"EUR","interest_rate":2.5}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/accounts",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Parse created ID.
	data := parseDataObject(t, resp)
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatalf("failed to parse created response: %v", err)
	}

	// Verify interest_rate is stored by fetching balance endpoint.
	balanceResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   fmt.Sprintf("/accounts/%d/balance", created.ID),
	})
	if err != nil {
		t.Fatalf("balance returned error: %v", err)
	}
	if balanceResp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", balanceResp.StatusCode)
	}

	balanceData := parseDataObject(t, balanceResp)
	var result struct {
		InterestRate *float64 `json:"interest_rate"`
	}
	if err := json.Unmarshal(balanceData, &result); err != nil {
		t.Fatalf("failed to parse balance response: %v", err)
	}
	if result.InterestRate == nil {
		t.Fatal("expected interest_rate to be set, got nil")
	}
	if *result.InterestRate != 2.5 {
		t.Errorf("expected interest_rate 2.5, got %f", *result.InterestRate)
	}
}

func TestCreateAccount_InterestRateOnlyForSavings(t *testing.T) {
	p := newTestPlugin(t)

	// Attempting to set interest_rate on a checking account should fail.
	body := `{"name":"Bad Account","type":"checking","currency":"EUR","interest_rate":1.5}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/accounts",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestUpdateAccount(t *testing.T) {
	p := newTestPlugin(t)

	// Update the default account (id=1).
	body := `{"name":"Updated Main","type":"checking","currency":"USD"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "PUT",
		Path:   "/accounts/1",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify the update in the list.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/accounts",
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	if len(items) != 1 {
		t.Fatalf("expected 1 account, got %d", len(items))
	}

	var a struct {
		Name     string `json:"name"`
		Currency string `json:"currency"`
	}
	if err := json.Unmarshal(items[0], &a); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if a.Name != "Updated Main" {
		t.Errorf("expected name 'Updated Main', got '%s'", a.Name)
	}
	if a.Currency != "USD" {
		t.Errorf("expected currency 'USD', got '%s'", a.Currency)
	}
}

func TestUpdateAccount_NotFound(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name":"Ghost","type":"checking","currency":"EUR"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "PUT",
		Path:   "/accounts/999",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got '%s'", code)
	}
}

func TestArchiveAccount(t *testing.T) {
	p := newTestPlugin(t)

	// Create an account to archive.
	createBody := `{"name":"To Archive","type":"cash","currency":"EUR"}`
	createResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/accounts",
		Body:   []byte(createBody),
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	data := parseDataObject(t, createResp)
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Archive it.
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   fmt.Sprintf("/accounts/%d", created.ID),
	})
	if err != nil {
		t.Fatalf("archive returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}
}

func TestListAccounts_ExcludesArchived(t *testing.T) {
	p := newTestPlugin(t)

	// Create an account.
	createBody := `{"name":"Temporary","type":"cash","currency":"EUR"}`
	createResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/accounts",
		Body:   []byte(createBody),
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	data := parseDataObject(t, createResp)
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Archive it.
	_, err = p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   fmt.Sprintf("/accounts/%d", created.ID),
	})
	if err != nil {
		t.Fatalf("archive returned error: %v", err)
	}

	// List: should only show the default "Main Account".
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/accounts",
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	if len(items) != 1 {
		t.Fatalf("expected 1 active account, got %d", len(items))
	}

	var a struct {
		Name string `json:"name"`
	}
	if err := json.Unmarshal(items[0], &a); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if a.Name != "Main Account" {
		t.Errorf("expected 'Main Account', got '%s'", a.Name)
	}
}

func TestListAccounts_IncludesBalance(t *testing.T) {
	p := newTestPlugin(t)

	// Add income and expense transactions to the default account (id=1).
	createTransaction(t, p, `{"amount":3000,"type":"income","category":"salary","date":"2026-02-01"}`)
	createTransaction(t, p, `{"amount":800,"type":"expense","category":"groceries","date":"2026-02-05"}`)

	// List accounts with balance.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/accounts",
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	if len(items) != 1 {
		t.Fatalf("expected 1 account, got %d", len(items))
	}

	var a struct {
		Name    string  `json:"name"`
		Balance float64 `json:"balance"`
	}
	if err := json.Unmarshal(items[0], &a); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	expectedBalance := 3000.0 - 800.0
	if a.Balance != expectedBalance {
		t.Errorf("expected balance %f, got %f", expectedBalance, a.Balance)
	}
}

func TestAccountBalance_CalculatedFromTransactions(t *testing.T) {
	p := newTestPlugin(t)

	// Create transactions: 5000 income, 1200 expense, 300 expense.
	createTransaction(t, p, `{"amount":5000,"type":"income","category":"salary","date":"2026-02-01"}`)
	createTransaction(t, p, `{"amount":1200,"type":"expense","category":"bills","date":"2026-02-03"}`)
	createTransaction(t, p, `{"amount":300,"type":"expense","category":"groceries","date":"2026-02-05"}`)

	// GET /accounts/1/balance
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/accounts/1/balance",
	})
	if err != nil {
		t.Fatalf("balance returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	data := parseDataObject(t, resp)
	var result struct {
		Balance float64 `json:"balance"`
	}
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse balance: %v", err)
	}

	expectedBalance := 5000.0 - 1200.0 - 300.0
	if result.Balance != expectedBalance {
		t.Errorf("expected balance %f, got %f", expectedBalance, result.Balance)
	}
}

func TestAccountBalance_NotFound(t *testing.T) {
	p := newTestPlugin(t)

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/accounts/999/balance",
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Fatalf("expected 404, got %d", resp.StatusCode)
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "NOT_FOUND" {
		t.Errorf("expected NOT_FOUND, got '%s'", code)
	}
}

// --- Tags tests ---

func TestCreateTag(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name":"vacation","color":"#3B82F6"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/tags",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify it appears in the list.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/tags",
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	if len(items) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(items))
	}

	var tag struct {
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.Unmarshal(items[0], &tag); err != nil {
		t.Fatalf("failed to unmarshal tag: %v", err)
	}
	if tag.Name != "vacation" {
		t.Errorf("expected name 'vacation', got '%s'", tag.Name)
	}
	if tag.Color != "#3B82F6" {
		t.Errorf("expected color '#3B82F6', got '%s'", tag.Color)
	}
}

func TestCreateTag_DuplicateName(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"name":"recurring","color":"#EF4444"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/tags",
		Body:   []byte(body),
	})
	if err != nil || resp.StatusCode != 201 {
		t.Fatalf("first create failed: err=%v status=%d", err, resp.StatusCode)
	}

	// Attempt to create a tag with the same name.
	resp, err = p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/tags",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 409 {
		t.Fatalf("expected 409, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "CONFLICT" {
		t.Errorf("expected CONFLICT, got '%s'", code)
	}
}

func TestCreateTag_MissingName(t *testing.T) {
	p := newTestPlugin(t)

	body := `{"color":"#10B981"}`
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/tags",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("HandleAPI returned error: %v", err)
	}
	if resp.StatusCode != 400 {
		t.Fatalf("expected 400, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	code, _ := parseErrorResponse(t, resp)
	if code != "VALIDATION_ERROR" {
		t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
	}
}

func TestUpdateTag(t *testing.T) {
	p := newTestPlugin(t)

	// Create a tag to update.
	createBody := `{"name":"work","color":"#F59E0B"}`
	createResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/tags",
		Body:   []byte(createBody),
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}
	if createResp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", createResp.StatusCode, string(createResp.Body))
	}

	data := parseDataObject(t, createResp)
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatalf("failed to parse create response: %v", err)
	}

	// Update name and color.
	updateBody := `{"name":"office","color":"#8B5CF6"}`
	updateResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "PUT",
		Path:   fmt.Sprintf("/tags/%d", created.ID),
		Body:   []byte(updateBody),
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}
	if updateResp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", updateResp.StatusCode, string(updateResp.Body))
	}

	// Verify the update in the list.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/tags",
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	if len(items) != 1 {
		t.Fatalf("expected 1 tag, got %d", len(items))
	}

	var tag struct {
		ID    int64  `json:"id"`
		Name  string `json:"name"`
		Color string `json:"color"`
	}
	if err := json.Unmarshal(items[0], &tag); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if tag.ID != created.ID {
		t.Errorf("expected id %d, got %d", created.ID, tag.ID)
	}
	if tag.Name != "office" {
		t.Errorf("expected name 'office', got '%s'", tag.Name)
	}
	if tag.Color != "#8B5CF6" {
		t.Errorf("expected color '#8B5CF6', got '%s'", tag.Color)
	}
}

func TestDeleteTag(t *testing.T) {
	p := newTestPlugin(t)

	// Create a tag to delete.
	createBody := `{"name":"temporary","color":"#DC2626"}`
	createResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/tags",
		Body:   []byte(createBody),
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	data := parseDataObject(t, createResp)
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Delete it.
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   fmt.Sprintf("/tags/%d", created.ID),
	})
	if err != nil {
		t.Fatalf("delete returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	// Verify it no longer appears in the list.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/tags",
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	if len(items) != 0 {
		t.Fatalf("expected 0 tags after delete, got %d", len(items))
	}
}

func TestListTags_OrderedByName(t *testing.T) {
	p := newTestPlugin(t)

	// Create tags in non-alphabetical order.
	names := []string{"zebra", "alpha", "mango"}
	for _, name := range names {
		body := fmt.Sprintf(`{"name":"%s"}`, name)
		resp, err := p.HandleAPI(&sdk.APIRequest{
			Method: "POST",
			Path:   "/tags",
			Body:   []byte(body),
		})
		if err != nil || resp.StatusCode != 201 {
			t.Fatalf("create tag '%s' failed: err=%v status=%d", name, err, resp.StatusCode)
		}
	}

	// List and verify alphabetical order.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/tags",
	})
	if err != nil {
		t.Fatalf("list returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	if len(items) != 3 {
		t.Fatalf("expected 3 tags, got %d", len(items))
	}

	expectedOrder := []string{"alpha", "mango", "zebra"}
	for i, raw := range items {
		var tag struct {
			Name string `json:"name"`
		}
		if err := json.Unmarshal(raw, &tag); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if tag.Name != expectedOrder[i] {
			t.Errorf("position %d: expected '%s', got '%s'", i, expectedOrder[i], tag.Name)
		}
	}
}

func TestAccountBalance_InterestEstimation(t *testing.T) {
	p := newTestPlugin(t)

	// Create a savings account with 2.5% interest rate.
	createBody := `{"name":"Savings","type":"savings","currency":"EUR","interest_rate":2.5}`
	createResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/accounts",
		Body:   []byte(createBody),
	})
	if err != nil {
		t.Fatalf("create returned error: %v", err)
	}

	data := parseDataObject(t, createResp)
	var created struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &created); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	// Add income of 10000 to this account directly in the DB.
	_, err = p.db.Exec(
		"INSERT INTO transactions (amount, type, category, date, account_id) VALUES (?, ?, ?, ?, ?)",
		10000.0, "income", "salary", "2026-02-01", created.ID,
	)
	if err != nil {
		t.Fatalf("failed to insert transaction: %v", err)
	}

	// GET balance: should include estimated_interest = 10000 * (2.5 / 100) = 250.
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   fmt.Sprintf("/accounts/%d/balance", created.ID),
	})
	if err != nil {
		t.Fatalf("balance returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	balanceData := parseDataObject(t, resp)
	var result struct {
		Balance           float64  `json:"balance"`
		EstimatedInterest *float64 `json:"estimated_interest"`
	}
	if err := json.Unmarshal(balanceData, &result); err != nil {
		t.Fatalf("failed to parse: %v", err)
	}

	if result.Balance != 10000.0 {
		t.Errorf("expected balance 10000.0, got %f", result.Balance)
	}
	if result.EstimatedInterest == nil {
		t.Fatal("expected estimated_interest to be set, got nil")
	}
	if *result.EstimatedInterest != 250.0 {
		t.Errorf("expected estimated_interest 250.0, got %f", *result.EstimatedInterest)
	}
}

// --- Recurring Rules Tests ---

// createRecurringRule is a test helper that creates a recurring rule via the API and returns its ID.
func createRecurringRule(t *testing.T, p *FinancePlugin, body string) int64 {
	t.Helper()

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/recurring",
		Body:   []byte(body),
	})
	if err != nil {
		t.Fatalf("create recurring rule failed: %v", err)
	}
	if resp.StatusCode != 201 {
		t.Fatalf("expected 201, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	data := parseDataObject(t, resp)
	var rule struct {
		ID int64 `json:"id"`
	}
	if err := json.Unmarshal(data, &rule); err != nil {
		t.Fatalf("failed to parse created rule: %v", err)
	}
	return rule.ID
}

// generateRecurring calls POST /recurring/generate and returns the count of generated transactions.
func generateRecurring(t *testing.T, p *FinancePlugin) int {
	t.Helper()

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/recurring/generate",
	})
	if err != nil {
		t.Fatalf("generate recurring failed: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	data := parseDataObject(t, resp)
	var result recurring.GenerateResult
	if err := json.Unmarshal(data, &result); err != nil {
		t.Fatalf("failed to parse generate result: %v", err)
	}
	return result.Generated
}

func TestCreateRecurringRule_Monthly(t *testing.T) {
	p := newTestPlugin(t)

	body := `{
		"amount": 50.00,
		"type": "expense",
		"category": "subscriptions",
		"description": "Netflix",
		"frequency": "monthly",
		"day_of_month": 15,
		"start_date": "2026-01-15"
	}`

	ruleID := createRecurringRule(t, p, body)
	if ruleID == 0 {
		t.Fatal("expected rule ID > 0")
	}

	// Verify it appears in the list.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/recurring",
	})
	if err != nil {
		t.Fatalf("list recurring returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	if len(items) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(items))
	}

	var rule recurring.Rule
	if err := json.Unmarshal(items[0], &rule); err != nil {
		t.Fatalf("failed to unmarshal rule: %v", err)
	}
	if rule.Amount != 50.00 {
		t.Errorf("expected amount 50.00, got %f", rule.Amount)
	}
	if rule.Frequency != "monthly" {
		t.Errorf("expected frequency 'monthly', got '%s'", rule.Frequency)
	}
	if rule.DayOfMonth == nil || *rule.DayOfMonth != 15 {
		t.Errorf("expected day_of_month 15, got %v", rule.DayOfMonth)
	}
	if !rule.IsActive {
		t.Error("expected rule to be active")
	}
}

func TestCreateRecurringRule_MissingFields(t *testing.T) {
	p := newTestPlugin(t)

	tests := []struct {
		name string
		body string
	}{
		{
			name: "missing amount",
			body: `{"type":"expense","category":"test","frequency":"monthly","day_of_month":1,"start_date":"2026-01-01"}`,
		},
		{
			name: "missing type",
			body: `{"amount":100,"category":"test","frequency":"monthly","day_of_month":1,"start_date":"2026-01-01"}`,
		},
		{
			name: "missing frequency",
			body: `{"amount":100,"type":"expense","category":"test","day_of_month":1,"start_date":"2026-01-01"}`,
		},
		{
			name: "missing start_date",
			body: `{"amount":100,"type":"expense","category":"test","frequency":"monthly","day_of_month":1}`,
		},
		{
			name: "missing day_of_month for monthly",
			body: `{"amount":100,"type":"expense","category":"test","frequency":"monthly","start_date":"2026-01-01"}`,
		},
		{
			name: "missing day_of_week for weekly",
			body: `{"amount":100,"type":"expense","category":"test","frequency":"weekly","start_date":"2026-01-01"}`,
		},
		{
			name: "missing category for expense",
			body: `{"amount":100,"type":"expense","frequency":"monthly","day_of_month":1,"start_date":"2026-01-01"}`,
		},
		{
			name: "invalid frequency",
			body: `{"amount":100,"type":"expense","category":"test","frequency":"daily","day_of_month":1,"start_date":"2026-01-01"}`,
		},
		{
			name: "invalid date format",
			body: `{"amount":100,"type":"expense","category":"test","frequency":"monthly","day_of_month":1,"start_date":"01-01-2026"}`,
		},
		{
			name: "day_of_month out of range",
			body: `{"amount":100,"type":"expense","category":"test","frequency":"monthly","day_of_month":32,"start_date":"2026-01-01"}`,
		},
		{
			name: "day_of_week out of range",
			body: `{"amount":100,"type":"expense","category":"test","frequency":"weekly","day_of_week":7,"start_date":"2026-01-01"}`,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			resp, err := p.HandleAPI(&sdk.APIRequest{
				Method: "POST",
				Path:   "/recurring",
				Body:   []byte(tc.body),
			})
			if err != nil {
				t.Fatalf("HandleAPI returned error: %v", err)
			}
			if resp.StatusCode != 400 {
				t.Fatalf("expected 400, got %d. Body: %s", resp.StatusCode, string(resp.Body))
			}

			code, _ := parseErrorResponse(t, resp)
			if code != "VALIDATION_ERROR" {
				t.Errorf("expected VALIDATION_ERROR, got '%s'", code)
			}
		})
	}
}

func TestGenerateRecurring_Monthly(t *testing.T) {
	p := newTestPlugin(t)

	// Create a monthly rule starting 3 months ago on day 15.
	threeMonthsAgo := time.Now().AddDate(0, -3, 0)
	startDate := fmt.Sprintf("%d-%02d-15", threeMonthsAgo.Year(), threeMonthsAgo.Month())

	body := fmt.Sprintf(`{
		"amount": 9.99,
		"type": "expense",
		"category": "subscriptions",
		"description": "Streaming service",
		"frequency": "monthly",
		"day_of_month": 15,
		"start_date": "%s"
	}`, startDate)

	createRecurringRule(t, p, body)

	// Generate should create transactions for 3 months ago, 2 months ago, 1 month ago, and possibly this month.
	generated := generateRecurring(t, p)

	// Calculate expected count: from start_date to today, day 15 each month.
	today := time.Now()
	expected := 0
	current := time.Date(threeMonthsAgo.Year(), threeMonthsAgo.Month(), 15, 0, 0, 0, 0, time.UTC)
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	for !current.After(todayDate) {
		expected++
		current = current.AddDate(0, 1, 0)
		// Clamp to month boundaries.
		current = clampDay(current.Year(), current.Month(), 15)
	}

	if generated != expected {
		t.Errorf("expected %d generated transactions, got %d", expected, generated)
	}

	// Verify transactions exist by listing them for the start month.
	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query: map[string]string{
			"month": fmt.Sprintf("%d-%02d", threeMonthsAgo.Year(), threeMonthsAgo.Month()),
		},
	})
	if err != nil {
		t.Fatalf("list transactions returned error: %v", err)
	}

	items := parseDataArray(t, resp)
	if len(items) < 1 {
		t.Fatal("expected at least 1 transaction in the start month")
	}

	var tx transactions.Transaction
	if err := json.Unmarshal(items[0], &tx); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}
	if !tx.IsRecurringInstance {
		t.Error("expected is_recurring_instance=true")
	}
	if tx.Amount != 9.99 {
		t.Errorf("expected amount 9.99, got %f", tx.Amount)
	}
}

// clampDay is a test helper that clamps a day to the last day of the month.
func clampDay(year int, month time.Month, day int) time.Time {
	lastDay := time.Date(year, month+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if day > lastDay {
		day = lastDay
	}
	return time.Date(year, month, day, 0, 0, 0, 0, time.UTC)
}

func TestGenerateRecurring_NoDuplicates(t *testing.T) {
	p := newTestPlugin(t)

	// Create a monthly rule starting 2 months ago.
	twoMonthsAgo := time.Now().AddDate(0, -2, 0)
	startDate := fmt.Sprintf("%d-%02d-10", twoMonthsAgo.Year(), twoMonthsAgo.Month())

	body := fmt.Sprintf(`{
		"amount": 25.00,
		"type": "expense",
		"category": "utilities",
		"description": "Phone bill",
		"frequency": "monthly",
		"day_of_month": 10,
		"start_date": "%s"
	}`, startDate)

	createRecurringRule(t, p, body)

	// Generate once.
	firstCount := generateRecurring(t, p)
	if firstCount == 0 {
		t.Fatal("expected at least 1 generated transaction on first run")
	}

	// Generate again -- should not create duplicates.
	secondCount := generateRecurring(t, p)
	if secondCount != 0 {
		t.Errorf("expected 0 new transactions on second run, got %d", secondCount)
	}
}

func TestGenerateRecurring_RespectsEndDate(t *testing.T) {
	p := newTestPlugin(t)

	// Create a monthly rule that started 4 months ago but ended 2 months ago.
	fourMonthsAgo := time.Now().AddDate(0, -4, 0)
	twoMonthsAgo := time.Now().AddDate(0, -2, 0)
	startDate := fmt.Sprintf("%d-%02d-01", fourMonthsAgo.Year(), fourMonthsAgo.Month())
	endDate := fmt.Sprintf("%d-%02d-28", twoMonthsAgo.Year(), twoMonthsAgo.Month())

	body := fmt.Sprintf(`{
		"amount": 100.00,
		"type": "expense",
		"category": "rent",
		"description": "Parking spot",
		"frequency": "monthly",
		"day_of_month": 1,
		"start_date": "%s",
		"end_date": "%s"
	}`, startDate, endDate)

	ruleID := createRecurringRule(t, p, body)
	generateRecurring(t, p)

	// Verify the rule is now inactive.
	ruleResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/recurring",
	})
	if err != nil {
		t.Fatalf("list recurring returned error: %v", err)
	}

	items := parseDataArray(t, ruleResp)
	found := false
	for _, item := range items {
		var rule recurring.Rule
		if err := json.Unmarshal(item, &rule); err != nil {
			t.Fatalf("failed to unmarshal rule: %v", err)
		}
		if rule.ID == ruleID {
			found = true
			if rule.IsActive {
				t.Error("expected rule to be inactive after end_date passed")
			}
			break
		}
	}
	if !found {
		t.Fatal("created rule not found in list")
	}

	// Verify that generated transactions only go up to end_date.
	// Transactions should NOT have dates after end_date.
	endDateParsed, _ := time.Parse("2006-01-02", endDate)
	txResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"type": "expense"},
	})
	if err != nil {
		t.Fatalf("list transactions returned error: %v", err)
	}

	txItems := parseDataArray(t, txResp)
	for _, item := range txItems {
		var tx transactions.Transaction
		if err := json.Unmarshal(item, &tx); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if tx.RecurringRuleID != nil && *tx.RecurringRuleID == ruleID {
			txDate, _ := time.Parse("2006-01-02", tx.Date)
			if txDate.After(endDateParsed) {
				t.Errorf("generated transaction date %s is after end_date %s", tx.Date, endDate)
			}
		}
	}
}

func TestGenerateRecurring_Weekly(t *testing.T) {
	p := newTestPlugin(t)

	// Create a weekly rule starting 4 weeks ago, on Monday (day_of_week=1).
	fourWeeksAgo := time.Now().AddDate(0, 0, -28)
	startDate := fourWeeksAgo.Format("2006-01-02")

	body := fmt.Sprintf(`{
		"amount": 30.00,
		"type": "expense",
		"category": "groceries",
		"description": "Weekly groceries",
		"frequency": "weekly",
		"day_of_week": 1,
		"start_date": "%s"
	}`, startDate)

	createRecurringRule(t, p, body)

	generated := generateRecurring(t, p)

	// Calculate expected: from startDate to today, every Monday.
	today := time.Now()
	todayDate := time.Date(today.Year(), today.Month(), today.Day(), 0, 0, 0, 0, time.UTC)
	start := time.Date(fourWeeksAgo.Year(), fourWeeksAgo.Month(), fourWeeksAgo.Day(), 0, 0, 0, 0, time.UTC)

	// Find first Monday on or after startDate.
	current := start
	for current.Weekday() != time.Monday {
		current = current.AddDate(0, 0, 1)
	}

	expected := 0
	for !current.After(todayDate) {
		expected++
		current = current.AddDate(0, 0, 7)
	}

	if generated != expected {
		t.Errorf("expected %d weekly transactions, got %d", expected, generated)
	}
}

func TestUpdateRecurringRule(t *testing.T) {
	p := newTestPlugin(t)

	body := `{
		"amount": 50.00,
		"type": "expense",
		"category": "subscriptions",
		"description": "Netflix",
		"frequency": "monthly",
		"day_of_month": 15,
		"start_date": "2026-01-15"
	}`

	ruleID := createRecurringRule(t, p, body)

	// Update amount and frequency to weekly.
	updateBody := `{
		"amount": 75.00,
		"type": "expense",
		"category": "subscriptions",
		"description": "Netflix Premium",
		"frequency": "weekly",
		"day_of_week": 3,
		"start_date": "2026-01-15"
	}`

	resp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "PUT",
		Path:   fmt.Sprintf("/recurring/%d", ruleID),
		Body:   []byte(updateBody),
	})
	if err != nil {
		t.Fatalf("update returned error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", resp.StatusCode, string(resp.Body))
	}

	data := parseDataObject(t, resp)
	var rule recurring.Rule
	if err := json.Unmarshal(data, &rule); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if rule.Amount != 75.00 {
		t.Errorf("expected amount 75.00, got %f", rule.Amount)
	}
	if rule.Frequency != "weekly" {
		t.Errorf("expected frequency 'weekly', got '%s'", rule.Frequency)
	}
	if rule.Description != "Netflix Premium" {
		t.Errorf("expected description 'Netflix Premium', got '%s'", rule.Description)
	}
}

func TestDeleteRecurringRule(t *testing.T) {
	p := newTestPlugin(t)

	// Create a rule and generate some transactions.
	twoMonthsAgo := time.Now().AddDate(0, -2, 0)
	startDate := fmt.Sprintf("%d-%02d-01", twoMonthsAgo.Year(), twoMonthsAgo.Month())

	body := fmt.Sprintf(`{
		"amount": 100.00,
		"type": "expense",
		"category": "rent",
		"description": "Storage unit",
		"frequency": "monthly",
		"day_of_month": 1,
		"start_date": "%s"
	}`, startDate)

	ruleID := createRecurringRule(t, p, body)
	generateRecurring(t, p)

	// Count generated transactions before delete.
	txResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"type": "expense"},
	})
	if err != nil {
		t.Fatalf("list transactions failed: %v", err)
	}
	txCountBefore := len(parseDataArray(t, txResp))
	if txCountBefore == 0 {
		t.Fatal("expected generated transactions before delete")
	}

	// Delete (deactivate) the rule.
	delResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "DELETE",
		Path:   fmt.Sprintf("/recurring/%d", ruleID),
	})
	if err != nil {
		t.Fatalf("delete returned error: %v", err)
	}
	if delResp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d. Body: %s", delResp.StatusCode, string(delResp.Body))
	}

	// Verify the rule is now inactive.
	listResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/recurring",
	})
	if err != nil {
		t.Fatalf("list recurring returned error: %v", err)
	}

	items := parseDataArray(t, listResp)
	for _, item := range items {
		var rule recurring.Rule
		if err := json.Unmarshal(item, &rule); err != nil {
			t.Fatalf("failed to unmarshal: %v", err)
		}
		if rule.ID == ruleID && rule.IsActive {
			t.Error("expected rule to be inactive after delete")
		}
	}

	// Verify generated transactions still exist (soft delete does not remove them).
	txRespAfter, err := p.HandleAPI(&sdk.APIRequest{
		Method: "GET",
		Path:   "/transactions",
		Query:  map[string]string{"type": "expense"},
	})
	if err != nil {
		t.Fatalf("list transactions failed: %v", err)
	}
	txCountAfter := len(parseDataArray(t, txRespAfter))
	if txCountAfter != txCountBefore {
		t.Errorf("expected %d transactions after deactivation, got %d", txCountBefore, txCountAfter)
	}
}
