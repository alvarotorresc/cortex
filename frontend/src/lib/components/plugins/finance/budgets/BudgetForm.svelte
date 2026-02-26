<script lang="ts">
  import { t } from 'svelte-i18n';
  import X from 'lucide-svelte/icons/x';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import { createBudget, updateBudget, deleteBudget } from '../api';
  import type {
    BudgetWithProgress,
    Category,
    CreateBudgetInput,
    UpdateBudgetInput,
  } from '../types';

  interface Props {
    budget?: BudgetWithProgress | null;
    categories: Category[];
    month: string;
    onsave: () => void;
    oncancel: () => void;
    ondelete?: () => void;
  }

  const { budget = null, categories, month, onsave, oncancel, ondelete }: Props = $props();

  const isEditing = $derived(budget !== null);

  // Form state — initialized from budget prop (form mounts fresh each time).
  // svelte-ignore state_referenced_locally
  let name = $state(budget?.name ?? '');
  // svelte-ignore state_referenced_locally
  let categoryName = $state(budget?.category ?? '');
  // svelte-ignore state_referenced_locally
  let amount = $state(budget?.amount ?? 0);

  let saving = $state(false);
  let deleting = $state(false);
  let error = $state('');

  async function handleSubmit(): Promise<void> {
    if (saving || deleting) return;
    error = '';
    saving = true;

    try {
      const input: CreateBudgetInput | UpdateBudgetInput = {
        name,
        category: categoryName,
        amount,
        month,
      };

      if (isEditing && budget) {
        await updateBudget(budget.id, input);
      } else {
        await createBudget(input);
      }
      onsave();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to save budget';
    } finally {
      saving = false;
    }
  }

  async function handleDelete(): Promise<void> {
    if (!isEditing || !budget || deleting || saving) return;

    const confirmed = window.confirm($t('finance.budgets.confirmDelete'));
    if (!confirmed) return;

    error = '';
    deleting = true;
    try {
      await deleteBudget(budget.id);
      ondelete?.();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete budget';
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
  aria-label={isEditing ? $t('finance.budgets.editBudget') : $t('finance.budgets.addBudget')}
>
  <!-- Header -->
  <div class="flex items-center justify-between border-b border-[var(--color-border)] px-6 py-4">
    <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">
      {isEditing ? $t('finance.budgets.editBudget') : $t('finance.budgets.addBudget')}
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
    <!-- Name -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.budgets.name')}
      </span>
      <input
        type="text"
        bind:value={name}
        required
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Category -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.category')}
      </span>
      <select
        bind:value={categoryName}
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      >
        <option value="">{$t('finance.budgets.global')}</option>
        {#each categories as cat (cat.id)}
          <option value={cat.name}>{cat.name}</option>
        {/each}
      </select>
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

    <!-- Month (readonly) -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.date')}
      </span>
      <input
        type="text"
        value={month}
        readonly
        class="cursor-not-allowed rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-tertiary)] outline-none"
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
