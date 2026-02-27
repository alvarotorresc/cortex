<script lang="ts">
  import { t } from 'svelte-i18n';
  import Plus from 'lucide-svelte/icons/plus';
  import PiggyBank from 'lucide-svelte/icons/piggy-bank';
  import Loader2 from 'lucide-svelte/icons/loader-2';
  import { listBudgets, listCategories } from '../api';
  import type { BudgetWithProgress, Category } from '../types';
  import EmptyState from '../shared/EmptyState.svelte';
  import BudgetCard from './BudgetCard.svelte';
  import BudgetForm from './BudgetForm.svelte';

  interface Props {
    month: string;
  }

  const { month }: Props = $props();

  // State
  let budgets = $state<BudgetWithProgress[]>([]);
  let categories = $state<Category[]>([]);
  let loading = $state(true);
  let error = $state('');
  let showForm = $state(false);
  let editingBudget = $state<BudgetWithProgress | null>(null);

  // Reload when month changes
  $effect(() => {
    loadData(month);
  });

  async function loadData(m: string): Promise<void> {
    loading = true;
    error = '';
    try {
      const [budgetList, catList] = await Promise.all([
        listBudgets(m),
        listCategories(),
      ]);
      budgets = budgetList;
      categories = catList;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load data';
    } finally {
      loading = false;
    }
  }

  async function loadBudgets(): Promise<void> {
    try {
      budgets = await listBudgets(month);
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load budgets';
    }
  }

  function handleEdit(budget: BudgetWithProgress): void {
    editingBudget = budget;
    showForm = true;
  }

  function handleAdd(): void {
    editingBudget = null;
    showForm = true;
  }

  function handleSave(): void {
    showForm = false;
    editingBudget = null;
    loadBudgets();
  }

  function handleDelete(): void {
    showForm = false;
    editingBudget = null;
    loadBudgets();
  }

  function handleCancel(): void {
    showForm = false;
    editingBudget = null;
  }
</script>

<div class="space-y-4">
  <!-- Header row with Add button -->
  <div class="flex items-center justify-end">
    <button
      onclick={handleAdd}
      class="flex shrink-0 items-center gap-1.5 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-3 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
    >
      <Plus size={16} />
      <span class="hidden sm:inline">{$t('finance.budgets.addBudget')}</span>
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
  {:else if budgets.length === 0}
    <EmptyState
      icon={PiggyBank}
      message={$t('finance.budgets.noBudgets')}
      actionLabel={$t('finance.budgets.addBudget')}
      onaction={handleAdd}
    />
  {:else}
    <div class="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
      {#each budgets as budget (budget.id)}
        <BudgetCard {budget} onedit={handleEdit} />
      {/each}
    </div>
  {/if}
</div>

<!-- Form modal -->
{#if showForm}
  <BudgetForm
    budget={editingBudget}
    {categories}
    {month}
    onsave={handleSave}
    oncancel={handleCancel}
    ondelete={handleDelete}
  />
{/if}
