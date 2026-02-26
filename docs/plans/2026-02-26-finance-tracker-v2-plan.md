# Finance Tracker v2 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Evolve the Finance Tracker plugin from a basic income/expense tracker into a full personal finance management tool with accounts, budgets, recurring transactions, savings goals, investments, and charts.

**Architecture:** Monolith plugin with internal modules (handler → service → repository per feature). Single SQLite database, gRPC plugin interface via Cortex SDK. Frontend uses Svelte 5 (runes) with tab-based navigation and Chart.js for visualizations.

**Tech Stack:** Go 1.25 (chi patterns, modernc.org/sqlite), SvelteKit (Svelte 5 runes), Tailwind CSS, Chart.js (svelte-chartjs), lucide-svelte icons.

**Design Doc:** `docs/plans/2026-02-26-finance-tracker-v2-design.md`

**Project Standards:** Follow `~/.claude/plugins/cache/alvarotc/project-standards/1.0.0/` — especially `agents/backend.md` (handler→service→repo layers, validate at boundary), `core/principles.md` (TDD, KISS, type safety, security), `standards/architecture.md` (feature-based folders), `standards/code-quality.md` (naming, functions <30 lines).

**Git:** Feature branch `feature/finance-tracker-v2`. Conventional Commits in English. No Co-Authored-By. No push.

---

## Task 0: Create feature branch

**Step 1: Create branch**

```bash
cd /home/alvarotc/Documents/apps/cortex
git checkout -b feature/finance-tracker-v2
```

**Step 2: Verify**

```bash
git branch --show-current
```

Expected: `feature/finance-tracker-v2`

---

## Task 1: Shared module — errors, response helpers, DB utilities

Extract the existing `jsonSuccess`/`jsonError` helpers and add typed errors. This is the foundation all other modules depend on.

**Files:**
- Create: `plugins/finance-tracker/backend/shared/errors.go`
- Create: `plugins/finance-tracker/backend/shared/response.go`
- Create: `plugins/finance-tracker/backend/shared/db.go`
- Create: `plugins/finance-tracker/backend/shared/shared_test.go`

**Step 1: Write test for shared error types and response helpers**

```go
// plugins/finance-tracker/backend/shared/shared_test.go
package shared

import (
	"encoding/json"
	"testing"
)

func TestNewAppError(t *testing.T) {
	err := NewAppError("VALIDATION_ERROR", "amount must be greater than 0", 400)
	if err.Code != "VALIDATION_ERROR" {
		t.Errorf("expected code VALIDATION_ERROR, got %s", err.Code)
	}
	if err.StatusCode != 400 {
		t.Errorf("expected status 400, got %d", err.StatusCode)
	}
	if err.Error() != "amount must be greater than 0" {
		t.Errorf("unexpected error message: %s", err.Error())
	}
}

func TestNewValidationError(t *testing.T) {
	err := NewValidationError("email is required")
	if err.Code != "VALIDATION_ERROR" || err.StatusCode != 400 {
		t.Errorf("unexpected error: %+v", err)
	}
}

func TestNewNotFoundError(t *testing.T) {
	err := NewNotFoundError("transaction", "42")
	if err.Code != "NOT_FOUND" || err.StatusCode != 404 {
		t.Errorf("unexpected error: %+v", err)
	}
	if err.Message != "transaction 42 not found" {
		t.Errorf("unexpected message: %s", err.Message)
	}
}

func TestNewConflictError(t *testing.T) {
	err := NewConflictError("category already exists")
	if err.Code != "CONFLICT" || err.StatusCode != 409 {
		t.Errorf("unexpected error: %+v", err)
	}
}

func TestJSONSuccess(t *testing.T) {
	resp, err := JSONSuccess(200, map[string]string{"name": "test"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", resp.StatusCode)
	}
	if resp.ContentType != "application/json" {
		t.Errorf("expected application/json, got %s", resp.ContentType)
	}

	var body struct {
		Data map[string]string `json:"data"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	if body.Data["name"] != "test" {
		t.Errorf("expected name=test, got %s", body.Data["name"])
	}
}

func TestJSONError(t *testing.T) {
	resp, err := JSONError(&AppError{Code: "NOT_FOUND", Message: "not found", StatusCode: 404})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.StatusCode != 404 {
		t.Errorf("expected status 404, got %d", resp.StatusCode)
	}

	var body struct {
		Error struct {
			Code    string `json:"code"`
			Message string `json:"message"`
		} `json:"error"`
	}
	if err := json.Unmarshal(resp.Body, &body); err != nil {
		t.Fatalf("failed to parse body: %v", err)
	}
	if body.Error.Code != "NOT_FOUND" {
		t.Errorf("expected code NOT_FOUND, got %s", body.Error.Code)
	}
}
```

**Step 2: Run test to verify it fails**

```bash
cd /home/alvarotc/Documents/apps/cortex && go test ./plugins/finance-tracker/backend/shared/ -v
```

Expected: FAIL — package/types don't exist yet.

**Step 3: Implement shared package**

```go
// plugins/finance-tracker/backend/shared/errors.go
package shared

import "fmt"

// AppError represents a typed application error with HTTP status code.
type AppError struct {
	Code       string
	Message    string
	StatusCode int
}

func (e *AppError) Error() string {
	return e.Message
}

// NewAppError creates a new AppError.
func NewAppError(code string, message string, statusCode int) *AppError {
	return &AppError{Code: code, Message: message, StatusCode: statusCode}
}

// NewValidationError creates a 400 validation error.
func NewValidationError(message string) *AppError {
	return NewAppError("VALIDATION_ERROR", message, 400)
}

// NewNotFoundError creates a 404 not found error.
func NewNotFoundError(resource string, id string) *AppError {
	return NewAppError("NOT_FOUND", fmt.Sprintf("%s %s not found", resource, id), 404)
}

