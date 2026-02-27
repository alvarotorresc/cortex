<script lang="ts">
  import { t } from 'svelte-i18n';
  import Loader2 from 'lucide-svelte/icons/loader-2';
  import { getSummary, getTrends, listAccounts, getNetWorth } from '../api';
  import type { MonthlySummary, TrendPoint, AccountWithBalance, NetWorth } from '../types';
  import BalanceCard from './BalanceCard.svelte';
  import CategoryChart from './CategoryChart.svelte';
  import TrendChart from './TrendChart.svelte';
  import AccountsList from './AccountsList.svelte';
  import NetWorthCard from './NetWorthCard.svelte';

  interface Props {
    month: string;
  }

  const { month }: Props = $props();

  let summary = $state<MonthlySummary | null>(null);
  let trends = $state<TrendPoint[]>([]);
  let accounts = $state<AccountWithBalance[]>([]);
  let netWorth = $state<NetWorth | null>(null);
  let loading = $state(true);
  let error = $state('');

  const trendsFrom = $derived(() => {
    const [y, m] = month.split('-').map(Number);
    const d = new Date(y, m - 6);
    return `${d.getFullYear()}-${String(d.getMonth() + 1).padStart(2, '0')}`;
  });

  $effect(() => {
    loadData(month);
  });

  async function loadData(m: string): Promise<void> {
    loading = true;
    error = '';
    try {
      const [s, t, a, n] = await Promise.all([
        getSummary(m),
        getTrends(trendsFrom(), m),
        listAccounts(),
        getNetWorth(),
      ]);
      summary = s;
      trends = t;
      accounts = a;
      netWorth = n;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load overview data';
    } finally {
      loading = false;
    }
  }
</script>

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
{:else}
  <div class="space-y-4">
    <!-- Balance summary (full width) -->
    {#if summary}
      <BalanceCard {summary} />
    {/if}

    <!-- Charts row -->
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
      {#if summary}
        <CategoryChart categories={summary.by_category} />
      {/if}
      <TrendChart {trends} />
    </div>

    <!-- Accounts + Net Worth row -->
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
      <AccountsList {accounts} />
      {#if netWorth}
        <NetWorthCard {netWorth} />
      {/if}
    </div>
  </div>
{/if}
