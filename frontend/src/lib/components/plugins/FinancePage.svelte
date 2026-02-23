<script lang="ts">
  import { t } from 'svelte-i18n';
  import { onMount } from 'svelte';
  import Plus from 'lucide-svelte/icons/plus';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import ChevronLeft from 'lucide-svelte/icons/chevron-left';
  import ChevronRight from 'lucide-svelte/icons/chevron-right';
  import TrendingUp from 'lucide-svelte/icons/trending-up';
  import TrendingDown from 'lucide-svelte/icons/trending-down';
  import Wallet from 'lucide-svelte/icons/wallet';
  import { pluginApi } from '$lib/api';

  interface Transaction {
    id: string;
    amount: number;
    type: 'income' | 'expense';
    category: string;
    description: string;
    date: string;
    created_at: string;
  }

  interface MonthlySummary {
    income: number;
    expense: number;
    balance: number;
    currency: string;
    by_category: Record<string, number>;
  }

  const api = pluginApi('finance-tracker');

  let transactions = $state<Transaction[]>([]);
  let summary = $state<MonthlySummary | null>(null);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let submitting = $state(false);

  // Month navigation
  let currentDate = $state(new Date());
  const currentMonth = $derived(
    `${currentDate.getFullYear()}-${String(currentDate.getMonth() + 1).padStart(2, '0')}`,
  );
  const monthLabel = $derived(
    currentDate.toLocaleDateString(undefined, { year: 'numeric', month: 'long' }),
  );

  // Form state
  let formAmount = $state('');
  let formType = $state<'income' | 'expense'>('expense');
  let formCategory = $state('');
  let formDescription = $state('');
  let formDate = $state(new Date().toISOString().slice(0, 10));
  let showForm = $state(false);

  const categories = [
    'food',
    'transport',
    'housing',
    'entertainment',
    'health',
    'education',
    'shopping',
    'salary',
    'freelance',
    'other',
  ];

  async function loadData() {
    loading = true;
    error = null;

    try {
      const [txRes, sumRes] = await Promise.all([
        api.fetch<{ data: Transaction[] }>(`/transactions?month=${currentMonth}`),
        api.fetch<{ data: MonthlySummary }>(`/summary?month=${currentMonth}`),
      ]);
      transactions = txRes.data ?? [];
      summary = sumRes.data ?? null;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load data';
    } finally {
      loading = false;
    }
  }

  async function addTransaction() {
    if (!formAmount || !formCategory) return;

    submitting = true;
    try {
      await api.fetch('/transactions', {
        method: 'POST',
        body: JSON.stringify({
          amount: parseFloat(formAmount),
          type: formType,
          category: formCategory,
          description: formDescription,
          date: formDate,
        }),
      });

      // Reset form
      formAmount = '';
      formType = 'expense';
      formCategory = '';
      formDescription = '';
      formDate = new Date().toISOString().slice(0, 10);
      showForm = false;

      await loadData();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to add transaction';
    } finally {
      submitting = false;
    }
  }

  async function deleteTransaction(id: string) {
    try {
      await api.fetch(`/transactions/${id}`, { method: 'DELETE' });
      await loadData();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete transaction';
    }
  }

  function prevMonth() {
    const d = new Date(currentDate);
    d.setMonth(d.getMonth() - 1);
    currentDate = d;
  }

  function nextMonth() {
    const d = new Date(currentDate);
    d.setMonth(d.getMonth() + 1);
    currentDate = d;
  }

  function formatCurrency(amount: number, currency: string = 'EUR'): string {
    return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(amount);
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString(undefined, {
      day: 'numeric',
      month: 'short',
    });
  }

  $effect(() => {
    currentMonth;
    loadData();
  });

  onMount(() => {
    loadData();
  });
</script>

