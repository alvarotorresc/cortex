<script lang="ts">
  import { t } from 'svelte-i18n';
  import X from 'lucide-svelte/icons/x';
  import LayoutList from 'lucide-svelte/icons/layout-list';
  import TagIcon from 'lucide-svelte/icons/tag';
  import Landmark from 'lucide-svelte/icons/landmark';
  import Repeat from 'lucide-svelte/icons/repeat';
  import type { Component } from 'svelte';
  import CategoriesManager from './CategoriesManager.svelte';
  import TagsManager from './TagsManager.svelte';
  import AccountsManager from './AccountsManager.svelte';
  import RecurringManager from '../recurring/RecurringManager.svelte';

  interface Props {
    onclose: () => void;
  }

  const { onclose }: Props = $props();

  type SettingsTab = 'categories' | 'tags' | 'accounts' | 'recurring';

  interface TabDefinition {
    id: SettingsTab;
    labelKey: string;
    icon: Component<{ size?: number }>;
  }

  const tabs: TabDefinition[] = [
    { id: 'categories', labelKey: 'finance.settingsPanel.categories', icon: LayoutList },
    { id: 'tags', labelKey: 'finance.settingsPanel.tags', icon: TagIcon },
    { id: 'accounts', labelKey: 'finance.settingsPanel.accounts', icon: Landmark },
    { id: 'recurring', labelKey: 'finance.settingsPanel.recurring', icon: Repeat },
  ];

  let activeTab = $state<SettingsTab>('categories');
</script>

<div class="space-y-4">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">
      {$t('finance.settings')}
    </h3>
    <button
      onclick={onclose}
      class="rounded-[var(--radius-md)] p-2 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
      aria-label={$t('finance.closeSettings')}
    >
      <X size={20} />
    </button>
  </div>

  <!-- Sub-tab bar -->
  <div
    class="-mb-px flex gap-1 overflow-x-auto border-b border-[var(--color-border)]"
    role="tablist"
    aria-label={$t('finance.settings')}
  >
    {#each tabs as tab (tab.id)}
      {@const isActive = activeTab === tab.id}
      {@const IconComponent = tab.icon}
      <button
        role="tab"
        aria-selected={isActive}
        aria-controls="settings-tabpanel-{tab.id}"
        id="settings-tab-{tab.id}"
        onclick={() => (activeTab = tab.id)}
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
    id="settings-tabpanel-{activeTab}"
    role="tabpanel"
    aria-labelledby="settings-tab-{activeTab}"
  >
    {#if activeTab === 'categories'}
      <CategoriesManager />
    {:else if activeTab === 'tags'}
      <TagsManager />
    {:else if activeTab === 'accounts'}
      <AccountsManager />
    {:else if activeTab === 'recurring'}
      <RecurringManager />
    {/if}
  </div>
</div>