// NewConflictError creates a 409 conflict error.
func NewConflictError(message string) *AppError {
	return NewAppError("CONFLICT", message, 409)
}
```

```go
// plugins/finance-tracker/backend/shared/response.go
package shared

import (
	"encoding/json"
	"fmt"

	"github.com/alvarotorresc/cortex/pkg/sdk"
)

// JSONSuccess wraps data in { "data": ... } format per PATTERNS.md.
func JSONSuccess(status int, data interface{}) (*sdk.APIResponse, error) {
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

// JSONError wraps an AppError in { "error": { "code": ..., "message": ... } } format.
func JSONError(appErr *AppError) (*sdk.APIResponse, error) {
	body, _ := json.Marshal(map[string]interface{}{
		"error": map[string]interface{}{
			"code":    appErr.Code,
			"message": appErr.Message,
		},
	})
	return &sdk.APIResponse{
		StatusCode:  appErr.StatusCode,
		Body:        body,
		ContentType: "application/json",
	}, nil
}
```

```go
// plugins/finance-tracker/backend/shared/db.go
package shared

import (
	"database/sql"
	"fmt"
	"strconv"
	"strings"
)

// ExtractIDFromPath extracts the numeric ID from a path like "/resources/{id}".
// Returns the ID as int64 or an AppError if invalid.
func ExtractIDFromPath(path string) (int64, *AppError) {
	parts := strings.Split(strings.TrimPrefix(path, "/"), "/")
	if len(parts) < 2 || parts[1] == "" {
		return 0, NewValidationError("missing resource ID in path")
	}
	id, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return 0, NewValidationError("invalid resource ID: must be a number")
	}
	return id, nil
}

// OpenDatabase opens a SQLite database and enables WAL mode.
func OpenDatabase(path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("opening database: %w", err)
	}

	if _, err := db.Exec("PRAGMA journal_mode=WAL"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling WAL mode: %w", err)
	}

	if _, err := db.Exec("PRAGMA foreign_keys=ON"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enabling foreign keys: %w", err)
	}

	return db, nil
}
```

**Step 4: Run tests to verify they pass**

```bash
cd /home/alvarotc/Documents/apps/cortex && go test ./plugins/finance-tracker/backend/shared/ -v
```

Expected: All PASS.

**Step 5: Commit**

```bash
git add plugins/finance-tracker/backend/shared/
git commit -m "feat(finance): add shared module with typed errors, response helpers, and db utilities"
```

---

## Task 2: Migration 002 — new schema

Write the migration that adds all new tables and alters existing ones. Preserve existing data.

**Files:**
- Create: `plugins/finance-tracker/backend/migrations/002_enhanced.sql`

**Step 1: Write the migration**

```sql
-- Finance Tracker v2: enhanced schema
-- Adds accounts, tags, recurring rules, budgets, savings goals, investments.
-- Alters existing transactions and categories tables for v2 compatibility.

-- 1. New tables

CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('checking', 'savings', 'cash', 'investment')),
    currency TEXT NOT NULL DEFAULT 'EUR',
    interest_rate REAL,
    icon TEXT,
    color TEXT,
    is_archived INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    color TEXT
);

CREATE TABLE IF NOT EXISTS transaction_tags (
    transaction_id INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (transaction_id, tag_id)
);

CREATE TABLE IF NOT EXISTS recurring_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount REAL NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('income', 'expense', 'transfer')),
    account_id INTEGER NOT NULL REFERENCES accounts(id),
    dest_account_id INTEGER REFERENCES accounts(id),
    category_id INTEGER REFERENCES categories(id),
    description TEXT,
    frequency TEXT NOT NULL CHECK(frequency IN ('weekly', 'biweekly', 'monthly', 'yearly')),
    day_of_month INTEGER,
    day_of_week INTEGER,
    month_of_year INTEGER,
    start_date TEXT NOT NULL,
    end_date TEXT,
    last_generated TEXT,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS budgets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT,
    category_id INTEGER REFERENCES categories(id),
    amount REAL NOT NULL,
    month TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS savings_goals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    target_amount REAL NOT NULL,
    current_amount REAL NOT NULL DEFAULT 0,
    target_date TEXT,
    icon TEXT,
    color TEXT,
    is_completed INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE TABLE IF NOT EXISTS investments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('crypto', 'etf', 'fund', 'stock', 'other')),
    account_id INTEGER REFERENCES accounts(id),
    units REAL,
    avg_buy_price REAL,
    current_price REAL,
    currency TEXT NOT NULL DEFAULT 'EUR',
    notes TEXT,
    last_updated TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- 2. Default account for existing data migration
INSERT OR IGNORE INTO accounts (id, name, type, currency, icon, color)
VALUES (1, 'Main Account', 'checking', 'EUR', 'wallet', '#0070F3');

-- 3. Alter transactions table for v2
-- SQLite doesn't support ADD COLUMN with FK constraints, so we add columns without constraints
-- and enforce FKs at application level (which we already do).
ALTER TABLE transactions ADD COLUMN account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id);
ALTER TABLE transactions ADD COLUMN dest_account_id INTEGER REFERENCES accounts(id);
ALTER TABLE transactions ADD COLUMN is_recurring_instance INTEGER NOT NULL DEFAULT 0;
ALTER TABLE transactions ADD COLUMN recurring_rule_id INTEGER REFERENCES recurring_rules(id);

-- Update type CHECK constraint: SQLite can't alter CHECK constraints on existing tables.
-- The old CHECK allows ('income', 'expense'). We enforce 'transfer' at the application level.
-- New transactions will be validated by the handler.

-- 4. Alter categories table for v2
ALTER TABLE categories ADD COLUMN type TEXT NOT NULL DEFAULT 'both' CHECK(type IN ('income', 'expense', 'both'));
ALTER TABLE categories ADD COLUMN color TEXT;
ALTER TABLE categories ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

