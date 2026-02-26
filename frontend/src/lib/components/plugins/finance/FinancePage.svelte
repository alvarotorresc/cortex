<script lang="ts">
  import { t } from 'svelte-i18n';
  import { onMount } from 'svelte';
  import Wallet from 'lucide-svelte/icons/wallet';
  import Settings from 'lucide-svelte/icons/settings';
  import LayoutDashboard from 'lucide-svelte/icons/layout-dashboard';
  import ArrowLeftRight from 'lucide-svelte/icons/arrow-left-right';
  import PiggyBank from 'lucide-svelte/icons/piggy-bank';
  import Target from 'lucide-svelte/icons/target';
  import TrendingUp from 'lucide-svelte/icons/trending-up';
  import type { Component } from 'svelte';
  import { generateRecurring } from './api';
  import MonthPicker from './shared/MonthPicker.svelte';
  import OverviewTab from './overview/OverviewTab.svelte';
  import TransactionsTab from './transactions/TransactionsTab.svelte';
  import BudgetsTab from './budgets/BudgetsTab.svelte';
  import GoalsTab from './goals/GoalsTab.svelte';
  import InvestmentsTab from './investments/InvestmentsTab.svelte';
  import SettingsPanel from './settings/SettingsPanel.svelte';

  type TabId = 'overview' | 'transactions' | 'budgets' | 'goals' | 'investments';

  interface TabDefinition {
    id: TabId;
    labelKey: string;
    icon: Component<{ size?: number }>;
  }

  const tabs: TabDefinition[] = [
    { id: 'overview', labelKey: 'finance.tabs.overview', icon: LayoutDashboard },
    { id: 'transactions', labelKey: 'finance.tabs.transactions', icon: ArrowLeftRight },
    { id: 'budgets', labelKey: 'finance.tabs.budgets', icon: PiggyBank },
    { id: 'goals', labelKey: 'finance.tabs.goals', icon: Target },
    { id: 'investments', labelKey: 'finance.tabs.investments', icon: TrendingUp },
  ];

  let activeTab = $state<TabId>('overview');
  let showSettings = $state(false);

  const now = new Date();
  let currentMonth = $state(
    `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`,
  );

  function handleMonthChange(month: string): void {
    currentMonth = month;
  }

  function selectTab(tabId: TabId): void {
    showSettings = false;
    activeTab = tabId;
  }

  function toggleSettings(): void {
    showSettings = !showSettings;
  }

  onMount(() => {
    generateRecurring().catch(() => {
      // Silently ignore — recurring generation is best-effort on page load
    });
  });
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
    <div class="flex items-center gap-3">
      <Wallet size={20} class="text-[var(--color-plugin-finance)]" />
      <h2 class="text-2xl font-semibold text-[var(--color-text-primary)]">
        {$t('finance.title')}
      </h2>
    </div>

    <div class="flex items-center gap-3">
      <MonthPicker month={currentMonth} onchange={handleMonthChange} />

      <button
        onclick={toggleSettings}
        class="rounded-[var(--radius-md)] p-2 transition-colors {showSettings
          ? 'bg-[var(--color-bg-tertiary)] text-[var(--color-brand-blue)]'
          : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]'}"
        aria-label={$t('finance.settings')}
        aria-pressed={showSettings}
      >
        <Settings size={20} />
      </button>
    </div>
  </div>

  <!-- Settings panel (replaces tab content when open) -->
  {#if showSettings}
    <SettingsPanel onclose={() => (showSettings = false)} />
  {:else}
    <!-- Tab bar -->
    <div
      class="-mb-px flex gap-1 overflow-x-auto border-b border-[var(--color-border)]"
      role="tablist"
      aria-label={$t('finance.title')}
    >
      {#each tabs as tab (tab.id)}
        {@const isActive = activeTab === tab.id}
        {@const IconComponent = tab.icon}
        <button
          role="tab"
          aria-selected={isActive}
          aria-controls="finance-tabpanel-{tab.id}"
          id="finance-tab-{tab.id}"
          onclick={() => selectTab(tab.id)}
          class="flex shrink-0 items-center gap-2 border-b-2 px-4 py-2.5 text-sm font-medium transition-colors {isActive
            ? 'border-[var(--color-brand-blue)] text-[var(--color-brand-blue)]'
            : 'border-transparent text-[var(--color-text-tertiary)] hover:border-[var(--color-border)] hover:text-[var(--color-text-secondary)]'}"
        >
          <IconComponent size={16} />
          <span class="hidden sm:inline">{$t(tab.labelKey)}</span>
        </button>
      {/each}
    </div>

    <!-- Tab content -->
    <div
      id="finance-tabpanel-{activeTab}"
      role="tabpanel"
      aria-labelledby="finance-tab-{activeTab}"
    >
      {#if activeTab === 'overview'}
        <OverviewTab month={currentMonth} />
      {:else if activeTab === 'transactions'}
        <TransactionsTab month={currentMonth} />
      {:else if activeTab === 'budgets'}
        <BudgetsTab month={currentMonth} />
      {:else if activeTab === 'goals'}
        <GoalsTab />
      {:else if activeTab === 'investments'}
        <InvestmentsTab />
      {/if}
    </div>
  {/if}
</div>
