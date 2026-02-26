// ──────────────────────────────────────────────
// Finance Tracker v2 — Typed API client
// Wraps pluginApi('finance-tracker') with typed methods
// ──────────────────────────────────────────────

import { pluginApi } from '$lib/api';
import type {
  AccountWithBalance,
  ApiSuccessResponse,
  BudgetWithProgress,
  Category,
  CategoryComparison,
  ContributeInput,
  CreateAccountInput,
  CreateBudgetInput,
  CreateCategoryInput,
  CreateGoalInput,
  CreateInvestmentInput,
  CreateRecurringRuleInput,
  CreateTagInput,
  CreateTransactionInput,
  GenerateResult,
  InvestmentWithPnL,
  MonthlySummary,
  NetWorth,
  RecurringRule,
  ReorderItem,
  SavingsGoal,
  Tag,
  Transaction,
  TransactionFilter,
  TrendPoint,
  UpdateAccountInput,
  UpdateBudgetInput,
  UpdateCategoryInput,
  UpdateGoalInput,
  UpdateInvestmentInput,
  UpdateRecurringRuleInput,
  UpdateTagInput,
  UpdateTransactionInput,
  Account,
  Budget,
  Investment,
  WidgetData,
} from './types';

const api = pluginApi('finance-tracker');

/** Extract the `data` field from a standard `{ data: T }` response. */
function extractData<T>(response: ApiSuccessResponse<T>): T {
  return response.data;
}

// ── Accounts ──────────────────────────────────

export async function listAccounts(): Promise<AccountWithBalance[]> {
  const res = await api.fetch<ApiSuccessResponse<AccountWithBalance[]>>('/accounts');
  return extractData(res);
}

