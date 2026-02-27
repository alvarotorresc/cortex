<script lang="ts">
  import { t } from 'svelte-i18n';
  import Plus from 'lucide-svelte/icons/plus';
  import Pencil from 'lucide-svelte/icons/pencil';
  import Archive from 'lucide-svelte/icons/archive';
  import Landmark from 'lucide-svelte/icons/landmark';
  import Loader2 from 'lucide-svelte/icons/loader-2';
  import Check from 'lucide-svelte/icons/check';
  import X from 'lucide-svelte/icons/x';
  import { listAccounts, createAccount, updateAccount, archiveAccount } from '../api';
  import type {
    AccountWithBalance,
    AccountType,
    CreateAccountInput,
    UpdateAccountInput,
  } from '../types';
  import EmptyState from '../shared/EmptyState.svelte';
  import AmountDisplay from '../shared/AmountDisplay.svelte';
  import LucideIcon from '../shared/LucideIcon.svelte';

  // State
  let accounts = $state<AccountWithBalance[]>([]);
  let loading = $state(true);
  let error = $state('');

  // Form state
  let showForm = $state(false);
  let editingId = $state<number | null>(null);
  let formName = $state('');
  let formType = $state<AccountType>('checking');
  let formCurrency = $state('EUR');
  let formInterestRate = $state<number | undefined>(undefined);
  let formIcon = $state('');
  let formColor = $state('#3b82f6');
  let saving = $state(false);
  let formError = $state('');

  const DEFAULT_COLORS = [
    '#ef4444', '#f97316', '#eab308', '#22c55e', '#06b6d4',
    '#3b82f6', '#6366f1', '#8b5cf6', '#ec4899', '#64748b',
  ];

  const ACCOUNT_TYPES: AccountType[] = ['checking', 'savings', 'cash', 'investment'];

  $effect(() => {
    loadAccounts();
  });

  async function loadAccounts(): Promise<void> {
    loading = true;
    error = '';
    try {
      accounts = await listAccounts();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load accounts';
    } finally {
      loading = false;
    }
  }

  function openAddForm(): void {
    editingId = null;
    formName = '';
    formType = 'checking';
    formCurrency = 'EUR';
    formInterestRate = undefined;
    formIcon = '';
    formColor = '#3b82f6';
    formError = '';
    showForm = true;
  }

  function openEditForm(acct: AccountWithBalance): void {
    editingId = acct.id;
    formName = acct.name;
    formType = acct.type;
    formCurrency = acct.currency;
    formInterestRate = acct.interest_rate;
    formIcon = acct.icon;
    formColor = acct.color;
    formError = '';
    showForm = true;
  }

  function cancelForm(): void {
    showForm = false;
    editingId = null;
    formError = '';
  }

  async function handleSubmit(): Promise<void> {
    if (saving || !formName.trim()) return;
    saving = true;
    formError = '';

    try {
      if (editingId !== null) {
        const input: UpdateAccountInput = {
          name: formName.trim(),
          type: formType,
          currency: formCurrency,
          interest_rate: formType === 'savings' ? formInterestRate : undefined,
          icon: formIcon,
          color: formColor,
        };
        await updateAccount(editingId, input);
      } else {
        const input: CreateAccountInput = {
          name: formName.trim(),
          type: formType,
          currency: formCurrency || undefined,
          interest_rate: formType === 'savings' ? formInterestRate : undefined,
          icon: formIcon,
          color: formColor,
        };
        await createAccount(input);
      }

      showForm = false;
      editingId = null;
      await loadAccounts();
    } catch (err) {
      formError = err instanceof Error ? err.message : 'Failed to save account';
    } finally {
      saving = false;
    }
  }

  async function handleArchive(acct: AccountWithBalance): Promise<void> {
    const confirmed = window.confirm($t('finance.settingsPanel.confirmArchiveAccount'));
    if (!confirmed) return;

    try {
      await archiveAccount(acct.id);
      await loadAccounts();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to archive account';
    }
  }

  function getAccountTypeLabel(type: AccountType): string {
    return $t(`finance.overview.accountTypes.${type}`);
  }
</script>

