# Finance Tracker v2 — Design Document

> **Date:** 2026-02-26
> **Status:** Approved
> **Plugin:** finance-tracker (Cortex)
> **Scope:** Full financial management — accounts, transactions, budgets, goals, investments, reports

---

## 1. Overview

Evolve the Finance Tracker from a basic income/expense tracker into a complete personal finance management tool. The plugin remains a single cohesive unit (monolith with internal modules) given the strong domain coupling between accounts, transactions, budgets, and investments.

### Goals

- Full CRUD for all entities (including edit transactions, which is currently missing)
- Multiple accounts (checking, savings, cash, investment) with balance tracking
- Recurring transactions (salary, rent, subscriptions) with automatic generation
- Monthly budgets (global + per category) with real-time progress
- Savings goals with target amounts and progress tracking
- Basic investment tracking (crypto, ETFs, funds) with manual price updates
- Interest estimation for remunerated savings accounts
- Charts and reports: trends, category breakdown, net worth
- Tags for flexible transaction labeling beyond categories
- Enhanced dashboard widget with sparkline and budget progress

### Non-Goals

- CSV import/export (deferred)
- Automatic price fetching for investments (Cortex is offline-first)
- Multi-currency conversion (single currency per account, no auto-conversion)
- Automatic interest transaction generation (user adds manually when bank settles)

---

## 2. Data Model

### 2.1 Tables

#### `accounts`

| Column | Type | Constraints |
|--------|------|-------------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| name | TEXT | NOT NULL |
| type | TEXT | NOT NULL, CHECK(type IN ('checking', 'savings', 'cash', 'investment')) |
| currency | TEXT | DEFAULT 'EUR' |
| interest_rate | REAL | NULL, only for savings |
| icon | TEXT | Lucide icon identifier |
| color | TEXT | Hex color |
| is_archived | INTEGER | DEFAULT 0 |
| created_at | TEXT | DEFAULT datetime('now') |

#### `transactions` (extended from v1)

| Column | Type | Constraints |
|--------|------|-------------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| amount | REAL | NOT NULL, CHECK(amount > 0) |
| type | TEXT | NOT NULL, CHECK(type IN ('income', 'expense', 'transfer')) |
| account_id | INTEGER | FK -> accounts, NOT NULL |
| dest_account_id | INTEGER | FK -> accounts, NULL (only for transfers) |
| category_id | INTEGER | FK -> categories |
| description | TEXT | |
| date | TEXT | NOT NULL (YYYY-MM-DD) |
| is_recurring_instance | INTEGER | DEFAULT 0 |
| recurring_rule_id | INTEGER | FK -> recurring_rules, NULL |
| created_at | TEXT | DEFAULT datetime('now') |

Indexes: `idx_transactions_date`, `idx_transactions_type`, `idx_transactions_account_id`, `idx_transactions_category_id`

#### `categories` (extended from v1)

| Column | Type | Constraints |
|--------|------|-------------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| name | TEXT | NOT NULL UNIQUE |
| type | TEXT | NOT NULL, CHECK(type IN ('income', 'expense', 'both')) |
| icon | TEXT | |
| color | TEXT | |
| is_default | INTEGER | DEFAULT 0 |
| sort_order | INTEGER | DEFAULT 0 |

#### `tags`

| Column | Type | Constraints |
|--------|------|-------------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| name | TEXT | NOT NULL UNIQUE |
| color | TEXT | |

#### `transaction_tags`

| Column | Type | Constraints |
|--------|------|-------------|
| transaction_id | INTEGER | FK -> transactions |
| tag_id | INTEGER | FK -> tags |
| PRIMARY KEY | | (transaction_id, tag_id) |

#### `recurring_rules`

| Column | Type | Constraints |
|--------|------|-------------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| amount | REAL | NOT NULL |
| type | TEXT | NOT NULL, CHECK(type IN ('income', 'expense', 'transfer')) |
| account_id | INTEGER | FK -> accounts |
| dest_account_id | INTEGER | FK -> accounts, NULL |
| category_id | INTEGER | FK -> categories |
| description | TEXT | |
| frequency | TEXT | NOT NULL, CHECK(frequency IN ('weekly', 'biweekly', 'monthly', 'yearly')) |
| day_of_month | INTEGER | 1-31, for monthly/yearly |
| day_of_week | INTEGER | 0-6, for weekly/biweekly |
| month_of_year | INTEGER | 1-12, for yearly only |
| start_date | TEXT | NOT NULL (YYYY-MM-DD) |
| end_date | TEXT | NULL = no end |
| last_generated | TEXT | Last date instances were generated up to |
| is_active | INTEGER | DEFAULT 1 |
| created_at | TEXT | DEFAULT datetime('now') |