export async function createAccount(input: CreateAccountInput): Promise<Account> {
  const res = await api.fetch<ApiSuccessResponse<Account>>('/accounts', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function updateAccount(id: number, input: UpdateAccountInput): Promise<Account> {
  const res = await api.fetch<ApiSuccessResponse<Account>>(`/accounts/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function archiveAccount(id: number): Promise<void> {
  await api.fetch<ApiSuccessResponse<null>>(`/accounts/${id}`, {
    method: 'DELETE',
  });
}

// ── Transactions ──────────────────────────────

export async function listTransactions(
  filters?: TransactionFilter,
): Promise<Transaction[]> {
  const params = new URLSearchParams();
  if (filters) {
    for (const [key, value] of Object.entries(filters)) {
      if (value) params.set(key, value);
    }
  }
  const query = params.toString();
  const path = query ? `/transactions?${query}` : '/transactions';
  const res = await api.fetch<ApiSuccessResponse<Transaction[]>>(path);
  return extractData(res);
}

export async function createTransaction(input: CreateTransactionInput): Promise<Transaction> {
  const res = await api.fetch<ApiSuccessResponse<Transaction>>('/transactions', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function updateTransaction(
  id: number,
  input: UpdateTransactionInput,
): Promise<Transaction> {
  const res = await api.fetch<ApiSuccessResponse<Transaction>>(`/transactions/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function deleteTransaction(id: number): Promise<void> {
  await api.fetch<ApiSuccessResponse<null>>(`/transactions/${id}`, {
    method: 'DELETE',
  });
}

// ── Categories ────────────────────────────────

export async function listCategories(type?: string): Promise<Category[]> {
  const path = type ? `/categories?type=${encodeURIComponent(type)}` : '/categories';
  const res = await api.fetch<ApiSuccessResponse<Category[]>>(path);
  return extractData(res);
}

export async function createCategory(input: CreateCategoryInput): Promise<Category> {
  const res = await api.fetch<ApiSuccessResponse<Category>>('/categories', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function updateCategory(id: number, input: UpdateCategoryInput): Promise<Category> {
  const res = await api.fetch<ApiSuccessResponse<Category>>(`/categories/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function deleteCategory(id: number): Promise<void> {
  await api.fetch<ApiSuccessResponse<null>>(`/categories/${id}`, {
    method: 'DELETE',
  });
}

export async function reorderCategories(items: ReorderItem[]): Promise<void> {
  await api.fetch<ApiSuccessResponse<null>>('/categories/reorder', {
    method: 'PUT',
    body: JSON.stringify(items),
  });
}

// ── Tags ──────────────────────────────────────

export async function listTags(): Promise<Tag[]> {
  const res = await api.fetch<ApiSuccessResponse<Tag[]>>('/tags');
  return extractData(res);
}

export async function createTag(input: CreateTagInput): Promise<Tag> {
  const res = await api.fetch<ApiSuccessResponse<Tag>>('/tags', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function updateTag(id: number, input: UpdateTagInput): Promise<Tag> {
  const res = await api.fetch<ApiSuccessResponse<Tag>>(`/tags/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function deleteTag(id: number): Promise<void> {
  await api.fetch<ApiSuccessResponse<null>>(`/tags/${id}`, {
    method: 'DELETE',
  });
}

// ── Recurring Rules ───────────────────────────

export async function listRecurringRules(): Promise<RecurringRule[]> {
  const res = await api.fetch<ApiSuccessResponse<RecurringRule[]>>('/recurring');
  return extractData(res);
}

export async function createRecurringRule(
  input: CreateRecurringRuleInput,
): Promise<RecurringRule> {
  const res = await api.fetch<ApiSuccessResponse<RecurringRule>>('/recurring', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function updateRecurringRule(
  id: number,
  input: UpdateRecurringRuleInput,
): Promise<RecurringRule> {
  const res = await api.fetch<ApiSuccessResponse<RecurringRule>>(`/recurring/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function deleteRecurringRule(id: number): Promise<void> {
  await api.fetch<ApiSuccessResponse<null>>(`/recurring/${id}`, {
    method: 'DELETE',
  });
}

export async function generateRecurring(): Promise<GenerateResult> {
  const res = await api.fetch<ApiSuccessResponse<GenerateResult>>('/recurring/generate', {
    method: 'POST',
  });
  return extractData(res);
}

// ── Budgets ───────────────────────────────────

export async function listBudgets(month: string): Promise<BudgetWithProgress[]> {
  const res = await api.fetch<ApiSuccessResponse<BudgetWithProgress[]>>(
    `/budgets?month=${encodeURIComponent(month)}`,
  );
  return extractData(res);
}

export async function createBudget(input: CreateBudgetInput): Promise<Budget> {
  const res = await api.fetch<ApiSuccessResponse<Budget>>('/budgets', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function updateBudget(id: number, input: UpdateBudgetInput): Promise<Budget> {
  const res = await api.fetch<ApiSuccessResponse<Budget>>(`/budgets/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function deleteBudget(id: number): Promise<void> {
  await api.fetch<ApiSuccessResponse<null>>(`/budgets/${id}`, {
    method: 'DELETE',
  });
}

// ── Savings Goals ─────────────────────────────

export async function listGoals(): Promise<SavingsGoal[]> {
  const res = await api.fetch<ApiSuccessResponse<SavingsGoal[]>>('/goals');
  return extractData(res);
}

export async function createGoal(input: CreateGoalInput): Promise<SavingsGoal> {
  const res = await api.fetch<ApiSuccessResponse<SavingsGoal>>('/goals', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function updateGoal(id: number, input: UpdateGoalInput): Promise<SavingsGoal> {
  const res = await api.fetch<ApiSuccessResponse<SavingsGoal>>(`/goals/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function deleteGoal(id: number): Promise<void> {
  await api.fetch<ApiSuccessResponse<null>>(`/goals/${id}`, {
    method: 'DELETE',
  });
}

export async function contributeToGoal(
  id: number,
  amount: number,
): Promise<SavingsGoal> {
  const input: ContributeInput = { amount };
  const res = await api.fetch<ApiSuccessResponse<SavingsGoal>>(`/goals/${id}/contribute`, {
    method: 'POST',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

// ── Investments ───────────────────────────────

export async function listInvestments(): Promise<InvestmentWithPnL[]> {
  const res = await api.fetch<ApiSuccessResponse<InvestmentWithPnL[]>>('/investments');
  return extractData(res);
}

export async function createInvestment(input: CreateInvestmentInput): Promise<Investment> {
  const res = await api.fetch<ApiSuccessResponse<Investment>>('/investments', {
    method: 'POST',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function updateInvestment(
  id: number,
  input: UpdateInvestmentInput,
): Promise<Investment> {
  const res = await api.fetch<ApiSuccessResponse<Investment>>(`/investments/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
  return extractData(res);
}

export async function deleteInvestment(id: number): Promise<void> {
  await api.fetch<ApiSuccessResponse<null>>(`/investments/${id}`, {
    method: 'DELETE',
  });
}

// ── Reports ───────────────────────────────────

export async function getSummary(month: string): Promise<MonthlySummary> {
  const res = await api.fetch<ApiSuccessResponse<MonthlySummary>>(
    `/reports/summary?month=${encodeURIComponent(month)}`,
  );
  return extractData(res);
}

export async function getTrends(from: string, to: string): Promise<TrendPoint[]> {
  const res = await api.fetch<ApiSuccessResponse<TrendPoint[]>>(
    `/reports/trends?from=${encodeURIComponent(from)}&to=${encodeURIComponent(to)}`,
  );
  return extractData(res);
}

export async function getCategoryComparison(month: string): Promise<CategoryComparison[]> {
  const res = await api.fetch<ApiSuccessResponse<CategoryComparison[]>>(
    `/reports/categories?month=${encodeURIComponent(month)}`,
  );
  return extractData(res);
}

export async function getNetWorth(): Promise<NetWorth> {
  const res = await api.fetch<ApiSuccessResponse<NetWorth>>('/reports/net-worth');
  return extractData(res);
}

// ── Widget ────────────────────────────────────

export async function getWidgetData(): Promise<WidgetData> {
  const res = await api.widget<ApiSuccessResponse<WidgetData>>('dashboard-widget');
  return extractData(res);
}