-- 5. Additional indexes
CREATE INDEX IF NOT EXISTS idx_transactions_account_id ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category);
CREATE INDEX IF NOT EXISTS idx_recurring_rules_active ON recurring_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_budgets_month ON budgets(month);
CREATE INDEX IF NOT EXISTS idx_investments_type ON investments(type);
```

**Step 2: Update Migrate() in plugin.go to run both migrations**

In `plugin.go`, update the `Migrate` method to read and execute all migration files in order. Use the existing `embed.FS`.

The current code reads only `001_init.sql`. Change it to iterate `migrations/` directory:

```go
// In Migrate(), replace the single-file read with:
entries, err := migrations.ReadDir("migrations")
if err != nil {
    return fmt.Errorf("reading migrations dir: %w", err)
}

// Sort entries by name (they're already sorted alphabetically by embed.FS)
for _, entry := range entries {
    if entry.IsDir() {
        continue
    }
    migrationSQL, err := migrations.ReadFile("migrations/" + entry.Name())
    if err != nil {
        return fmt.Errorf("reading migration %s: %w", entry.Name(), err)
    }
    if _, err := p.db.Exec(string(migrationSQL)); err != nil {
        return fmt.Errorf("running migration %s: %w", entry.Name(), err)
    }
}
```

**Step 3: Write test for migration — verify all tables exist after both migrations**

Add to the existing `plugin_test.go`:

```go
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

func TestMigrate_ExistingTransactionsLinkedToDefaultAccount(t *testing.T) {
	p := newTestPlugin(t)

	// Create a transaction (goes to default account_id=1 via DEFAULT)
	_, err := p.HandleAPI(&sdk.APIRequest{
		Method: "POST",
		Path:   "/transactions",
		Body:   []byte(`{"amount": 100, "type": "expense", "category": "groceries", "date": "2026-02-01"}`),
	})
	if err != nil {
		t.Fatalf("create failed: %v", err)
	}

	var accountID int64
	err = p.db.QueryRow("SELECT account_id FROM transactions WHERE id = 1").Scan(&accountID)
	if err != nil {
		t.Fatalf("query failed: %v", err)
	}
	if accountID != 1 {
		t.Errorf("expected account_id=1, got %d", accountID)
	}
}
```

**Step 4: Run all tests**

```bash
cd /home/alvarotc/Documents/apps/cortex && go test ./plugins/finance-tracker/backend/... -v -count=1
```

Expected: All PASS, including the new migration tests.

**Note:** The `002_enhanced.sql` ALTERs may fail if run twice (column already exists). The migration should be idempotent. Since SQLite `ALTER TABLE ADD COLUMN` does not support `IF NOT EXISTS`, we need to handle this at the application level. The simplest approach: use a migrations tracking table OR catch the "duplicate column" error. For simplicity, since Cortex plugins always run all migrations on startup, we can wrap each ALTER in a check:

Instead of bare `ALTER TABLE`, use this pattern in the SQL:
```sql
-- Check if column exists before adding (SQLite workaround)
-- We use a separate approach: the Migrate() function will track applied migrations.
```

**Better approach for the Migrate() function:** Track applied migrations in a `_migrations` table. Only run each migration once.

Update `Migrate()`:

```go
func (p *FinancePlugin) Migrate(databasePath string) error {
	db, err := shared.OpenDatabase(databasePath)
	if err != nil {
		return err
	}
	p.db = db

	// Create migrations tracking table
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

		// Check if already applied
		var count int
		if err := p.db.QueryRow("SELECT COUNT(*) FROM _migrations WHERE filename = ?", entry.Name()).Scan(&count); err != nil {
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

		if _, err := p.db.Exec("INSERT INTO _migrations (filename) VALUES (?)", entry.Name()); err != nil {
			return fmt.Errorf("recording migration %s: %w", entry.Name(), err)
		}
	}

	return nil
}
```

**Step 5: Run tests again after migration tracking update**

```bash
cd /home/alvarotc/Documents/apps/cortex && go test ./plugins/finance-tracker/backend/... -v -count=1
```

Expected: All PASS. Migrations are now idempotent.

**Step 6: Run linter**

```bash
cd /home/alvarotc/Documents/apps/cortex && make lint
```

Expected: No errors.

**Step 7: Commit**

```bash
git add plugins/finance-tracker/backend/migrations/002_enhanced.sql plugins/finance-tracker/backend/plugin.go plugins/finance-tracker/backend/plugin_test.go
git commit -m "feat(finance): add v2 migration with accounts, tags, budgets, goals, investments tables

Adds migration tracking to prevent duplicate execution. Creates default
account so existing transactions remain linked."
```

---

## Task 3: Accounts module

**Files:**
- Create: `plugins/finance-tracker/backend/accounts/models.go`
- Create: `plugins/finance-tracker/backend/accounts/repository.go`
- Create: `plugins/finance-tracker/backend/accounts/service.go`
- Create: `plugins/finance-tracker/backend/accounts/handler.go`
- Create: `plugins/finance-tracker/backend/accounts/accounts_test.go`

### Models

```go
// plugins/finance-tracker/backend/accounts/models.go
package accounts

// Account represents a financial account (checking, savings, cash, investment).
type Account struct {
	ID           int64    `json:"id"`
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Currency     string   `json:"currency"`
	InterestRate *float64 `json:"interest_rate,omitempty"`
	Icon         string   `json:"icon"`
	Color        string   `json:"color"`
	IsArchived   bool     `json:"is_archived"`
	CreatedAt    string   `json:"created_at"`
}

// AccountWithBalance includes the calculated balance from transactions.
type AccountWithBalance struct {
	Account
	Balance          float64  `json:"balance"`
	EstimatedInterest *float64 `json:"estimated_interest,omitempty"`
}

