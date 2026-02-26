<script lang="ts">
  import { t } from 'svelte-i18n';
  import TrendingUp from 'lucide-svelte/icons/trending-up';
  import TrendingDown from 'lucide-svelte/icons/trending-down';
  import Scale from 'lucide-svelte/icons/scale';
  import type { MonthlySummary } from '../types';
  import AmountDisplay from '../shared/AmountDisplay.svelte';

  interface Props {
    summary: MonthlySummary;
  }

  const { summary }: Props = $props();
</script>

<div
  class="rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5"
>
  <h3 class="mb-4 text-sm font-medium text-[var(--color-text-secondary)]">
    {$t('finance.overview.monthlyBalance')}
  </h3>

  <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
    <!-- Income -->
    <div
      class="flex items-center gap-3 rounded-[var(--radius-md)] bg-[var(--color-bg-tertiary)] p-4"
    >
      <div
        class="flex h-10 w-10 shrink-0 items-center justify-center rounded-[var(--radius-md)] bg-[var(--color-success)]/10 text-[var(--color-success)]"
      >
        <TrendingUp size={20} />
      </div>
      <div class="min-w-0">
        <p class="text-xs text-[var(--color-text-tertiary)]">
          {$t('finance.income')}
        </p>
        <p class="text-lg font-semibold text-[var(--color-success)]">
          <AmountDisplay amount={summary.income} />
        </p>
      </div>
    </div>

    <!-- Expenses -->
    <div
      class="flex items-center gap-3 rounded-[var(--radius-md)] bg-[var(--color-bg-tertiary)] p-4"
    >
      <div
        class="flex h-10 w-10 shrink-0 items-center justify-center rounded-[var(--radius-md)] bg-[var(--color-error)]/10 text-[var(--color-error)]"
      >
        <TrendingDown size={20} />
      </div>
      <div class="min-w-0">
        <p class="text-xs text-[var(--color-text-tertiary)]">
          {$t('finance.expense')}
        </p>
        <p class="text-lg font-semibold text-[var(--color-error)]">
          <AmountDisplay amount={-summary.expense} />
        </p>
      </div>
    </div>

    <!-- Balance -->
    <div
      class="flex items-center gap-3 rounded-[var(--radius-md)] bg-[var(--color-bg-tertiary)] p-4"
    >
      <div
        class="flex h-10 w-10 shrink-0 items-center justify-center rounded-[var(--radius-md)] bg-[var(--color-brand-blue)]/10 text-[var(--color-brand-blue)]"
      >
        <Scale size={20} />
      </div>
      <div class="min-w-0">
        <p class="text-xs text-[var(--color-text-tertiary)]">
          {$t('finance.balance')}
        </p>
        <p class="text-lg font-semibold">
          <AmountDisplay amount={summary.balance} showSign />
        </p>
      </div>
    </div>
  </div>
</div>
