<script lang="ts">
  import { t } from 'svelte-i18n';
  import Plus from 'lucide-svelte/icons/plus';
  import ArrowLeftRight from 'lucide-svelte/icons/arrow-left-right';
  import Loader2 from 'lucide-svelte/icons/loader-2';
  import { listTransactions, listCategories, listAccounts, listTags } from '../api';
  import type { Transaction, Category, AccountWithBalance, Tag, TransactionFilter } from '../types';
  import EmptyState from '../shared/EmptyState.svelte';
  import TransactionFilters from './TransactionFilters.svelte';
  import TransactionRow from './TransactionRow.svelte';
  import TransactionForm from './TransactionForm.svelte';

  interface Props {
    month: string;
  }

  const { month }: Props = $props();

  // State
  let transactions = $state<Transaction[]>([]);
  let categories = $state<Category[]>([]);
  let accounts = $state<AccountWithBalance[]>([]);
  let tags = $state<Tag[]>([]);
  let loading = $state(true);
  let error = $state('');
  let showForm = $state(false);
  let editingTransaction = $state<Transaction | null>(null);
  let filters = $state<TransactionFilter>({});

  // Reload when month changes
  $effect(() => {
    loadData(month);
  });

  async function loadData(m: string): Promise<void> {
    loading = true;
    error = '';
    try {
      const queryFilters: TransactionFilter = { ...filters, month: m };
      const [txList, catList, acctList, tagList] = await Promise.all([
        listTransactions(queryFilters),
        listCategories(),
        listAccounts(),
        listTags(),
      ]);
      transactions = txList;
      categories = catList;
      accounts = acctList;
      tags = tagList;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load data';
    } finally {
      loading = false;
    }
  }

  async function loadTransactions(): Promise<void> {
    try {
      const queryFilters: TransactionFilter = { ...filters, month };
      transactions = await listTransactions(queryFilters);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load transactions';
    }
  }

  function handleFilter(newFilters: TransactionFilter): void {
    filters = newFilters;
    loadTransactions();
  }

  function handleEdit(tx: Transaction): void {
    editingTransaction = tx;
    showForm = true;
  }

  function handleAdd(): void {
    editingTransaction = null;
    showForm = true;
  }

  function handleSave(): void {
    showForm = false;
    editingTransaction = null;
    loadTransactions();
  }

  function handleDelete(): void {
    showForm = false;
    editingTransaction = null;
    loadTransactions();
  }

  function handleCancel(): void {
    showForm = false;
    editingTransaction = null;
  }
</script>

<div class="space-y-4">
  <!-- Header row with Add button -->
  <div class="flex items-center justify-between">
    <TransactionFilters
      {categories}
      {accounts}
      {tags}
      onfilter={handleFilter}
    />
    <button
      onclick={handleAdd}
      class="ml-3 flex shrink-0 items-center gap-1.5 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-3 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
    >
      <Plus size={16} />
      <span class="hidden sm:inline">{$t('finance.addTransaction')}</span>
    </button>
  </div>

  <!-- Content area -->
  {#if loading}
    <div class="flex items-center justify-center py-16">
      <Loader2 size={24} class="animate-spin text-[var(--color-text-tertiary)]" />
    </div>
  {:else if error}
    <div
      class="flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-6 py-12 text-center"
    >
      <p class="text-sm text-[var(--color-error)]">{error}</p>
      <button
        onclick={() => loadData(month)}
        class="mt-3 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
      >
        {$t('common.retry')}
      </button>
    </div>
  {:else if transactions.length === 0}
    <EmptyState
      icon={ArrowLeftRight}
      message={$t('finance.noTransactions')}
      actionLabel={$t('finance.addTransaction')}
      onaction={handleAdd}
    />
  {:else}
    <div class="flex flex-col gap-2">
      {#each transactions as tx (tx.id)}
        <TransactionRow
          transaction={tx}
          {categories}
          {accounts}
          onedit={handleEdit}
        />
      {/each}
    </div>
  {/if}
</div>

<!-- Form modal -->
{#if showForm}
  <TransactionForm
    transaction={editingTransaction}
    {categories}
    {accounts}
    {tags}
    onsave={handleSave}
    oncancel={handleCancel}
    ondelete={handleDelete}
  />
{/if}