// CreateAccountInput is the validated input for creating an account.
type CreateAccountInput struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Currency     string   `json:"currency"`
	InterestRate *float64 `json:"interest_rate"`
	Icon         string   `json:"icon"`
	Color        string   `json:"color"`
}

// UpdateAccountInput is the validated input for updating an account.
type UpdateAccountInput struct {
	Name         string   `json:"name"`
	Type         string   `json:"type"`
	Currency     string   `json:"currency"`
	InterestRate *float64 `json:"interest_rate"`
	Icon         string   `json:"icon"`
	Color        string   `json:"color"`
}
```

### Implementation approach

**Step 1: Write failing tests** — Test CRUD operations, balance calculation, interest estimation. Test through the handler (integration-style, matching existing test patterns in the project).

**Step 2: Implement repository** — SQL queries for CRUD + balance query (SUM of transactions by account).

**Step 3: Implement service** — Balance calculation, interest estimation for savings accounts.

**Step 4: Implement handler** — Parse APIRequest, validate input (name required, type in enum, interest_rate only for savings), delegate to service.

**Step 5: Wire into plugin.go** — Add route matching for `/accounts*` paths, delegate to `accounts.Handler`.

**Step 6: Run all tests**

```bash
cd /home/alvarotc/Documents/apps/cortex && go test ./plugins/finance-tracker/backend/... -v -count=1
```

**Step 7: Run linter**

```bash
make lint
```

**Step 8: Commit**

```bash
git add plugins/finance-tracker/backend/accounts/ plugins/finance-tracker/backend/plugin.go
git commit -m "feat(finance): add accounts module with CRUD, balance calculation, and interest estimation"
```

### Key test cases for accounts

- `TestCreateAccount_Valid` — Create checking account, verify in list
- `TestCreateAccount_MissingName` — Validation error
- `TestCreateAccount_InvalidType` — Validation error
- `TestCreateAccount_InterestRateOnlyForSavings` — Reject interest_rate on checking/cash
- `TestUpdateAccount` — Update name and type
- `TestArchiveAccount` — Soft delete (DELETE sets is_archived=1)
- `TestListAccounts_ExcludesArchived` — Archived accounts not in list
- `TestAccountBalance_CalculatedFromTransactions` — Income adds, expense subtracts
- `TestAccountBalance_InterestEstimation` — savings with 2.5% rate, balance 10000 → estimated 250

---

## Task 4: Categories module (enhanced)

Refactor from current inline handler to modular CRUD with type, color, sort_order.

**Files:**
- Create: `plugins/finance-tracker/backend/categories/models.go`
- Create: `plugins/finance-tracker/backend/categories/repository.go`
- Create: `plugins/finance-tracker/backend/categories/handler.go`
- Create: `plugins/finance-tracker/backend/categories/categories_test.go`

### Key changes from v1

- Categories now have `type` (income/expense/both), `color`, `sort_order`
- Full CRUD: create, update, delete (with FK check), reorder
- Filter by type in GET
- No service layer needed — CRUD is simple enough for handler → repository

### Key test cases

- `TestListCategories_FilterByType` — GET with ?type=income returns only income categories
- `TestCreateCategory` — Create custom category with type and color
- `TestUpdateCategory` — Change name, type, icon
- `TestDeleteCategory_NoTransactions` — Delete succeeds
- `TestDeleteCategory_HasTransactions` — Returns CONFLICT error
- `TestReorderCategories` — PUT /categories/reorder with ordered IDs

### Steps

**Step 1:** Write failing tests → **Step 2:** Implement repository → **Step 3:** Implement handler → **Step 4:** Wire into plugin.go → **Step 5:** Run tests + lint → **Step 6:** Commit

```bash
git commit -m "feat(finance): add categories module with full CRUD, type filtering, and reorder"
```

---

## Task 5: Tags module

Simple CRUD for tags + transaction_tags linking table.

**Files:**
- Create: `plugins/finance-tracker/backend/tags/models.go`
- Create: `plugins/finance-tracker/backend/tags/repository.go`
- Create: `plugins/finance-tracker/backend/tags/handler.go`
- Create: `plugins/finance-tracker/backend/tags/tags_test.go`

### Key test cases

- `TestCreateTag` — Create tag with name and color
- `TestCreateTag_DuplicateName` — Returns CONFLICT
- `TestUpdateTag` — Change name and color
- `TestDeleteTag_UnlinksFromTransactions` — Tag removed, transaction_tags rows deleted via CASCADE
- `TestListTags` — Returns all tags ordered by name

### Steps

Same pattern: tests → repository → handler → wire → test + lint → commit.

```bash
git commit -m "feat(finance): add tags module with CRUD"
```

---

## Task 6: Transactions module (refactored)

The biggest refactor. Move from inline plugin.go handlers to a proper module with:
- Full CRUD (add UPDATE/edit)
- Account association (account_id, dest_account_id for transfers)
- Tag linking (tag_ids in create/update)
- Combinable filters (month, account, category, tag, type, search)
- Transfer type support

**Files:**
- Create: `plugins/finance-tracker/backend/transactions/models.go`
- Create: `plugins/finance-tracker/backend/transactions/repository.go`
- Create: `plugins/finance-tracker/backend/transactions/service.go`
- Create: `plugins/finance-tracker/backend/transactions/handler.go`
- Create: `plugins/finance-tracker/backend/transactions/transactions_test.go`
- Modify: `plugins/finance-tracker/backend/plugin.go` — Remove old handlers, delegate to module

### Models

```go
// Transaction with v2 fields
type Transaction struct {
	ID                  int64    `json:"id"`
	Amount              float64  `json:"amount"`
	Type                string   `json:"type"` // income, expense, transfer
	AccountID           int64    `json:"account_id"`
	DestAccountID       *int64   `json:"dest_account_id,omitempty"`
	CategoryID          *int64   `json:"category_id,omitempty"`
	Category            string   `json:"category"` // denormalized for display
	Description         string   `json:"description"`
	Date                string   `json:"date"`
	IsRecurringInstance bool     `json:"is_recurring_instance"`
	RecurringRuleID     *int64   `json:"recurring_rule_id,omitempty"`
	Tags                []Tag    `json:"tags"`
	CreatedAt           string   `json:"created_at"`
}

