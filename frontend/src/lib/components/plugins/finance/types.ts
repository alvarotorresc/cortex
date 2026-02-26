// ──────────────────────────────────────────────
// Finance Tracker v2 — TypeScript type definitions
// Mirrors backend Go models in plugins/finance-tracker/backend/
// ──────────────────────────────────────────────

// --- Accounts ---

export type AccountType = 'checking' | 'savings' | 'cash' | 'investment';

export interface Account {
  id: number;
  name: string;
  type: AccountType;
  currency: string;
  interest_rate?: number;
  icon: string;
  color: string;
  is_archived: boolean;
  created_at: string;
}

export interface AccountWithBalance extends Account {
  balance: number;
  estimated_interest?: number;
}

export interface CreateAccountInput {
  name: string;
  type: AccountType;
  currency?: string;
  interest_rate?: number;
  icon: string;
  color: string;
}

export interface UpdateAccountInput {
  name: string;
  type: AccountType;
  currency: string;
  interest_rate?: number;
  icon: string;
  color: string;
}

// --- Transactions ---

export type TransactionType = 'income' | 'expense' | 'transfer';

export interface Transaction {
  id: number;
  amount: number;
  type: TransactionType;
  account_id: number;
  dest_account_id?: number;
  category: string;
  description: string;
  date: string;
  is_recurring_instance: boolean;
  recurring_rule_id?: number;
  tags: Tag[];
  created_at: string;
}

export interface TransactionFilter {
  month?: string;
  account?: string;
  category?: string;
  tag?: string;
  type?: string;
  search?: string;
}

export interface CreateTransactionInput {
  amount: number;
  type: TransactionType;
  account_id?: number;
  dest_account_id?: number;
  category: string;
  description: string;
  date: string;
  tag_ids?: number[];
}

export interface UpdateTransactionInput {
  amount: number;
  type: TransactionType;
  account_id?: number;
  dest_account_id?: number;
  category: string;
  description: string;
  date: string;
  tag_ids?: number[];
}

// --- Categories ---

export type CategoryType = 'income' | 'expense' | 'both';

export interface Category {
  id: number;
  name: string;
  type: CategoryType;
  icon: string;
  color: string;
  is_default: boolean;
  sort_order: number;
}

export interface CreateCategoryInput {
  name: string;
  type: CategoryType;
  icon: string;
  color: string;
}

export interface UpdateCategoryInput {
  name: string;
  type: CategoryType;
  icon: string;
  color: string;
}

export interface ReorderItem {
  id: number;
  sort_order: number;
}

// --- Tags ---

export interface Tag {
  id: number;
  name: string;
  color: string;
}

export interface CreateTagInput {
  name: string;
  color: string;
}

export interface UpdateTagInput {
  name: string;
  color: string;
}

// --- Recurring Rules ---

export type Frequency = 'weekly' | 'biweekly' | 'monthly' | 'yearly';

export interface RecurringRule {
  id: number;
  amount: number;
  type: TransactionType;
  account_id: number;
  dest_account_id?: number;
  category: string;
  description: string;
  frequency: Frequency;
  day_of_month?: number;
  day_of_week?: number;
  month_of_year?: number;
  start_date: string;
  end_date?: string;
  last_generated?: string;
  is_active: boolean;
  created_at: string;
}

export interface CreateRecurringRuleInput {
  amount: number;
  type: TransactionType;
  account_id?: number;
  dest_account_id?: number;
  category: string;
  description: string;
  frequency: Frequency;
  day_of_month?: number;
  day_of_week?: number;
  month_of_year?: number;
  start_date: string;
  end_date?: string;
}

export interface UpdateRecurringRuleInput {
  amount: number;
  type: TransactionType;
  account_id?: number;
  dest_account_id?: number;
  category: string;
  description: string;
  frequency: Frequency;
  day_of_month?: number;
  day_of_week?: number;
  month_of_year?: number;
  start_date: string;
  end_date?: string;
}

export interface GenerateResult {
  generated: number;
}

// --- Budgets ---

export interface Budget {
  id: number;
  name: string;
  category: string;
  amount: number;
  month: string;
  created_at: string;
}

export interface BudgetWithProgress extends Budget {
  spent: number;
  remaining: number;
  percentage: number;
}

export interface CreateBudgetInput {
  name: string;
  category: string;
  amount: number;
  month: string;
}

export interface UpdateBudgetInput {
  name: string;
  category: string;
  amount: number;
  month: string;
}

// --- Savings Goals ---

export interface SavingsGoal {
  id: number;
  name: string;
  target_amount: number;
  current_amount: number;
  target_date?: string;
  icon: string;
  color: string;
  is_completed: boolean;
  created_at: string;
}

export interface CreateGoalInput {
  name: string;
  target_amount: number;
  target_date?: string;
  icon: string;
  color: string;
}

export interface UpdateGoalInput {
  name: string;
  target_amount: number;
  target_date?: string;
  icon: string;
  color: string;
}

export interface ContributeInput {
  amount: number;
}

// --- Investments ---

export type InvestmentType = 'crypto' | 'etf' | 'fund' | 'stock' | 'other';

export interface Investment {
  id: number;
  name: string;
  type: InvestmentType;
  account_id?: number;
  units?: number;
  avg_buy_price?: number;
  current_price?: number;
  currency: string;
  notes: string;
  last_updated?: string;
  created_at: string;
}

export interface InvestmentWithPnL extends Investment {
  total_invested?: number;
  current_value?: number;
  pnl?: number;
  pnl_percentage?: number;
}

export interface CreateInvestmentInput {
  name: string;
  type: InvestmentType;
  account_id?: number;
  units?: number;
  avg_buy_price?: number;
  current_price?: number;
  currency: string;
  notes: string;
  last_updated?: string;
}

export interface UpdateInvestmentInput {
  name: string;
  type: InvestmentType;
  account_id?: number;
  units?: number;
  avg_buy_price?: number;
  current_price?: number;
  currency: string;
  notes: string;
  last_updated?: string;
}

// --- Reports ---

export interface CategoryTotal {
  category: string;
  total: number;
}

export interface AccountTotal {
  account_id: number;
  account_name: string;
  total: number;
}

export interface MonthlySummary {
  month: string;
  income: number;
  expense: number;
  balance: number;
  by_category: CategoryTotal[];
  by_account: AccountTotal[];
}

export interface TrendPoint {
  month: string;
  income: number;
  expense: number;
  balance: number;
}

export interface CategoryComparison {
  category: string;
  current_month: number;
  previous_month: number;
  change: number;
}

export interface NetWorth {
  accounts_total: number;
  investments_total: number;
  net_worth: number;
}

// --- Widget ---

export interface WidgetSparklineEntry {
  month: string;
  balance: number;
}

export interface WidgetBudgetProgress {
  amount: number;
  spent: number;
  remaining: number;
  percentage: number;
}

export interface WidgetData {
  income: number;
  expense: number;
  balance: number;
  month: string;
  sparkline: WidgetSparklineEntry[];
  budget: WidgetBudgetProgress | null;
}

// --- API Response wrappers ---

export interface ApiSuccessResponse<T> {
  data: T;
}

export interface ApiErrorResponse {
  error: {
    code: string;
    message: string;
  };
}
