<script lang="ts">
  import { t } from 'svelte-i18n';
  import X from 'lucide-svelte/icons/x';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import { createTransaction, updateTransaction, deleteTransaction } from '../api';
  import type {
    Transaction,
    TransactionType,
    Category,
    AccountWithBalance,
    Tag,
    CreateTransactionInput,
    UpdateTransactionInput,
  } from '../types';

  interface Props {
    transaction?: Transaction | null;
    categories: Category[];
    accounts: AccountWithBalance[];
    tags: Tag[];
    onsave: () => void;
    oncancel: () => void;
    ondelete?: () => void;
  }

  const { transaction = null, categories, accounts, tags, onsave, oncancel, ondelete }: Props =
    $props();

  const isEditing = $derived(transaction !== null);

  // Form state — initialized from transaction prop (form mounts fresh each time).
  // svelte-ignore state_referenced_locally
  let amount = $state(transaction?.amount ?? 0);
  // svelte-ignore state_referenced_locally
  let type = $state<TransactionType>(transaction?.type ?? 'expense');
  // svelte-ignore state_referenced_locally
  let accountId = $state(transaction?.account_id ?? (accounts[0]?.id ?? 0));
  // svelte-ignore state_referenced_locally
  let destAccountId = $state(transaction?.dest_account_id ?? 0);
  // svelte-ignore state_referenced_locally
  let categoryName = $state(transaction?.category ?? '');
  // svelte-ignore state_referenced_locally
  let description = $state(transaction?.description ?? '');
  // svelte-ignore state_referenced_locally
  let date = $state(transaction?.date?.split('T')[0] ?? new Date().toISOString().split('T')[0]);
  // svelte-ignore state_referenced_locally
  let selectedTagIds = $state<Set<number>>(
    new Set(transaction?.tags.map((tg) => tg.id) ?? []),
  );

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

  // Reset category when type changes and current selection is invalid
  $effect(() => {
    // Depend on type to trigger
    const _type = type;
    const validNames = filteredCategories.map((c) => c.name);
    if (categoryName && !validNames.includes(categoryName)) {
      categoryName = validNames[0] ?? '';
    }
  });

  function toggleTag(tagId: number): void {
    const next = new Set(selectedTagIds);
    if (next.has(tagId)) {
      next.delete(tagId);
    } else {
      next.add(tagId);
    }
    selectedTagIds = next;
  }

  async function handleSubmit(): Promise<void> {
    if (saving || deleting) return;
    error = '';
    saving = true;

    try {
      const input: CreateTransactionInput | UpdateTransactionInput = {
        amount,
        type,
        account_id: accountId || undefined,
        dest_account_id: type === 'transfer' && destAccountId ? destAccountId : undefined,
        category: categoryName,
        description,
        date,
        tag_ids: selectedTagIds.size > 0 ? [...selectedTagIds] : undefined,
      };

      if (isEditing && transaction) {
        await updateTransaction(transaction.id, input);
      } else {
        await createTransaction(input);
      }
      onsave();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to save transaction';
    } finally {
      saving = false;
    }
  }

  async function handleDelete(): Promise<void> {
    if (!isEditing || !transaction || deleting || saving) return;

    const confirmed = window.confirm($t('finance.confirmDelete'));
    if (!confirmed) return;

    error = '';
    deleting = true;
    try {
      await deleteTransaction(transaction.id);
      ondelete?.();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete transaction';
    } finally {
      deleting = false;
    }
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
  aria-label={isEditing ? $t('finance.editTransaction') : $t('finance.addTransaction')}
>
  <!-- Header -->
  <div class="flex items-center justify-between border-b border-[var(--color-border)] px-6 py-4">
    <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">
      {isEditing ? $t('finance.editTransaction') : $t('finance.addTransaction')}
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

    <!-- Date -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.date')}
      </span>
      <input
        type="date"
        bind:value={date}
        required
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Tags -->
    {#if tags.length > 0}
      <fieldset>
        <legend class="mb-1.5 text-sm font-medium text-[var(--color-text-secondary)]">
          {$t('finance.tags')}
        </legend>
        <div class="flex flex-wrap gap-2">
          {#each tags as tag (tag.id)}
            {@const isSelected = selectedTagIds.has(tag.id)}
            <button
              type="button"
              onclick={() => toggleTag(tag.id)}
              class="rounded-full border px-3 py-1 text-xs font-medium transition-colors {isSelected
                ? 'border-transparent'
                : 'border-[var(--color-border)] text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]'}"
              style:background-color={isSelected ? `${tag.color}20` : undefined}
              style:color={isSelected ? tag.color : undefined}
              style:border-color={isSelected ? tag.color : undefined}
              aria-pressed={isSelected}
            >
              {tag.name}
            </button>
          {/each}
        </div>
      </fieldset>
    {/if}

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