#### `budgets`

| Column | Type | Constraints |
|--------|------|-------------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| name | TEXT | Optional label |
| category_id | INTEGER | FK -> categories, NULL = global budget |
| amount | REAL | NOT NULL (monthly limit) |
| month | TEXT | 'YYYY-MM' or NULL for all months |
| created_at | TEXT | DEFAULT datetime('now') |

#### `savings_goals`

| Column | Type | Constraints |
|--------|------|-------------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| name | TEXT | NOT NULL |
| target_amount | REAL | NOT NULL |
| current_amount | REAL | DEFAULT 0 |
| target_date | TEXT | Optional deadline |
| icon | TEXT | |
| color | TEXT | |
| is_completed | INTEGER | DEFAULT 0 |
| created_at | TEXT | DEFAULT datetime('now') |

#### `investments`

| Column | Type | Constraints |
|--------|------|-------------|
| id | INTEGER | PRIMARY KEY AUTOINCREMENT |
| name | TEXT | NOT NULL |
| type | TEXT | NOT NULL, CHECK(type IN ('crypto', 'etf', 'fund', 'stock', 'other')) |
| account_id | INTEGER | FK -> accounts |
| units | REAL | Number of units/shares |
| avg_buy_price | REAL | Average purchase price per unit |
| current_price | REAL | Latest known price (manual update) |
| currency | TEXT | DEFAULT 'EUR' |
| notes | TEXT | |
| last_updated | TEXT | Date of last price update |
| created_at | TEXT | DEFAULT datetime('now') |

### 2.2 Migration Strategy

- `001_init.sql` — Keep as-is for backward compatibility
- `002_enhanced.sql` — New migration:
  1. Create new tables (accounts, tags, transaction_tags, recurring_rules, budgets, savings_goals, investments)
  2. Insert default account ("Main Account", type: checking, currency: EUR)
  3. ALTER transactions: add account_id (default to main account), dest_account_id, is_recurring_instance, recurring_rule_id
  4. ALTER categories: add type (default 'both'), color, sort_order
  5. Existing data preserved — transactions linked to default account

---

## 3. API Endpoints

All prefixed with `/api/plugins/finance-tracker/`.

### Accounts

| Method | Path | Description |
|--------|------|-------------|
| GET | /accounts | List accounts with calculated balance |
| POST | /accounts | Create account |
| PUT | /accounts/{id} | Update account |
| DELETE | /accounts/{id} | Archive account (soft delete) |
| GET | /accounts/{id}/balance | Balance + interest estimate for savings |

### Transactions

| Method | Path | Description |
|--------|------|-------------|
| GET | /transactions?month=&account=&category=&tag=&type=&search= | List with combinable filters |
| POST | /transactions | Create (with optional tag IDs) |
| PUT | /transactions/{id} | Update transaction |
| DELETE | /transactions/{id} | Delete transaction |

### Categories

| Method | Path | Description |
|--------|------|-------------|
| GET | /categories?type= | List (filterable by income/expense) |
| POST | /categories | Create |
| PUT | /categories/{id} | Update |
| DELETE | /categories/{id} | Delete (only if no transactions reference it) |
| PUT | /categories/reorder | Reorder |

### Tags

| Method | Path | Description |
|--------|------|-------------|
| GET | /tags | List |
| POST | /tags | Create |
| PUT | /tags/{id} | Update |
| DELETE | /tags/{id} | Delete (unlinks from transactions) |

### Recurring Rules

| Method | Path | Description |
|--------|------|-------------|
| GET | /recurring | List rules |
| POST | /recurring | Create rule |
| PUT | /recurring/{id} | Update rule |
| DELETE | /recurring/{id} | Deactivate rule |
| POST | /recurring/generate | Generate pending instances up to today |

### Budgets

| Method | Path | Description |
|--------|------|-------------|
| GET | /budgets?month= | List with calculated spent/remaining |
| POST | /budgets | Create |
| PUT | /budgets/{id} | Update |
| DELETE | /budgets/{id} | Delete |

### Savings Goals