// TransactionFilter holds combinable query parameters.
type TransactionFilter struct {
	Month      string
	AccountID  *int64
	CategoryID *int64
	TagID      *int64
	Type       string
	Search     string
}

// CreateTransactionInput is the validated input.
type CreateTransactionInput struct {
	Amount        float64 `json:"amount"`
	Type          string  `json:"type"`
	AccountID     int64   `json:"account_id"`
	DestAccountID *int64  `json:"dest_account_id"`
	CategoryID    *int64  `json:"category_id"`
	Category      string  `json:"category"` // backward compat with v1
	Description   string  `json:"description"`
	Date          string  `json:"date"`
	TagIDs        []int64 `json:"tag_ids"`
}
```

### Key test cases

- `TestCreateTransaction_WithAccount` — Transaction linked to account
- `TestCreateTransaction_Transfer` — type=transfer with dest_account_id
- `TestCreateTransaction_TransferMissingDest` — Validation error
- `TestUpdateTransaction` — Edit amount, category, description
- `TestDeleteTransaction` — Same as v1 but through new module
- `TestListTransactions_FilterByAccount` — ?account=2
- `TestListTransactions_FilterByTag` — ?tag=1
- `TestListTransactions_FilterByType` — ?type=income
- `TestListTransactions_SearchDescription` — ?search=salary
- `TestListTransactions_CombinedFilters` — Multiple filters at once
- `TestCreateTransaction_WithTags` — Tag linking on create
- `TestUpdateTransaction_ChangeTags` — Tag re-linking on update
- `TestTransaction_BackwardCompatibility` — v1-style create (category string, no account) still works (defaults to account_id=1)

### Service responsibilities

- Validate account exists before linking
- Validate dest_account exists for transfers
- Manage tag linking (delete old + insert new on update)
- Default account_id to 1 if not provided (backward compat)

### Steps

**Step 1:** Write failing tests → **Step 2:** Implement models → **Step 3:** Implement repository (dynamic query builder for filters) → **Step 4:** Implement service → **Step 5:** Implement handler → **Step 6:** Update plugin.go (remove old handlers, delegate to transactions.Handler) → **Step 7:** Run ALL tests (including old ones that should still pass) → **Step 8:** Lint → **Step 9:** Commit

```bash
git commit -m "feat(finance): refactor transactions module with CRUD, transfers, tags, and combinable filters

Replaces inline handlers in plugin.go. Adds edit support, account linking,
tag association, and multi-filter list endpoint. Backward compatible with
v1 transaction format."
```

---

## Task 7: Recurring rules module

**Files:**
- Create: `plugins/finance-tracker/backend/recurring/models.go`
- Create: `plugins/finance-tracker/backend/recurring/repository.go`
- Create: `plugins/finance-tracker/backend/recurring/service.go`
- Create: `plugins/finance-tracker/backend/recurring/handler.go`
- Create: `plugins/finance-tracker/backend/recurring/recurring_test.go`

### Service — Generation engine

The core logic lives here. `Generate()` method:

1. Query all active rules where `last_generated IS NULL OR last_generated < ?` (today)
2. For each rule, calculate pending dates between `last_generated` (or `start_date`) and today
3. Date calculation by frequency:
   - `monthly`: use `day_of_month`, iterate month by month
   - `weekly`: use `day_of_week`, iterate week by week
   - `biweekly`: same as weekly but skip every other week
   - `yearly`: use `day_of_month` + `month_of_year`, iterate year by year
4. For each date: INSERT transaction with `is_recurring_instance=1`, `recurring_rule_id`
5. Update `last_generated`
6. If `end_date` passed, set `is_active=0`

### Key test cases

- `TestCreateRecurringRule_Monthly` — Create monthly rule
- `TestGenerateRecurring_MonthlyCreatesTransactions` — Rule from Jan to Mar, generate in Mar → 3 transactions
- `TestGenerateRecurring_NoDuplicates` — Generate twice → still only 3 transactions
- `TestGenerateRecurring_RespectsEndDate` — Rule with end_date, stops generating after
- `TestGenerateRecurring_Weekly` — Creates correct weekly dates
- `TestGenerateRecurring_DeactivatesExpired` — Rule past end_date marked inactive
- `TestUpdateRecurringRule` — Change amount, frequency
- `TestDeleteRecurringRule` — Deactivates (does not delete generated transactions)

### Steps

Tests → models → repository → service (generation engine) → handler → wire → test + lint → commit.

```bash
git commit -m "feat(finance): add recurring rules module with automatic transaction generation"
```

---

## Task 8: Budgets module

**Files:**
- Create: `plugins/finance-tracker/backend/budgets/models.go`
- Create: `plugins/finance-tracker/backend/budgets/repository.go`
- Create: `plugins/finance-tracker/backend/budgets/service.go`
- Create: `plugins/finance-tracker/backend/budgets/handler.go`
- Create: `plugins/finance-tracker/backend/budgets/budgets_test.go`

### Service

Calculates `spent`, `remaining`, `percentage` by querying transactions:

```sql
-- For global budget (category_id IS NULL):
SELECT COALESCE(SUM(amount), 0) FROM transactions
WHERE type = 'expense' AND date LIKE ?

