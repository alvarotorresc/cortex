<script lang="ts">
  import { t } from 'svelte-i18n';
  import Gem from 'lucide-svelte/icons/gem';
  import Landmark from 'lucide-svelte/icons/landmark';
  import TrendingUp from 'lucide-svelte/icons/trending-up';
  import type { NetWorth } from '../types';
  import AmountDisplay from '../shared/AmountDisplay.svelte';

  interface Props {
    netWorth: NetWorth;
  }

  const { netWorth }: Props = $props();
</script>

<div
  class="rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-5"
>
  <h3 class="mb-4 text-sm font-medium text-[var(--color-text-secondary)]">
    {$t('finance.overview.netWorth')}
  </h3>

  <div class="mb-5 flex items-center gap-3">
    <div
      class="flex h-12 w-12 shrink-0 items-center justify-center rounded-[var(--radius-md)] bg-[var(--color-brand-blue)]/10 text-[var(--color-brand-blue)]"
    >
      <Gem size={24} />
    </div>
    <div>
      <p class="text-xs text-[var(--color-text-tertiary)]">
        {$t('finance.overview.totalNetWorth')}
      </p>
      <p class="text-2xl font-bold">
        <AmountDisplay amount={netWorth.net_worth} showSign />
      </p>
    </div>
  </div>

  <div class="flex flex-col gap-3">
    <!-- Accounts total -->
    <div
      class="flex items-center justify-between rounded-[var(--radius-md)] bg-[var(--color-bg-tertiary)] px-4 py-3"
    >
      <div class="flex items-center gap-2">
        <Landmark size={16} class="text-[var(--color-text-tertiary)]" />
        <span class="text-sm text-[var(--color-text-secondary)]">
          {$t('finance.overview.accountsTotal')}
        </span>
      </div>
      <span class="text-sm font-semibold">
        <AmountDisplay amount={netWorth.accounts_total} />
      </span>
    </div>

    <!-- Investments total -->
    <div
      class="flex items-center justify-between rounded-[var(--radius-md)] bg-[var(--color-bg-tertiary)] px-4 py-3"
    >
      <div class="flex items-center gap-2">
        <TrendingUp size={16} class="text-[var(--color-text-tertiary)]" />
        <span class="text-sm text-[var(--color-text-secondary)]">
          {$t('finance.overview.investmentsTotal')}
        </span>
      </div>
      <span class="text-sm font-semibold">
        <AmountDisplay amount={netWorth.investments_total} />
      </span>
    </div>
  </div>
</div>