| Method | Path | Description |
|--------|------|-------------|
| GET | /goals | List |
| POST | /goals | Create |
| PUT | /goals/{id} | Update |
| DELETE | /goals/{id} | Delete |
| POST | /goals/{id}/contribute | Add contribution |

### Investments

| Method | Path | Description |
|--------|------|-------------|
| GET | /investments | List with current value and P&L |
| POST | /investments | Create |
| PUT | /investments/{id} | Update (including price) |
| DELETE | /investments/{id} | Delete |

### Reports

| Method | Path | Description |
|--------|------|-------------|
| GET | /reports/summary?month= | Monthly summary (income, expense, balance, by category, by account) |
| GET | /reports/trends?from=&to= | Monthly evolution over range |
| GET | /reports/categories?month= | Category breakdown with month-over-month comparison |
| GET | /reports/net-worth | Total patrimony: accounts + investments |

### Widget

| Method | Path | Description |
|--------|------|-------------|
| GET | /widget/dashboard-widget | Monthly balance + sparkline data + budget progress |

---

## 4. Backend Architecture

### Module Structure

```
plugins/finance-tracker/backend/
├── main.go
├── plugin.go                  # CortexPlugin implementation, routes to modules
├── shared/
│   ├── db.go                  # DB connection, migration runner
│   ├── errors.go              # Typed errors (NotFound, Validation, Conflict)
│   └── response.go            # { data } / { error: { code, message } } helpers
├── accounts/
│   ├── handler.go             # Parse request, validate input, call service
│   ├── service.go             # Balance calculation, interest estimation
│   ├── repository.go          # SQL queries
│   ├── models.go
│   └── accounts_test.go
├── transactions/
│   ├── handler.go
│   ├── service.go             # Tag linking, account validation
│   ├── repository.go
│   ├── models.go
│   └── transactions_test.go
├── categories/
│   ├── handler.go
│   ├── repository.go
│   ├── models.go
│   └── categories_test.go
├── tags/
│   ├── handler.go
│   ├── repository.go
│   ├── models.go
│   └── tags_test.go
├── recurring/
│   ├── handler.go
│   ├── service.go             # Instance generation engine
│   ├── repository.go
│   ├── models.go
│   └── recurring_test.go
├── budgets/
│   ├── handler.go
│   ├── service.go             # Calculate spent/remaining from transactions
│   ├── repository.go
│   ├── models.go
│   └── budgets_test.go
├── goals/
│   ├── handler.go
│   ├── repository.go
│   ├── models.go
│   └── goals_test.go
├── investments/
│   ├── handler.go
│   ├── service.go             # Value and P&L calculation
│   ├── repository.go
│   ├── models.go
│   └── investments_test.go
├── reports/
│   ├── handler.go
│   ├── service.go             # Complex aggregation queries
│   ├── models.go
│   └── reports_test.go
└── migrations/
    ├── 001_init.sql
    └── 002_enhanced.sql
```

### Layer Responsibilities

- **Handler**: Parse APIRequest, validate input at boundary, call service, format response
- **Service**: Business logic. Receives validated data. Calls repository. Returns domain types
- **Repository**: Raw SQL queries with parameterized statements. Returns Go structs

Each layer only knows the one below it (handler -> service -> repository).

---

## 5. Frontend Architecture

### Tab Navigation

```
[Overview] [Transactions] [Budgets] [Goals] [Investments]
```

Settings (categories, tags, accounts, recurring) accessible via gear icon or within relevant tabs.

### Component Structure

```
frontend/src/lib/components/plugins/finance/
├── FinancePage.svelte           # Shell: tab routing
├── api.ts                       # Typed API client
├── types.ts                     # Shared interfaces
├── overview/
│   ├── OverviewTab.svelte       # Feature
│   ├── BalanceCard.svelte       # UI
│   ├── CategoryChart.svelte     # UI (Chart.js donut)
│   ├── TrendChart.svelte        # UI (Chart.js line)
│   ├── AccountsList.svelte      # UI
│   └── NetWorthCard.svelte      # UI
├── transactions/
│   ├── TransactionsTab.svelte   # Feature
│   ├── TransactionForm.svelte   # UI
│   ├── TransactionRow.svelte    # UI
│   └── TransactionFilters.svelte # UI
├── budgets/
│   ├── BudgetsTab.svelte        # Feature
│   ├── BudgetCard.svelte        # UI
│   └── BudgetForm.svelte        # UI
├── goals/
│   ├── GoalsTab.svelte          # Feature
│   ├── GoalCard.svelte          # UI
│   └── GoalForm.svelte          # UI
├── investments/
│   ├── InvestmentsTab.svelte    # Feature
│   ├── InvestmentCard.svelte    # UI
│   └── InvestmentForm.svelte    # UI
├── recurring/
│   ├── RecurringManager.svelte  # Feature
│   └── RecurringForm.svelte     # UI
├── settings/
│   ├── CategoriesManager.svelte # Feature
│   ├── TagsManager.svelte       # Feature
│   └── AccountsManager.svelte   # Feature
└── shared/
    ├── MonthPicker.svelte       # UI
    ├── AmountDisplay.svelte     # UI
    ├── ProgressBar.svelte       # UI
    └── EmptyState.svelte        # UI
```

