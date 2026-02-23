-- Finance Tracker: initial schema
-- Creates transactions and categories tables with default category data.

CREATE TABLE IF NOT EXISTS transactions (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    amount REAL NOT NULL,
    type TEXT NOT NULL CHECK(type IN ('income', 'expense')),
    category TEXT NOT NULL,
    description TEXT,
    date TEXT NOT NULL,
    created_at TEXT NOT NULL DEFAULT (datetime('now'))
);

CREATE INDEX IF NOT EXISTS idx_transactions_date ON transactions(date);
CREATE INDEX IF NOT EXISTS idx_transactions_type ON transactions(type);

CREATE TABLE IF NOT EXISTS categories (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    name TEXT NOT NULL UNIQUE,
    icon TEXT,
    is_default INTEGER NOT NULL DEFAULT 0
);

-- Default categories
INSERT OR IGNORE INTO categories (name, icon, is_default) VALUES
    ('salary', 'banknote', 1),
    ('groceries', 'shopping-cart', 1),
    ('transport', 'car', 1),
    ('entertainment', 'gamepad-2', 1),
    ('restaurants', 'utensils', 1),
    ('bills', 'receipt', 1),
    ('health', 'heart-pulse', 1),
    ('other', 'circle-dot', 1);