-- For category budget:
SELECT COALESCE(SUM(amount), 0) FROM transactions
WHERE type = 'expense' AND category = ? AND date LIKE ?
```

### Key test cases

- `TestCreateBudget_Global` — Budget with no category_id
- `TestCreateBudget_PerCategory` — Budget for specific category
- `TestListBudgets_CalculatesSpent` — Create budget 200, add 150 expense → spent=150, remaining=50, percentage=75
- `TestListBudgets_OverBudget` — Budget 100, spend 150 → remaining=-50, percentage=150
- `TestUpdateBudget` — Change amount
- `TestDeleteBudget`

### Steps

Tests → models → repository → service → handler → wire → test + lint → commit.

```bash
git commit -m "feat(finance): add budgets module with global and per-category spending tracking"
```

---

## Task 9: Savings goals module

**Files:**
- Create: `plugins/finance-tracker/backend/goals/models.go`
- Create: `plugins/finance-tracker/backend/goals/repository.go`
- Create: `plugins/finance-tracker/backend/goals/handler.go`
- Create: `plugins/finance-tracker/backend/goals/goals_test.go`

### Key test cases

- `TestCreateGoal` — Create savings goal
- `TestContribute` — Add 500 to goal, verify current_amount
- `TestContribute_AutoCompletes` — Goal of 1000, contribute 1000 → is_completed=1
- `TestUpdateGoal` — Change target_amount
- `TestDeleteGoal`
- `TestContribute_NegativeAmount` — Validation error

### Steps

Tests → models → repository → handler (contribute endpoint increments + auto-complete) → wire → test + lint → commit.

```bash
git commit -m "feat(finance): add savings goals module with contributions and auto-completion"
```

---

## Task 10: Investments module

**Files:**
- Create: `plugins/finance-tracker/backend/investments/models.go`
- Create: `plugins/finance-tracker/backend/investments/repository.go`
- Create: `plugins/finance-tracker/backend/investments/service.go`
- Create: `plugins/finance-tracker/backend/investments/handler.go`
- Create: `plugins/finance-tracker/backend/investments/investments_test.go`

### Service

Calculates P&L fields:
- `total_invested = units * avg_buy_price`
- `current_value = units * current_price`
- `pnl = current_value - total_invested`
- `pnl_percentage = (pnl / total_invested) * 100` (guard against division by zero)

### Key test cases

- `TestCreateInvestment` — Create crypto investment
- `TestListInvestments_CalculatesPnL` — units=0.5, avg_buy=45000, current=52000 → pnl=3500
- `TestUpdateInvestment_PriceUpdate` — Update current_price, verify P&L recalculated
- `TestDeleteInvestment`
- `TestCreateInvestment_InvalidType` — Validation error
- `TestPnLPercentage_ZeroInvested` — Guard against div/0

### Steps

Tests → models → repository → service → handler → wire → test + lint → commit.

```bash
git commit -m "feat(finance): add investments module with P&L calculation"
```

---

## Task 11: Reports module

**Files:**
- Create: `plugins/finance-tracker/backend/reports/models.go`
- Create: `plugins/finance-tracker/backend/reports/service.go`
- Create: `plugins/finance-tracker/backend/reports/handler.go`
- Create: `plugins/finance-tracker/backend/reports/reports_test.go`

### Endpoints

1. **GET /reports/summary?month=** — Monthly income, expense, balance, by_category, by_account
2. **GET /reports/trends?from=&to=** — Array of monthly totals for line chart
3. **GET /reports/categories?month=** — Category breakdown with previous month comparison
4. **GET /reports/net-worth** — Sum of account balances + investment values

### Models

```go
type MonthlySummary struct {
	Month      string           `json:"month"`
	Income     float64          `json:"income"`
	Expense    float64          `json:"expense"`
	Balance    float64          `json:"balance"`
	ByCategory []CategoryTotal  `json:"by_category"`
	ByAccount  []AccountTotal   `json:"by_account"`
}

type TrendPoint struct {
	Month   string  `json:"month"`
	Income  float64 `json:"income"`
	Expense float64 `json:"expense"`
	Balance float64 `json:"balance"`
}

type CategoryComparison struct {
	Category      string  `json:"category"`
	CurrentMonth  float64 `json:"current_month"`
	PreviousMonth float64 `json:"previous_month"`
	Change        float64 `json:"change"` // percentage
}