### Charts

Library: Chart.js with svelte-chartjs wrapper.

Key charts:
- **Overview**: Donut (category distribution), Line (6-month trend), Bar (income vs expense)
- **Budgets**: CSS progress bars (no chart library needed)
- **Goals**: CSS progress bars

### State Management

Svelte 5 runes ($state, $derived, $effect) for local state per tab. No global store — each tab loads its own data on activation.

---

## 6. Business Logic Details

### Recurring Transaction Generation

Triggered by `POST /recurring/generate` (called on app open or month navigation):

1. For each active rule where `last_generated < today`
2. Calculate all pending dates based on frequency + day_of_month/day_of_week
3. Create a transaction per date with `is_recurring_instance = 1` and `recurring_rule_id` set
4. Update `last_generated` to the last generated date
5. If `end_date` has passed, set `is_active = 0`

Generated instances are normal transactions — editable and deletable individually. The rule continues generating future instances.

### Remunerated Account Interest

For savings accounts with `interest_rate`:
- `GET /accounts/{id}/balance` calculates estimated annual interest: `balance * (interest_rate / 100)`
- Shown as informational data only
- No automatic interest transactions — user adds manually when bank settles
- This keeps data accurate to real bank statements

### Budget Calculation

`GET /budgets?month=YYYY-MM` returns each budget with:
- `spent`: sum of expense transactions for the month (filtered by category_id if set, all expenses if global)
- `remaining`: amount - spent
- `percentage`: (spent / amount) * 100

Percentage > 100 = over budget (UI shows red).

### Savings Goal Contributions

`POST /goals/{id}/contribute` with `{ amount }`:
- Increments `current_amount`
- Auto-marks `is_completed = 1` when `current_amount >= target_amount`

### Investment P&L

Calculated fields in GET response:
- `total_invested = units * avg_buy_price`
- `current_value = units * current_price`
- `pnl = current_value - total_invested`
- `pnl_percentage = (pnl / total_invested) * 100`

Prices updated manually via PUT (Cortex is offline-first).

### Net Worth

`GET /reports/net-worth`:
- `accounts_total`: sum of all account balances (from transactions)
- `investments_total`: sum of all investment current values
- `net_worth = accounts_total + investments_total`

### Dashboard Widget

Enhanced to show:
- Monthly balance (income - expense)
- Sparkline: last 6 months trend
- Budget progress bar (if global budget exists)

---

## 7. Testing Strategy

### Backend (Go)

One `*_test.go` per module, using SQLite in-memory with migrations applied.

Key test areas:
- Input validation (amounts, enums, foreign keys, date formats)
- Recurring generation (correct instances, no duplicates, end_date respect)
- Budget calculation (spent from real transactions)
- Transfer balance effects (source decreases, destination increases)
- Migration: existing v1 data survives 002_enhanced.sql
- Interest estimation accuracy
- P&L calculation correctness

### Frontend

No frontend tests planned — priority is backend where business logic lives.

---

## 8. Error Handling

### Typed Errors

```go
type AppError struct {
    Code       string
    Message    string
    StatusCode int
}
```

Standard error codes: `NOT_FOUND`, `VALIDATION_ERROR`, `CONFLICT`, `INTERNAL_ERROR`.

### Response Format

Success: `{ "data": T }` (200/201)
Error: `{ "error": { "code": "X", "message": "Y" } }` (4xx/5xx)

Internal SQLite errors never exposed to the client.

---

## 9. i18n

Extend `en.json` and `es.json` with keys for all new modules: accounts, budgets, goals, investments, recurring, tags, reports.
