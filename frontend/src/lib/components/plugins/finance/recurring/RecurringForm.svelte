<script lang="ts">
  import { t } from 'svelte-i18n';
  import X from 'lucide-svelte/icons/x';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import {
    createRecurringRule,
    updateRecurringRule,
    deleteRecurringRule,
  } from '../api';
  import type {
    RecurringRule,
    TransactionType,
    Frequency,
    Category,
    AccountWithBalance,
    CreateRecurringRuleInput,
    UpdateRecurringRuleInput,
  } from '../types';

  interface Props {
    rule?: RecurringRule | null;
    categories: Category[];
    accounts: AccountWithBalance[];
    onsave: () => void;
    oncancel: () => void;
    ondelete?: () => void;
  }

  const { rule = null, categories, accounts, onsave, oncancel, ondelete }: Props = $props();

  const isEditing = $derived(rule !== null);

  // Form state — initialized from rule prop (form mounts fresh each time).
  // svelte-ignore state_referenced_locally
  let description = $state(rule?.description ?? '');
  // svelte-ignore state_referenced_locally
  let amount = $state(rule?.amount ?? 0);
  // svelte-ignore state_referenced_locally
  let type = $state<TransactionType>(rule?.type ?? 'expense');
  // svelte-ignore state_referenced_locally
  let accountId = $state(rule?.account_id ?? (accounts[0]?.id ?? 0));
  // svelte-ignore state_referenced_locally
  let destAccountId = $state(rule?.dest_account_id ?? 0);
  // svelte-ignore state_referenced_locally
  let categoryName = $state(rule?.category ?? '');
  // svelte-ignore state_referenced_locally
  let frequency = $state<Frequency>(rule?.frequency ?? 'monthly');
  // svelte-ignore state_referenced_locally
  let dayOfMonth = $state<number | undefined>(rule?.day_of_month ?? 1);
  // svelte-ignore state_referenced_locally
  let dayOfWeek = $state<number | undefined>(rule?.day_of_week ?? 1);
  // svelte-ignore state_referenced_locally
  let monthOfYear = $state<number | undefined>(rule?.month_of_year ?? 1);
  // svelte-ignore state_referenced_locally
  let startDate = $state(
    rule?.start_date?.split('T')[0] ?? new Date().toISOString().split('T')[0],
  );
  // svelte-ignore state_referenced_locally
  let endDate = $state(rule?.end_date?.split('T')[0] ?? '');

  let saving = $state(false);
  let deleting = $state(false);
  let error = $state('');

  // Filter categories by transaction type
  const filteredCategories = $derived(
    categories.filter((c) => {
      if (type === 'transfer') return true;
      return c.type === type || c.type === 'both';
    }),
  );

  // Show day_of_week for weekly/biweekly
  const showDayOfWeek = $derived(frequency === 'weekly' || frequency === 'biweekly');
  // Show day_of_month for monthly/yearly
  const showDayOfMonth = $derived(frequency === 'monthly' || frequency === 'yearly');
  // Show month_of_year only for yearly
  const showMonthOfYear = $derived(frequency === 'yearly');

  // Reset category when type changes and current selection is invalid
  $effect(() => {
    const _type = type;
    const validNames = filteredCategories.map((c) => c.name);
    if (categoryName && !validNames.includes(categoryName)) {
      categoryName = validNames[0] ?? '';
    }
  });

  async function handleSubmit(): Promise<void> {
    if (saving || deleting) return;
    error = '';
    saving = true;

    try {
      const input: CreateRecurringRuleInput | UpdateRecurringRuleInput = {
        description,
        amount,
        type,
        account_id: accountId || undefined,
        dest_account_id: type === 'transfer' && destAccountId ? destAccountId : undefined,
        category: categoryName,
        frequency,
        day_of_month: showDayOfMonth ? dayOfMonth : undefined,
        day_of_week: showDayOfWeek ? dayOfWeek : undefined,
        month_of_year: showMonthOfYear ? monthOfYear : undefined,
        start_date: startDate,
        end_date: endDate || undefined,
      };

      if (isEditing && rule) {
        await updateRecurringRule(rule.id, input);
      } else {
        await createRecurringRule(input);
      }
      onsave();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to save recurring rule';
    } finally {
      saving = false;
    }
  }

  async function handleDelete(): Promise<void> {
    if (!isEditing || !rule || deleting || saving) return;

    const confirmed = window.confirm($t('finance.settingsPanel.confirmDeleteRule'));
    if (!confirmed) return;

    error = '';
    deleting = true;
    try {
      await deleteRecurringRule(rule.id);
      ondelete?.();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete recurring rule';
    } finally {
      deleting = false;
    }
  }

  function getFrequencyLabel(freq: Frequency): string {
    return $t(`finance.settingsPanel.${freq}`);
  }
</script>

<!-- Overlay backdrop -->
<div
  class="fixed inset-0 z-40 bg-black/50"
  role="presentation"
  onclick={oncancel}
  onkeydown={(e) => e.key === 'Escape' && oncancel()}
></div>

<!-- Modal panel -->
<div
  class="fixed inset-y-0 right-0 z-50 flex w-full max-w-md flex-col bg-[var(--color-bg-secondary)] shadow-xl"
  role="dialog"
  aria-modal="true"
  aria-label={isEditing ? $t('finance.settingsPanel.editRecurring') : $t('finance.settingsPanel.addRecurring')}
