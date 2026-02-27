<script lang="ts">
  import { t } from 'svelte-i18n';
  import Plus from 'lucide-svelte/icons/plus';
  import Repeat from 'lucide-svelte/icons/repeat';
  import Loader2 from 'lucide-svelte/icons/loader-2';
  import {
    listRecurringRules,
    listCategories,
    listAccounts,
  } from '../api';
  import type { RecurringRule, Category, AccountWithBalance, Frequency } from '../types';
  import EmptyState from '../shared/EmptyState.svelte';
  import AmountDisplay from '../shared/AmountDisplay.svelte';
  import RecurringForm from './RecurringForm.svelte';

  // State
  let rules = $state<RecurringRule[]>([]);
  let categories = $state<Category[]>([]);
  let accounts = $state<AccountWithBalance[]>([]);
  let loading = $state(true);
  let error = $state('');
  let showForm = $state(false);
  let editingRule = $state<RecurringRule | null>(null);

  $effect(() => {
    loadData();
  });

  async function loadData(): Promise<void> {
    loading = true;
    error = '';
    try {
      const [ruleList, catList, acctList] = await Promise.all([
        listRecurringRules(),
        listCategories(),
        listAccounts(),
      ]);
      rules = ruleList;
      categories = catList;
      accounts = acctList;
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load data';
    } finally {
      loading = false;
    }
  }

  async function loadRules(): Promise<void> {
    try {
      rules = await listRecurringRules();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load rules';
    }
  }

  function handleAdd(): void {
    editingRule = null;
    showForm = true;
  }

  function handleEdit(rule: RecurringRule): void {
    editingRule = rule;
    showForm = true;
  }

  function handleSave(): void {
    showForm = false;
    editingRule = null;
    loadRules();
  }

  function handleDelete(): void {
    showForm = false;
    editingRule = null;
    loadRules();
  }

  function handleCancel(): void {
    showForm = false;
    editingRule = null;
  }

  function getFrequencyLabel(freq: Frequency): string {
    return $t(`finance.settingsPanel.${freq}`);
  }

  function getAccountName(accountId: number): string {
    const acct = accounts.find((a) => a.id === accountId);
    return acct?.name ?? '—';
  }

  function handleRuleKeydown(e: KeyboardEvent, rule: RecurringRule): void {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleEdit(rule);
    }
  }
</script>

<div class="space-y-3">
  <!-- Header with Add button -->
  <div class="flex items-center justify-end">
    <button
      onclick={handleAdd}
      class="flex shrink-0 items-center gap-1.5 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-3 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
    >
      <Plus size={16} />
      <span class="hidden sm:inline">{$t('finance.settingsPanel.addRecurring')}</span>
    </button>
  </div>

  <!-- Content area -->
  {#if loading}
    <div class="flex items-center justify-center py-12">
      <Loader2 size={24} class="animate-spin text-[var(--color-text-tertiary)]" />
    </div>
  {:else if error}
    <div
      class="flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-6 py-12 text-center"
    >
      <p class="text-sm text-[var(--color-error)]">{error}</p>
      <button
        onclick={loadData}
        class="mt-3 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
      >
        {$t('common.retry')}
      </button>
    </div>
  {:else if rules.length === 0}
    <EmptyState
      icon={Repeat}
      message={$t('finance.settingsPanel.noRecurring')}
      actionLabel={$t('finance.settingsPanel.addRecurring')}
      onaction={handleAdd}
    />
  {:else}
    <div class="divide-y divide-[var(--color-border)] rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]">
      {#each rules as rule (rule.id)}
        <!-- svelte-ignore a11y_no_static_element_interactions -->
        <div
          role="button"
          tabindex="0"
          onclick={() => handleEdit(rule)}
          onkeydown={(e) => handleRuleKeydown(e, rule)}
          class="flex cursor-pointer items-center gap-3 px-4 py-3 transition-colors hover:bg-[var(--color-bg-tertiary)] {rule.is_active ? '' : 'opacity-50'}"
        >
          <!-- Description + Category -->
          <div class="min-w-0 flex-1">
            <p class="truncate text-sm font-medium text-[var(--color-text-primary)]">
              {rule.description}
            </p>
            <div class="flex items-center gap-2 text-xs text-[var(--color-text-tertiary)]">
              <span>{rule.category}</span>
              <span>·</span>
              <span>{getAccountName(rule.account_id)}</span>
            </div>
          </div>

          <!-- Amount -->
          <span class="shrink-0 text-sm font-semibold {rule.type === 'income' ? 'text-[var(--color-success)]' : rule.type === 'expense' ? 'text-[var(--color-error)]' : 'text-[var(--color-brand-blue)]'}">
            {rule.type === 'income' ? '+' : rule.type === 'expense' ? '-' : ''}<AmountDisplay amount={rule.amount} />
          </span>

          <!-- Frequency badge -->
          <span class="hidden shrink-0 rounded-full bg-[var(--color-bg-tertiary)] px-2 py-0.5 text-xs font-medium text-[var(--color-text-secondary)] sm:inline">
            {getFrequencyLabel(rule.frequency)}
          </span>

          <!-- Active/Inactive badge -->
          {#if !rule.is_active}
            <span class="shrink-0 rounded-full bg-[var(--color-error)]/10 px-2 py-0.5 text-xs font-medium text-[var(--color-error)]">
              {$t('finance.settingsPanel.inactive')}
            </span>
          {/if}
        </div>
      {/each}
    </div>
  {/if}
</div>

<!-- Form modal -->
{#if showForm}
  <RecurringForm
    rule={editingRule}
    {categories}
    {accounts}
    onsave={handleSave}
    oncancel={handleCancel}
    ondelete={handleDelete}
  />
{/if}
