-- Finance Tracker v2: enhanced schema
-- Adds accounts, tags, recurring rules, budgets, savings goals, investments.
-- Extends transactions and categories with new columns.

-- Accounts table
CREATE TABLE IF NOT EXISTS accounts (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('checking', 'savings', 'cash', 'investment')),
    currency TEXT NOT NULL DEFAULT 'EUR',
    initial_balance REAL NOT NULL DEFAULT 0,
    is_active INTEGER NOT NULL DEFAULT 1,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Default account
INSERT OR IGNORE INTO accounts (id, name, type, currency) VALUES (1, 'Main Account', 'checking', 'EUR');

-- Tags table
CREATE TABLE IF NOT EXISTS tags (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    color TEXT
);

-- Transaction-tags join table
CREATE TABLE IF NOT EXISTS transaction_tags (
    transaction_id INTEGER NOT NULL REFERENCES transactions(id) ON DELETE CASCADE,
    tag_id INTEGER NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    PRIMARY KEY (transaction_id, tag_id)
);

-- Recurring rules table
CREATE TABLE IF NOT EXISTS recurring_rules (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount REAL NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('income', 'expense', 'transfer')),
    category TEXT NOT NULL,
    description TEXT,
    account_id INTEGER NOT NULL DEFAULT 1 REFERENCES accounts(id),
    frequency TEXT NOT NULL CHECK(frequency IN ('weekly', 'biweekly', 'monthly', 'yearly')),
    start_date TEXT NOT NULL,
    end_date TEXT,
    next_due TEXT NOT NULL,
    is_active INTEGER NOT NULL DEFAULT 1,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Budgets table
CREATE TABLE IF NOT EXISTS budgets (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    category TEXT NOT NULL,
    amount REAL NOT NULL,
    month TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(category, month)
);

-- Savings goals table
CREATE TABLE IF NOT EXISTS savings_goals (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    target_amount REAL NOT NULL,
    current_amount REAL NOT NULL DEFAULT 0,
    deadline TEXT,
    color TEXT,
    is_completed INTEGER NOT NULL DEFAULT 0,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- Investments table
CREATE TABLE IF NOT EXISTS investments (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('crypto', 'etf', 'fund', 'stock', 'other')),
    symbol TEXT,
    units REAL NOT NULL DEFAULT 0,
    avg_buy_price REAL NOT NULL DEFAULT 0,
    current_price REAL NOT NULL DEFAULT 0,
    currency TEXT NOT NULL DEFAULT 'EUR',
    notes TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now'))
);

-- ALTER transactions: add new columns
-- Each ALTER TABLE must be a separate statement for SQLite.
-- NOTE: The v1 transactions table has CHECK(type IN ('income', 'expense')).
-- SQLite cannot ALTER CHECK constraints, so 'transfer' type validation
-- is enforced at the application/handler level, not in SQL.
-- NOTE: The v1 transactions table uses a TEXT 'category' column.
-- We intentionally keep it as TEXT rather than adding a category_id FK
-- to categories. The text category is sufficient and simpler for v2.
ALTER TABLE transactions ADD COLUMN account_id INTEGER NOT NULL DEFAULT 1;
ALTER TABLE transactions ADD COLUMN dest_account_id INTEGER;
ALTER TABLE transactions ADD COLUMN is_recurring_instance INTEGER NOT NULL DEFAULT 0;
ALTER TABLE transactions ADD COLUMN recurring_rule_id INTEGER;

-- ALTER categories: add new columns
ALTER TABLE categories ADD COLUMN type TEXT NOT NULL DEFAULT 'both';
ALTER TABLE categories ADD COLUMN color TEXT;
ALTER TABLE categories ADD COLUMN sort_order INTEGER NOT NULL DEFAULT 0;

-- New indexes for v2
CREATE INDEX IF NOT EXISTS idx_transactions_account ON transactions(account_id);
CREATE INDEX IF NOT EXISTS idx_transactions_category ON transactions(category);
CREATE INDEX IF NOT EXISTS idx_recurring_rules_active ON recurring_rules(is_active);
CREATE INDEX IF NOT EXISTS idx_budgets_month ON budgets(month);
CREATE INDEX IF NOT EXISTS idx_investments_type ON investments(type);