>
  <!-- Header -->
  <div class="flex items-center justify-between border-b border-[var(--color-border)] px-6 py-4">
    <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">
      {isEditing ? $t('finance.settingsPanel.editRecurring') : $t('finance.settingsPanel.addRecurring')}
    </h3>
    <button
      type="button"
      onclick={oncancel}
      class="rounded-[var(--radius-md)] p-2 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
      aria-label={$t('finance.cancel')}
    >
      <X size={20} />
    </button>
  </div>

  <!-- Form body -->
  <form
    onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}
    class="flex flex-1 flex-col gap-4 overflow-y-auto px-6 py-4"
  >
    <!-- Description -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.description')}
      </span>
      <input
        type="text"
        bind:value={description}
        required
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Amount -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.amount')}
      </span>
      <input
        type="number"
        bind:value={amount}
        min="0.01"
        step="0.01"
        required
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Type selector -->
    <fieldset>
      <legend class="mb-1.5 text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.type')}
      </legend>
      <div class="flex gap-1 rounded-[var(--radius-md)] border border-[var(--color-border)] p-1">
        {#each ['expense', 'income', 'transfer'] as txType (txType)}
          <button
            type="button"
            onclick={() => (type = txType as TransactionType)}
            class="flex-1 rounded-[var(--radius-sm)] px-3 py-1.5 text-sm font-medium transition-colors {type ===
            txType
              ? 'bg-[var(--color-brand-blue)] text-white'
              : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]'}"
          >
            {$t(`finance.${txType}`)}
          </button>
        {/each}
      </div>
    </fieldset>

    <!-- Account -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.account')}
      </span>
      <select
        bind:value={accountId}
        required
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      >
        {#each accounts as acct (acct.id)}
          <option value={acct.id}>{acct.name}</option>
        {/each}
      </select>
    </label>

    <!-- Destination account (transfer only) -->
    {#if type === 'transfer'}
      <label class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-[var(--color-text-secondary)]">
          {$t('finance.destinationAccount')}
        </span>
        <select
          bind:value={destAccountId}
          required
          class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
        >
          {#each accounts.filter((a) => a.id !== accountId) as acct (acct.id)}
            <option value={acct.id}>{acct.name}</option>
          {/each}
        </select>
      </label>
    {/if}

    <!-- Category -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.category')}
      </span>
      <select
        bind:value={categoryName}
        required
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      >
        {#each filteredCategories as cat (cat.id)}
          <option value={cat.name}>{cat.name}</option>
        {/each}
      </select>
    </label>

    <!-- Frequency -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.settingsPanel.frequency')}
      </span>
      <select
        bind:value={frequency}
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      >
        {#each ['weekly', 'biweekly', 'monthly', 'yearly'] as freq (freq)}
          <option value={freq}>{getFrequencyLabel(freq as Frequency)}</option>
        {/each}
      </select>
    </label>

    <!-- Day of Week (weekly/biweekly) -->
    {#if showDayOfWeek}
      <label class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-[var(--color-text-secondary)]">
          {$t('finance.settingsPanel.dayOfWeek')}
        </span>
        <select
          bind:value={dayOfWeek}
          class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
        >
          {#each [1, 2, 3, 4, 5, 6, 7] as day (day)}
            <option value={day}>
              {new Date(2026, 0, day + 4).toLocaleDateString(undefined, { weekday: 'long' })}
            </option>
          {/each}
        </select>
      </label>
    {/if}

    <!-- Day of Month (monthly/yearly) -->
    {#if showDayOfMonth}
      <label class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-[var(--color-text-secondary)]">
          {$t('finance.settingsPanel.dayOfMonth')}
        </span>
        <input
          type="number"
          bind:value={dayOfMonth}
          min={1}
          max={31}
          required
          class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
        />
      </label>
    {/if}

    <!-- Month of Year (yearly only) -->
    {#if showMonthOfYear}
      <label class="flex flex-col gap-1.5">
        <span class="text-sm font-medium text-[var(--color-text-secondary)]">
          {$t('finance.settingsPanel.monthOfYear')}
        </span>
        <select
          bind:value={monthOfYear}
          class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
        >
          {#each Array.from({ length: 12 }, (_, i) => i + 1) as m (m)}
            <option value={m}>
              {new Date(2026, m - 1, 1).toLocaleDateString(undefined, { month: 'long' })}
            </option>
          {/each}
        </select>
      </label>
    {/if}

    <!-- Start Date -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.settingsPanel.startDate')}
      </span>
      <input
        type="date"
        bind:value={startDate}
        required
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- End Date (optional) -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.settingsPanel.endDate')}
      </span>
      <input
        type="date"
        bind:value={endDate}
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Error -->
    {#if error}
      <p class="rounded-[var(--radius-md)] bg-[var(--color-error)]/10 px-3 py-2 text-sm text-[var(--color-error)]">
        {error}
      </p>
    {/if}

    <!-- Spacer to push buttons to bottom -->
    <div class="flex-1"></div>

    <!-- Actions -->
    <div class="flex items-center gap-3 border-t border-[var(--color-border)] pt-4">
      {#if isEditing}
        <button
          type="button"
          onclick={handleDelete}
          disabled={deleting || saving}
          class="flex items-center gap-1.5 rounded-[var(--radius-md)] px-3 py-2 text-sm font-medium text-[var(--color-error)] transition-colors hover:bg-[var(--color-error)]/10 disabled:opacity-50"
        >
          <Trash2 size={16} />
          {deleting ? $t('finance.deleting') : $t('finance.delete')}
        </button>
      {/if}

      <div class="flex-1"></div>

      <button
        type="button"
        onclick={oncancel}
        disabled={saving || deleting}
        class="rounded-[var(--radius-md)] px-4 py-2 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)] disabled:opacity-50"
      >
        {$t('finance.cancel')}
      </button>

      <button
        type="submit"
        disabled={saving || deleting}
        class="rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90 disabled:opacity-50"
      >
        {saving ? $t('finance.saving') : $t('finance.save')}
      </button>
    </div>
  </form>
</div>
