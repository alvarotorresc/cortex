<script lang="ts">
  import { t } from 'svelte-i18n';
  import X from 'lucide-svelte/icons/x';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import { createGoal, updateGoal, deleteGoal } from '../api';
  import type { SavingsGoal, CreateGoalInput, UpdateGoalInput } from '../types';

  interface Props {
    goal?: SavingsGoal | null;
    onsave: () => void;
    oncancel: () => void;
    ondelete?: () => void;
  }

  const { goal = null, onsave, oncancel, ondelete }: Props = $props();

  const isEditing = $derived(goal !== null);

  // Form state — initialized from goal prop (form mounts fresh each time).
  // svelte-ignore state_referenced_locally
  let name = $state(goal?.name ?? '');
  // svelte-ignore state_referenced_locally
  let targetAmount = $state(goal?.target_amount ?? 0);
  // svelte-ignore state_referenced_locally
  let targetDate = $state(goal?.target_date?.split('T')[0] ?? '');
  // svelte-ignore state_referenced_locally
  let icon = $state(goal?.icon ?? '');
  // svelte-ignore state_referenced_locally
  let color = $state(goal?.color ?? '#0070F3');

  let saving = $state(false);
  let deleting = $state(false);
  let error = $state('');

  async function handleSubmit(): Promise<void> {
    if (saving || deleting) return;
    error = '';
    saving = true;

    try {
      const input: CreateGoalInput | UpdateGoalInput = {
        name,
        target_amount: targetAmount,
        target_date: targetDate || undefined,
        icon,
        color,
      };

      if (isEditing && goal) {
        await updateGoal(goal.id, input);
      } else {
        await createGoal(input);
      }
      onsave();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to save goal';
    } finally {
      saving = false;
    }
  }

  async function handleDelete(): Promise<void> {
    if (!isEditing || !goal || deleting || saving) return;

    const confirmed = window.confirm($t('finance.confirmDelete'));
    if (!confirmed) return;

    error = '';
    deleting = true;
    try {
      await deleteGoal(goal.id);
      ondelete?.();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete goal';
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
  aria-label={isEditing ? $t('finance.goals.editGoal') : $t('finance.goals.addGoal')}
>
  <!-- Header -->
  <div class="flex items-center justify-between border-b border-[var(--color-border)] px-6 py-4">
    <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">
      {isEditing ? $t('finance.goals.editGoal') : $t('finance.goals.addGoal')}
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
        {$t('finance.description')}
      </span>
      <input
        type="text"
        bind:value={name}
        required
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Target amount -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.amount')}
      </span>
      <input
        type="number"
        bind:value={targetAmount}
        min="0.01"
        step="0.01"
        required
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Target date -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.goals.targetDate')}
      </span>
      <input
        type="date"
        bind:value={targetDate}
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Icon -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('projectHub.icon')}
      </span>
      <input
        type="text"
        bind:value={icon}
        placeholder="e.g. 🏠 🚗 ✈️"
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Color -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('projectHub.color')}
      </span>
      <div class="flex items-center gap-3">
        <input
          type="color"
          bind:value={color}
          class="h-9 w-9 cursor-pointer rounded-[var(--radius-md)] border border-[var(--color-border)] bg-transparent p-0.5"
        />
        <input
          type="text"
          bind:value={color}
          pattern="^#[0-9A-Fa-f]{6}$"
          class="flex-1 rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
        />
      </div>
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