<div class="space-y-3">
  <!-- Header with Add button -->
  <div class="flex items-center justify-end">
    <button
      onclick={openAddForm}
      class="flex shrink-0 items-center gap-1.5 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-3 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
    >
      <Plus size={16} />
      <span class="hidden sm:inline">{$t('finance.settingsPanel.addAccount')}</span>
    </button>
  </div>

  <!-- Inline form (add mode) -->
  {#if showForm && editingId === null}
    <form
      onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}
      class="rounded-[var(--radius-lg)] border border-[var(--color-brand-blue)] bg-[var(--color-bg-secondary)] p-4"
    >
      <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
        <!-- Name -->
        <label class="flex flex-col gap-1">
          <span class="text-xs font-medium text-[var(--color-text-secondary)]">
            {$t('finance.settingsPanel.name')}
          </span>
          <input
            type="text"
            bind:value={formName}
            required
            class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
          />
        </label>

        <!-- Type -->
        <label class="flex flex-col gap-1">
          <span class="text-xs font-medium text-[var(--color-text-secondary)]">
            {$t('finance.type')}
          </span>
          <select
            bind:value={formType}
            class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
          >
            {#each ACCOUNT_TYPES as accType (accType)}
              <option value={accType}>{getAccountTypeLabel(accType)}</option>
            {/each}
          </select>
        </label>

        <!-- Currency -->
        <label class="flex flex-col gap-1">
          <span class="text-xs font-medium text-[var(--color-text-secondary)]">
            {$t('finance.investments.currency')}
          </span>
          <input
            type="text"
            bind:value={formCurrency}
            maxlength={3}
            class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
          />
        </label>

        <!-- Interest Rate (only for savings) -->
        {#if formType === 'savings'}
          <label class="flex flex-col gap-1">
            <span class="text-xs font-medium text-[var(--color-text-secondary)]">
              {$t('finance.settingsPanel.interestRate')} (%)
            </span>
            <input
              type="number"
              bind:value={formInterestRate}
              min="0"
              step="0.01"
              class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
            />
          </label>
        {/if}

        <!-- Icon -->
        <label class="flex flex-col gap-1">
          <span class="text-xs font-medium text-[var(--color-text-secondary)]">
            {$t('finance.settingsPanel.icon')}
          </span>
          <input
            type="text"
            bind:value={formIcon}
            placeholder="e.g. emoji"
            class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
          />
        </label>

        <!-- Color -->
        <div class="flex flex-col gap-1">
          <span class="text-xs font-medium text-[var(--color-text-secondary)]">
            {$t('finance.settingsPanel.color')}
          </span>
          <div class="flex flex-wrap gap-1.5">
            {#each DEFAULT_COLORS as color (color)}
              <button
                type="button"
                onclick={() => (formColor = color)}
                class="h-6 w-6 rounded-full border-2 transition-transform hover:scale-110 {formColor === color ? 'border-[var(--color-text-primary)] scale-110' : 'border-transparent'}"
                style:background-color={color}
                aria-label={color}
              ></button>
            {/each}
          </div>
        </div>
      </div>

      {#if formError}
        <p class="mt-2 rounded-[var(--radius-md)] bg-[var(--color-error)]/10 px-3 py-1.5 text-xs text-[var(--color-error)]">
          {formError}
        </p>
      {/if}

      <div class="mt-3 flex items-center justify-end gap-2">
        <button
          type="button"
          onclick={cancelForm}
          disabled={saving}
          class="flex items-center gap-1 rounded-[var(--radius-md)] px-3 py-1.5 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)] disabled:opacity-50"
        >
          <X size={14} />
          {$t('finance.cancel')}
        </button>
        <button
          type="submit"
          disabled={saving || !formName.trim()}
          class="flex items-center gap-1 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-3 py-1.5 text-sm font-medium text-white transition-colors hover:opacity-90 disabled:opacity-50"
        >
          <Check size={14} />
          {saving ? $t('finance.saving') : $t('finance.save')}
        </button>
      </div>
    </form>
  {/if}

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
        onclick={loadAccounts}
        class="mt-3 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
      >
        {$t('common.retry')}
      </button>
    </div>
  {:else if accounts.length === 0}
    <EmptyState
      icon={Landmark}
      message={$t('finance.settingsPanel.noAccounts')}
      actionLabel={$t('finance.settingsPanel.addAccount')}
      onaction={openAddForm}
    />
  {:else}
    <div class="divide-y divide-[var(--color-border)] rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]">
      {#each accounts as acct (acct.id)}
        {#if showForm && editingId === acct.id}
          <!-- Inline edit form -->
          <form
            onsubmit={(e) => { e.preventDefault(); handleSubmit(); }}
            class="border-l-2 border-l-[var(--color-brand-blue)] p-4"
          >
            <div class="grid grid-cols-1 gap-3 sm:grid-cols-2">
              <label class="flex flex-col gap-1">
                <span class="text-xs font-medium text-[var(--color-text-secondary)]">
                  {$t('finance.settingsPanel.name')}
                </span>
                <input
                  type="text"
                  bind:value={formName}
                  required
                  class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
                />
              </label>

              <label class="flex flex-col gap-1">
                <span class="text-xs font-medium text-[var(--color-text-secondary)]">
                  {$t('finance.type')}
                </span>
                <select
                  bind:value={formType}
                  class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
                >
                  {#each ACCOUNT_TYPES as accType (accType)}
                    <option value={accType}>{getAccountTypeLabel(accType)}</option>
                  {/each}
                </select>
              </label>

              <label class="flex flex-col gap-1">
                <span class="text-xs font-medium text-[var(--color-text-secondary)]">
                  {$t('finance.investments.currency')}
                </span>
                <input
                  type="text"
                  bind:value={formCurrency}
                  maxlength={3}
                  class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
                />
              </label>

              {#if formType === 'savings'}
                <label class="flex flex-col gap-1">
                  <span class="text-xs font-medium text-[var(--color-text-secondary)]">
                    {$t('finance.settingsPanel.interestRate')} (%)
                  </span>
                  <input
                    type="number"
                    bind:value={formInterestRate}
                    min="0"
                    step="0.01"
                    class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
                  />
                </label>
              {/if}

              <label class="flex flex-col gap-1">
                <span class="text-xs font-medium text-[var(--color-text-secondary)]">
                  {$t('finance.settingsPanel.icon')}
                </span>
                <input
                  type="text"
                  bind:value={formIcon}
                  placeholder="e.g. emoji"
                  class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
                />
              </label>

              <div class="flex flex-col gap-1">
                <span class="text-xs font-medium text-[var(--color-text-secondary)]">
                  {$t('finance.settingsPanel.color')}
                </span>
                <div class="flex flex-wrap gap-1.5">
                  {#each DEFAULT_COLORS as color (color)}
                    <button
                      type="button"
                      onclick={() => (formColor = color)}
                      class="h-6 w-6 rounded-full border-2 transition-transform hover:scale-110 {formColor === color ? 'border-[var(--color-text-primary)] scale-110' : 'border-transparent'}"
                      style:background-color={color}
                      aria-label={color}
                    ></button>
                  {/each}
                </div>
              </div>
            </div>

            {#if formError}
              <p class="mt-2 rounded-[var(--radius-md)] bg-[var(--color-error)]/10 px-3 py-1.5 text-xs text-[var(--color-error)]">
                {formError}
              </p>
            {/if}

            <div class="mt-3 flex items-center justify-end gap-2">
              <button
                type="button"
                onclick={cancelForm}
                disabled={saving}
                class="flex items-center gap-1 rounded-[var(--radius-md)] px-3 py-1.5 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)] disabled:opacity-50"
              >
                <X size={14} />
                {$t('finance.cancel')}
              </button>
              <button
                type="submit"
                disabled={saving || !formName.trim()}
                class="flex items-center gap-1 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-3 py-1.5 text-sm font-medium text-white transition-colors hover:opacity-90 disabled:opacity-50"
              >
                <Check size={14} />
                {saving ? $t('finance.saving') : $t('finance.save')}
              </button>
            </div>
          </form>
        {:else}
          <!-- Account row -->
          <div class="flex items-center gap-3 px-4 py-3 {acct.is_archived ? 'opacity-50' : ''}">
            <!-- Icon or color dot -->
            {#if acct.icon}
              <span
                class="flex h-8 w-8 shrink-0 items-center justify-center rounded-[var(--radius-md)] text-sm"
                style:background-color="{acct.color}20"
                aria-hidden="true"
              >
                <LucideIcon name={acct.icon} />
              </span>
            {:else}
              <span
                class="h-3 w-3 shrink-0 rounded-full"
                style:background-color={acct.color}
                aria-hidden="true"
              ></span>
            {/if}

            <!-- Name -->
            <span class="min-w-0 flex-1 truncate text-sm font-medium text-[var(--color-text-primary)]">
              {acct.name}
              {#if acct.is_archived}
                <span class="ml-1 text-xs text-[var(--color-text-tertiary)]">
                  ({$t('finance.settingsPanel.archived')})
                </span>
              {/if}
            </span>

            <!-- Type badge -->
            <span class="hidden shrink-0 rounded-full bg-[var(--color-bg-tertiary)] px-2 py-0.5 text-xs font-medium text-[var(--color-text-secondary)] sm:inline">
              {getAccountTypeLabel(acct.type)}
            </span>

            <!-- Currency -->
            <span class="shrink-0 text-xs text-[var(--color-text-tertiary)]">
              {acct.currency}
            </span>

            <!-- Balance -->
            <span class="shrink-0 text-sm font-medium">
              <AmountDisplay amount={acct.balance} />
            </span>

            <!-- Actions -->
            <div class="flex shrink-0 items-center gap-1">
              {#if !acct.is_archived}
                <button
                  onclick={() => openEditForm(acct)}
                  class="rounded-[var(--radius-sm)] p-1.5 text-[var(--color-text-tertiary)] transition-colors hover:bg-[var(--color-bg-tertiary)] hover:text-[var(--color-text-secondary)]"
                  aria-label="Edit {acct.name}"
                >
                  <Pencil size={14} />
                </button>
                <button
                  onclick={() => handleArchive(acct)}
                  class="rounded-[var(--radius-sm)] p-1.5 text-[var(--color-text-tertiary)] transition-colors hover:bg-[var(--color-error)]/10 hover:text-[var(--color-error)]"
                  aria-label="{$t('finance.settingsPanel.archive')} {acct.name}"
                >
                  <Archive size={14} />
                </button>
              {/if}
            </div>
          </div>
        {/if}
      {/each}
    </div>
  {/if}
</div>
