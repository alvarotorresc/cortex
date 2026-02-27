<script lang="ts">
  import { t } from 'svelte-i18n';
  import type { BudgetWithProgress } from '../types';
  import ProgressBar from '../shared/ProgressBar.svelte';
  import AmountDisplay from '../shared/AmountDisplay.svelte';

  interface Props {
    budget: BudgetWithProgress;
    onedit: (budget: BudgetWithProgress) => void;
  }

  const { budget, onedit }: Props = $props();

  const isOverBudget = $derived(budget.percentage > 100);
</script>

<button
  type="button"
  onclick={() => onedit(budget)}
  class="flex w-full flex-col gap-3 rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-4 text-left transition-colors hover:bg-[var(--color-bg-tertiary)]"
  aria-label="{budget.name} — {Math.round(budget.percentage)}%"
>
  <!-- Header: name + category badge -->
  <div class="flex items-start justify-between gap-2">
    <div class="min-w-0">
      <h4 class="truncate text-sm font-semibold text-[var(--color-text-primary)]">
        {budget.name}
      </h4>
      {#if budget.category}
        <span class="text-xs text-[var(--color-text-tertiary)]">{budget.category}</span>
      {:else}
        <span class="text-xs text-[var(--color-text-tertiary)]">{$t('finance.budgets.global')}</span>
      {/if}
    </div>
  </div>

  <!-- Amounts: spent / amount -->
  <div class="flex items-baseline gap-1 text-sm">
    <AmountDisplay amount={-budget.spent} />
    <span class="text-[var(--color-text-tertiary)]">/</span>
    <span class="text-[var(--color-text-secondary)]">
      <AmountDisplay amount={budget.amount} />
    </span>
  </div>

  <!-- Progress bar -->
  <ProgressBar percentage={budget.percentage} />

  <!-- Footer: remaining or over budget warning -->
  <div class="flex items-center justify-between text-xs">
    {#if isOverBudget}
      <span class="font-medium text-[var(--color-error)]">
        {$t('finance.budgets.overBudget')}
      </span>
    {:else}
      <span class="text-[var(--color-text-tertiary)]">
        {$t('finance.budgets.remaining')}
      </span>
    {/if}
    <span class={isOverBudget ? 'font-medium text-[var(--color-error)]' : 'text-[var(--color-text-secondary)]'}>
      <AmountDisplay amount={budget.remaining} />
    </span>
  </div>
</button>