type NetWorth struct {
	AccountsTotal   float64 `json:"accounts_total"`
	InvestmentsTotal float64 `json:"investments_total"`
	NetWorth         float64 `json:"net_worth"`
}
```

### Key test cases

- `TestSummary_MonthlyTotals` — Income and expense totals
- `TestSummary_ByCategory` — Breakdown per category
- `TestSummary_ByAccount` — Breakdown per account
- `TestTrends_SixMonthRange` — Returns array of 6 TrendPoints
- `TestCategoryComparison_WithPreviousMonth` — Shows change percentage
- `TestNetWorth_AccountsAndInvestments` — Sum of both

### Steps

Tests → models → service (complex SQL aggregations) → handler → wire → test + lint → commit.

```bash
git commit -m "feat(finance): add reports module with summary, trends, category comparison, and net worth"
```

---

## Task 12: Enhanced dashboard widget

Update `GetWidgetData()` to return:
- Monthly income/expense/balance (existing)
- Sparkline data: last 6 months balance trend
- Budget progress: if global budget exists, show spent/amount/percentage

**Files:**
- Modify: `plugins/finance-tracker/backend/plugin.go` — Update GetWidgetData

### Key test cases

- `TestWidgetData_IncludesSparkline` — Returns 6 months of balance data
- `TestWidgetData_IncludesBudgetProgress` — If global budget exists, returns progress
- `TestWidgetData_NoBudget` — Budget field is null when no budget defined

### Steps

Tests → implement → test + lint → commit.

```bash
git commit -m "feat(finance): enhance dashboard widget with sparkline and budget progress"
```

---

## Task 13: Clean up plugin.go — final routing

After all modules are wired, `plugin.go` should be a thin routing shell:

```go
func (p *FinancePlugin) HandleAPI(req *sdk.APIRequest) (*sdk.APIResponse, error) {
	switch {
	case strings.HasPrefix(req.Path, "/accounts"):
		return p.accountsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/transactions"):
		return p.transactionsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/categories"):
		return p.categoriesHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/tags"):
		return p.tagsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/recurring"):
		return p.recurringHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/budgets"):
		return p.budgetsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/goals"):
		return p.goalsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/investments"):
		return p.investmentsHandler.Handle(req)
	case strings.HasPrefix(req.Path, "/reports"):
		return p.reportsHandler.Handle(req)
	// Legacy: /summary still works (redirects to reports)
	case req.Method == "GET" && req.Path == "/summary":
		req.Path = "/reports/summary"
		return p.reportsHandler.Handle(req)
	default:
		return shared.JSONError(shared.NewAppError("NOT_FOUND", "route not found", 404))
	}
}
```

**Steps:** Refactor → run all tests → lint → commit.

```bash
git commit -m "refactor(finance): clean up plugin.go as thin routing shell delegating to modules"
```

---

## Task 14: Update manifest version

**Files:**
- Modify: `plugins/finance-tracker/backend/plugin.go` — Update version in GetManifest()
- Modify: `plugins/finance-tracker/manifest.json` — Update version

Change version from `0.1.0` to `0.2.0`.

```bash
git commit -m "chore(finance): bump version to 0.2.0"
```

---

## Task 15: Frontend — Types, API client, and shared components

**Files:**
- Create: `frontend/src/lib/components/plugins/finance/types.ts`
- Create: `frontend/src/lib/components/plugins/finance/api.ts`
- Create: `frontend/src/lib/components/plugins/finance/shared/MonthPicker.svelte`
- Create: `frontend/src/lib/components/plugins/finance/shared/AmountDisplay.svelte`
- Create: `frontend/src/lib/components/plugins/finance/shared/ProgressBar.svelte`
- Create: `frontend/src/lib/components/plugins/finance/shared/EmptyState.svelte`

### types.ts — All TypeScript interfaces

All interfaces matching backend models: Account, Transaction, Category, Tag, Budget, SavingsGoal, Investment, MonthlySummary, TrendPoint, NetWorth, etc.

### api.ts — Typed API client

Wrapper around `pluginApi('finance-tracker')` with typed methods:
- `listAccounts()`, `createAccount()`, `updateAccount()`, etc.
- One method per endpoint, returns typed data

### Shared components

- **MonthPicker** — Extracted from current FinancePage (chevron nav + label)
- **AmountDisplay** — Formats currency with color (green positive, red negative)
- **ProgressBar** — Reusable bar with percentage, color changes when > 100%
- **EmptyState** — Icon + message + CTA button

### Steps

Create all files → verify frontend builds → commit.

```bash
cd /home/alvarotc/Documents/apps/cortex/frontend && pnpm build
git commit -m "feat(finance-ui): add types, api client, and shared components for v2"
```

---

## Task 16: Frontend — FinancePage shell with tab navigation

**Files:**
- Rewrite: `frontend/src/lib/components/plugins/FinancePage.svelte`
- Create: `frontend/src/lib/components/plugins/finance/FinancePage.svelte`

Replace the current monolithic component with a tab shell that lazy-loads tab content.

### Tabs

`Overview | Transactions | Budgets | Goals | Investments`

Settings (categories, tags, accounts, recurring) accessible via a gear icon that opens a side panel or modal.

### Steps

Write shell component → verify builds → commit.

```bash
git commit -m "feat(finance-ui): refactor FinancePage into tab-based navigation shell"
```

---

## Task 17: Frontend — Transactions tab

**Files:**
- Create: `frontend/src/lib/components/plugins/finance/transactions/TransactionsTab.svelte`
- Create: `frontend/src/lib/components/plugins/finance/transactions/TransactionForm.svelte`
- Create: `frontend/src/lib/components/plugins/finance/transactions/TransactionRow.svelte`
- Create: `frontend/src/lib/components/plugins/finance/transactions/TransactionFilters.svelte`

Replaces the current inline transaction list. Adds:
- Edit functionality (click row to edit)
- Account selector in form
- Tag multi-select in form
- Filter bar: account, category, tag, type, search text
- Transfer type with destination account

### Steps

Build components → verify frontend builds → commit.

```bash
git commit -m "feat(finance-ui): add transactions tab with edit, filters, transfers, and tags"
```

---

## Task 18: Frontend — Overview tab with charts

**Files:**
- Create: `frontend/src/lib/components/plugins/finance/overview/OverviewTab.svelte`
- Create: `frontend/src/lib/components/plugins/finance/overview/BalanceCard.svelte`
- Create: `frontend/src/lib/components/plugins/finance/overview/CategoryChart.svelte`
- Create: `frontend/src/lib/components/plugins/finance/overview/TrendChart.svelte`
- Create: `frontend/src/lib/components/plugins/finance/overview/AccountsList.svelte`
- Create: `frontend/src/lib/components/plugins/finance/overview/NetWorthCard.svelte`

### Chart.js dependency

```bash
cd /home/alvarotc/Documents/apps/cortex/frontend && pnpm add chart.js svelte-chartjs
```

### Components

- **BalanceCard** — Monthly income/expense/balance (from reports/summary)
- **CategoryChart** — Donut chart with category distribution (Chart.js)
- **TrendChart** — Line chart with 6-month evolution (Chart.js)
- **AccountsList** — List of accounts with balances
- **NetWorthCard** — Total patrimony from reports/net-worth

### Steps

Install Chart.js → build components → verify frontend builds → commit.

```bash
git commit -m "feat(finance-ui): add overview tab with balance, category donut, trend line, and net worth"
```

---

## Task 19: Frontend — Budgets tab

**Files:**
- Create: `frontend/src/lib/components/plugins/finance/budgets/BudgetsTab.svelte`
- Create: `frontend/src/lib/components/plugins/finance/budgets/BudgetCard.svelte`
- Create: `frontend/src/lib/components/plugins/finance/budgets/BudgetForm.svelte`

### Components

- **BudgetsTab** — Lists budgets for current month with create button
- **BudgetCard** — Shows budget name, amount, spent, remaining with ProgressBar. Red when > 100%
- **BudgetForm** — Create/edit with category selector (null = global)

### Steps

Build components → verify frontend builds → commit.

```bash
git commit -m "feat(finance-ui): add budgets tab with progress tracking"
```

---

## Task 20: Frontend — Goals tab

**Files:**
- Create: `frontend/src/lib/components/plugins/finance/goals/GoalsTab.svelte`
- Create: `frontend/src/lib/components/plugins/finance/goals/GoalCard.svelte`
- Create: `frontend/src/lib/components/plugins/finance/goals/GoalForm.svelte`

### Components

- **GoalsTab** — Lists goals with create button
- **GoalCard** — Name, progress bar (current/target), contribute button, target date
- **GoalForm** — Create/edit with name, target amount, target date, icon, color

### Steps

Build components → verify frontend builds → commit.

```bash
git commit -m "feat(finance-ui): add savings goals tab with contributions and progress"
```

---

## Task 21: Frontend — Investments tab

**Files:**
- Create: `frontend/src/lib/components/plugins/finance/investments/InvestmentsTab.svelte`
- Create: `frontend/src/lib/components/plugins/finance/investments/InvestmentCard.svelte`
- Create: `frontend/src/lib/components/plugins/finance/investments/InvestmentForm.svelte`

### Components

- **InvestmentsTab** — Lists investments with total portfolio value
- **InvestmentCard** — Name, type badge, units, avg price, current price, P&L (green/red), last updated
- **InvestmentForm** — Create/edit with type selector, units, prices

### Steps

Build components → verify frontend builds → commit.

```bash
git commit -m "feat(finance-ui): add investments tab with P&L display"
```

---

## Task 22: Frontend — Settings (categories, tags, accounts, recurring)

**Files:**
- Create: `frontend/src/lib/components/plugins/finance/settings/CategoriesManager.svelte`
- Create: `frontend/src/lib/components/plugins/finance/settings/TagsManager.svelte`
- Create: `frontend/src/lib/components/plugins/finance/settings/AccountsManager.svelte`
- Create: `frontend/src/lib/components/plugins/finance/recurring/RecurringManager.svelte`
- Create: `frontend/src/lib/components/plugins/finance/recurring/RecurringForm.svelte`

### Components

Each manager is a list + inline form for CRUD. Accessible from a gear icon in the FinancePage header.

- **CategoriesManager** — CRUD categories with type, icon, color, drag-to-reorder
- **TagsManager** — CRUD tags with name and color
- **AccountsManager** — CRUD accounts with type, currency, interest rate
- **RecurringManager** — CRUD rules with frequency, amount, account, category
- **RecurringForm** — Form with frequency selector, day pickers, start/end dates

### Steps

Build components → verify frontend builds → commit.

```bash
git commit -m "feat(finance-ui): add settings managers for categories, tags, accounts, and recurring rules"
```

---

## Task 23: i18n — Extend translations

**Files:**
- Modify: `frontend/src/lib/i18n/en.json`
- Modify: `frontend/src/lib/i18n/es.json`

Add keys for all new modules: accounts, budgets, goals, investments, recurring, tags, reports, settings, common form actions.

### Steps

Add keys → verify frontend builds → commit.

```bash
git commit -m "feat(finance-i18n): add EN and ES translations for all v2 modules"
```

---

## Task 24: Integration testing — full flow

Run the complete test suite to ensure everything works together.

**Step 1: Backend tests**

```bash
cd /home/alvarotc/Documents/apps/cortex && go test ./plugins/finance-tracker/backend/... -v -count=1 -race
```

**Step 2: Lint**

```bash
make lint
```

**Step 3: Frontend build**

```bash
cd frontend && pnpm build
```

**Step 4: Full build**

```bash
cd /home/alvarotc/Documents/apps/cortex && make build
```

All must pass. Fix any issues discovered.

---

## Task 25: Security and testing review

**MANDATORY per project-standards flow (MEMORY.md rule).**

Launch in parallel:
1. `project-standards:security-agent` — Review all new code for OWASP concerns, SQL injection, input validation
2. `project-standards:testing-agent` — Review test coverage, quality, gaps

Fix all issues reported before considering the feature complete.

```bash
git commit -m "fix(finance): address security and testing review findings"
```

---

## Summary

| Task | Module | Type |
|------|--------|------|
| 0 | Branch | Setup |
| 1 | shared/ | Backend foundation |
| 2 | migrations/ | Schema |
| 3 | accounts/ | Backend |
| 4 | categories/ | Backend |
| 5 | tags/ | Backend |
| 6 | transactions/ | Backend (refactor) |
| 7 | recurring/ | Backend |
| 8 | budgets/ | Backend |
| 9 | goals/ | Backend |
| 10 | investments/ | Backend |
| 11 | reports/ | Backend |
| 12 | widget | Backend |
| 13 | plugin.go | Backend (cleanup) |
| 14 | version | Chore |
| 15 | types + api + shared | Frontend foundation |
| 16 | FinancePage shell | Frontend |
| 17 | Transactions tab | Frontend |
| 18 | Overview tab + charts | Frontend |
| 19 | Budgets tab | Frontend |
| 20 | Goals tab | Frontend |
| 21 | Investments tab | Frontend |
| 22 | Settings managers | Frontend |
| 23 | i18n | Frontend |
| 24 | Integration testing | QA |
| 25 | Security + testing review | Review |
