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
	txBody := `{"amount":50,"type":"expense","category":"groceries","date":"2026-02-01"}`
	txResp, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/transactions",
		Body:   []byte(txBody),
	})
	if err != nil || txResp.StatusCode != 201 {
		t.Fatalf("failed to create transaction: err=%v status=%d", err, txResp.StatusCode)
	}

	// Find the "groceries" category ID.
	var groceriesID int64
	err = p.db.QueryRow("SELECT id FROM categories WHERE name = 'groceries'").Scan(&groceriesID)
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
	income := `{"amount":3000,"type":"income","category":"salary","date":"2026-02-01"}`
	expense := `{"amount":800,"type":"expense","category":"groceries","date":"2026-02-05"}`

	for _, body := range []string{income, expense} {
		resp, err := p.HandleAPI(&sdk.APIRequest{
			Method: "POST",
			Path:   "/transactions",
			Body:   []byte(body),
		})
		if err != nil || resp.StatusCode != 201 {
			t.Fatalf("failed to create transaction: err=%v status=%d", err, resp.StatusCode)
		}
	}

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
	transactions := []string{
		`{"amount":5000,"type":"income","category":"salary","date":"2026-02-01"}`,
		`{"amount":1200,"type":"expense","category":"bills","date":"2026-02-03"}`,
		`{"amount":300,"type":"expense","category":"groceries","date":"2026-02-05"}`,
	}
	for _, body := range transactions {
		resp, err := p.HandleAPI(&sdk.APIRequest{
			Method: "POST",
			Path:   "/transactions",
			Body:   []byte(body),
		})
		if err != nil || resp.StatusCode != 201 {
			t.Fatalf("failed to create transaction: err=%v status=%d", err, resp.StatusCode)
		}
	}

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
