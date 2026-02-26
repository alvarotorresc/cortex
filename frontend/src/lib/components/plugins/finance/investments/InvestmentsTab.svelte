<script lang="ts">
  import { t } from 'svelte-i18n';
  import Plus from 'lucide-svelte/icons/plus';
  import TrendingUp from 'lucide-svelte/icons/trending-up';
  import Loader2 from 'lucide-svelte/icons/loader-2';
  import { listInvestments } from '../api';
  import type { InvestmentWithPnL } from '../types';
  import EmptyState from '../shared/EmptyState.svelte';
  import AmountDisplay from '../shared/AmountDisplay.svelte';
  import InvestmentCard from './InvestmentCard.svelte';
  import InvestmentForm from './InvestmentForm.svelte';

  // State
  let investments = $state<InvestmentWithPnL[]>([]);
  let loading = $state(true);
  let error = $state('');
  let showForm = $state(false);
  let editingInvestment = $state<InvestmentWithPnL | null>(null);

  // Derived totals
  const totalValue = $derived(
    investments.reduce((sum, inv) => sum + (inv.current_value ?? 0), 0),
  );
  const totalPnL = $derived(
    investments.reduce((sum, inv) => sum + (inv.pnl ?? 0), 0),
  );
  const totalInvested = $derived(
    investments.reduce((sum, inv) => sum + (inv.total_invested ?? 0), 0),
  );
  const totalPnLPercentage = $derived(
    totalInvested > 0 ? (totalPnL / totalInvested) * 100 : 0,
  );
  const pnlPositive = $derived(totalPnL >= 0);

  // Load on mount
  $effect(() => {
    loadInvestments();
  });

  async function loadInvestments(): Promise<void> {
    loading = true;
    error = '';
    try {
      investments = await listInvestments();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load investments';
    } finally {
      loading = false;
    }
  }

  function handleEdit(inv: InvestmentWithPnL): void {
    editingInvestment = inv;
    showForm = true;
  }

  function handleAdd(): void {
    editingInvestment = null;
    showForm = true;
  }

  function handleSave(): void {
    showForm = false;
    editingInvestment = null;
    loadInvestments();
  }

  function handleDelete(): void {
    showForm = false;
    editingInvestment = null;
    loadInvestments();
  }

  function handleCancel(): void {
    showForm = false;
    editingInvestment = null;
  }
</script>

<div class="space-y-4">
  <!-- Header row with Add button -->
  <div class="flex items-center justify-between">
    <h2 class="text-sm font-semibold text-[var(--color-text-secondary)]">
      {$t('finance.investments.portfolio')}
    </h2>
    <button
      onclick={handleAdd}
      class="flex shrink-0 items-center gap-1.5 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-3 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
    >
      <Plus size={16} />
      <span class="hidden sm:inline">{$t('finance.investments.addInvestment')}</span>
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
        onclick={loadInvestments}
        class="mt-3 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
      >
        {$t('common.retry')}
      </button>
    </div>
  {:else if investments.length === 0}
    <EmptyState
      icon={TrendingUp}
      message={$t('finance.investments.noInvestments')}
      actionLabel={$t('finance.investments.addInvestment')}
      onaction={handleAdd}
    />
  {:else}
    <!-- Portfolio summary -->
    <div class="grid grid-cols-2 gap-3">
      <div class="rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-4">
        <p class="text-xs font-medium text-[var(--color-text-tertiary)]">
          {$t('finance.investments.totalValue')}
        </p>
        <p class="mt-1 text-lg font-bold text-[var(--color-text-primary)]">
          <AmountDisplay amount={totalValue} />
        </p>
      </div>
      <div class="rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-4">
        <p class="text-xs font-medium text-[var(--color-text-tertiary)]">
          {$t('finance.investments.totalPnL')}
        </p>
        <div class="mt-1 flex items-center gap-2">
          <span class="text-lg font-bold">
            <AmountDisplay amount={totalPnL} showSign />
          </span>
          {#if totalInvested > 0}
            <span
              class="rounded-full px-2 py-0.5 text-[11px] font-semibold {pnlPositive
                ? 'text-[var(--color-success)]'
                : 'text-[var(--color-error)]'}"
              style:background-color={pnlPositive
                ? 'color-mix(in srgb, var(--color-success) 10%, transparent)'
                : 'color-mix(in srgb, var(--color-error) 10%, transparent)'}
            >
              {pnlPositive ? '+' : ''}{totalPnLPercentage.toFixed(2)}%
            </span>
          {/if}
        </div>
      </div>
    </div>

    <!-- Investment cards grid -->
    <div class="grid gap-3 sm:grid-cols-2">
      {#each investments as inv (inv.id)}
        <InvestmentCard investment={inv} onedit={handleEdit} />
      {/each}
    </div>
  {/if}
</div>

<!-- Form modal -->
{#if showForm}
  <InvestmentForm
    investment={editingInvestment}
    onsave={handleSave}
    oncancel={handleCancel}
    ondelete={handleDelete}
  />
{/if}