<div class="space-y-6">
  <!-- Header with month navigation -->
  <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
    <div class="flex items-center gap-3">
      <Wallet size={20} class="text-[var(--color-plugin-finance)]" />
      <h2 class="text-2xl font-semibold text-[var(--color-text-primary)]">
        {$t('finance.title')}
      </h2>
    </div>

    <div class="flex items-center gap-2">
      <button
        onclick={prevMonth}
        class="rounded-[var(--radius-md)] p-2 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
        aria-label="Previous month"
      >
        <ChevronLeft size={20} />
      </button>
      <span class="min-w-[140px] text-center text-sm font-medium text-[var(--color-text-primary)]">
        {monthLabel}
      </span>
      <button
        onclick={nextMonth}
        class="rounded-[var(--radius-md)] p-2 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
        aria-label="Next month"
      >
        <ChevronRight size={20} />
      </button>
    </div>
  </div>

  {#if loading}
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      {#each [1, 2, 3] as _}
        <div
          class="h-24 animate-pulse rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]"
        ></div>
      {/each}
    </div>
  {:else if error}
    <div
      class="rounded-[var(--radius-lg)] border border-[var(--color-error)]/20 bg-[var(--color-error)]/5 p-4"
    >
      <p class="text-sm text-[var(--color-error)]">{error}</p>
      <button
        onclick={loadData}
        class="mt-2 text-sm font-medium text-[var(--color-brand-blue)] hover:underline"
      >
        {$t('common.retry')}
      </button>
    </div>
  {:else}
    <div class="grid grid-cols-1 gap-6 lg:grid-cols-3">
      <!-- Main content: transactions list -->
      <div class="space-y-4 lg:col-span-2">
        <!-- Add transaction button -->
        <button
          onclick={() => (showForm = !showForm)}
          class="flex w-full items-center justify-center gap-2 rounded-[var(--radius-md)] border border-dashed border-[var(--color-border)] px-4 py-3 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:border-[var(--color-brand-blue)] hover:text-[var(--color-brand-blue)]"
        >
          <Plus size={16} />
          {$t('finance.addTransaction')}
        </button>

        <!-- Add transaction form -->
        {#if showForm}
          <form
            onsubmit={(e) => {
              e.preventDefault();
              addTransaction();
            }}
            class="space-y-4 rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-6"
          >
            <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
              <div>
                <label
                  for="amount"
                  class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]"
                >
                  {$t('finance.amount')}
                </label>
                <input
                  id="amount"
                  type="number"
                  step="0.01"
                  min="0"
                  bind:value={formAmount}
                  required
                  class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
                  placeholder="0.00"
                />
              </div>

              <fieldset>
                <legend class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]">
                  {$t('finance.type')}
                </legend>
                <div class="flex gap-4 pt-1">
                  <label class="flex items-center gap-2 text-sm text-[var(--color-text-primary)]">
                    <input
                      type="radio"
                      name="type"
                      value="expense"
                      bind:group={formType}
                      class="accent-[var(--color-error)]"
                    />
                    {$t('finance.expense')}
                  </label>
                  <label class="flex items-center gap-2 text-sm text-[var(--color-text-primary)]">
                    <input
                      type="radio"
                      name="type"
                      value="income"
                      bind:group={formType}
                      class="accent-[var(--color-success)]"
                    />
                    {$t('finance.income')}
                  </label>
                </div>
              </fieldset>

              <div>
                <label
                  for="category"
                  class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]"
                >
                  {$t('finance.category')}
                </label>
                <select
                  id="category"
                  bind:value={formCategory}
                  required
                  class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
                >
                  <option value="" disabled>{$t('finance.category')}</option>
                  {#each categories as cat}
                    <option value={cat}>{cat}</option>
                  {/each}
                </select>
              </div>

              <div>
                <label
                  for="date"
                  class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]"
                >
                  {$t('finance.date')}
                </label>
                <input
                  id="date"
                  type="date"
                  bind:value={formDate}
                  required
                  class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
                />
              </div>

              <div class="sm:col-span-2">
                <label
                  for="description"
                  class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]"
                >
                  {$t('finance.description')}
                </label>
                <input
                  id="description"
                  type="text"
                  bind:value={formDescription}
                  class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
                  placeholder={$t('finance.description')}
                />
              </div>
            </div>

            <div class="flex justify-end gap-2">
              <button
                type="button"
                onclick={() => (showForm = false)}
                class="rounded-[var(--radius-md)] border border-[var(--color-border)] px-4 py-2 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
              >
                {$t('dashboard.cancel')}
              </button>
              <button
                type="submit"
                disabled={submitting}
                class="rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90 disabled:opacity-50"
              >
                {$t('finance.save')}
              </button>
            </div>
          </form>
        {/if}

        <!-- Transactions list -->
        {#if transactions.length === 0}
          <div
            class="rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-6 py-12 text-center"
          >
            <p class="text-sm text-[var(--color-text-tertiary)]">
              {$t('finance.noTransactions')}
            </p>
          </div>
        {:else}
          <div
            class="divide-y divide-[var(--color-border)] rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]"
          >
            {#each transactions as tx}
              <div
                class="flex items-center justify-between px-6 py-4 transition-colors hover:bg-[var(--color-bg-tertiary)]"
              >
                <div class="flex items-center gap-4">
                  <div
                    class="flex h-8 w-8 items-center justify-center rounded-[var(--radius-full)] {tx.type ===
                    'income'
                      ? 'bg-[var(--color-success)]/10 text-[var(--color-success)]'
                      : 'bg-[var(--color-error)]/10 text-[var(--color-error)]'}"
                  >
                    {#if tx.type === 'income'}
                      <TrendingUp size={16} />
                    {:else}
                      <TrendingDown size={16} />
                    {/if}
                  </div>
                  <div>
                    <p class="text-sm font-medium text-[var(--color-text-primary)]">
                      {tx.description || tx.category}
                    </p>
                    <p class="text-xs text-[var(--color-text-tertiary)]">
                      {formatDate(tx.date)} &middot;
                      <span class="capitalize">{tx.category}</span>
                    </p>
                  </div>
                </div>

                <div class="flex items-center gap-3">
                  <span
                    class="text-sm font-semibold {tx.type === 'income'
                      ? 'text-[var(--color-success)]'
                      : 'text-[var(--color-error)]'}"
                  >
                    {tx.type === 'income' ? '+' : '-'}{formatCurrency(
                      tx.amount,
                      summary?.currency,
                    )}
                  </span>
                  <button
                    onclick={() => deleteTransaction(tx.id)}
                    class="rounded-[var(--radius-sm)] p-1.5 text-[var(--color-text-tertiary)] transition-colors hover:bg-[var(--color-error)]/10 hover:text-[var(--color-error)]"
                    aria-label={$t('finance.delete')}
                  >
                    <Trash2 size={14} />
                  </button>
                </div>
              </div>
            {/each}
          </div>
        {/if}
      </div>

      <!-- Sidebar: monthly summary -->
      <div class="space-y-4">
        {#if summary}
          <!-- Balance card -->
          <div
            class="rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-6"
          >
            <h3 class="mb-4 text-sm font-semibold text-[var(--color-text-secondary)]">
              {$t('finance.summary')}
            </h3>

            <div class="space-y-3">
              <div class="flex items-center justify-between">
                <span class="text-sm text-[var(--color-text-secondary)]">
                  {$t('finance.income')}
                </span>
                <span class="text-sm font-semibold text-[var(--color-success)]">
                  +{formatCurrency(summary.income, summary.currency)}
                </span>
              </div>
              <div class="flex items-center justify-between">
                <span class="text-sm text-[var(--color-text-secondary)]">
                  {$t('finance.expense')}
                </span>
                <span class="text-sm font-semibold text-[var(--color-error)]">
                  -{formatCurrency(summary.expense, summary.currency)}
                </span>
              </div>
              <div
                class="border-t border-[var(--color-border)] pt-3"
              >
                <div class="flex items-center justify-between">
                  <span class="text-sm font-medium text-[var(--color-text-primary)]">
                    {$t('finance.balance')}
                  </span>
                  <span
                    class="text-lg font-bold {summary.balance >= 0
                      ? 'text-[var(--color-success)]'
                      : 'text-[var(--color-error)]'}"
                  >
                    {formatCurrency(summary.balance, summary.currency)}
                  </span>
                </div>
              </div>
            </div>
          </div>

          <!-- Categories breakdown -->
          {#if summary.by_category && Object.keys(summary.by_category).length > 0}
            <div
              class="rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-6"
            >
              <h3 class="mb-4 text-sm font-semibold text-[var(--color-text-secondary)]">
                {$t('finance.category')}
              </h3>
              <div class="space-y-2">
                {#each Object.entries(summary.by_category).sort((a, b) => b[1] - a[1]) as [cat, amount]}
                  <div class="flex items-center justify-between">
                    <span class="text-sm capitalize text-[var(--color-text-secondary)]">
                      {cat}
                    </span>
                    <span class="text-sm font-medium text-[var(--color-text-primary)]">
                      {formatCurrency(amount, summary.currency)}
                    </span>
                  </div>
                {/each}
              </div>
            </div>
          {/if}
        {/if}
      </div>
    </div>
  {/if}
</div>
