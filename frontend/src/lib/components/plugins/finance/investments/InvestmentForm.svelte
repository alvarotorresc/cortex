<script lang="ts">
  import { t } from 'svelte-i18n';
  import X from 'lucide-svelte/icons/x';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import { createInvestment, updateInvestment, deleteInvestment } from '../api';
  import type {
    InvestmentWithPnL,
    InvestmentType,
    CreateInvestmentInput,
    UpdateInvestmentInput,
  } from '../types';

  interface Props {
    investment?: InvestmentWithPnL | null;
    onsave: () => void;
    oncancel: () => void;
    ondelete?: () => void;
  }

  const { investment = null, onsave, oncancel, ondelete }: Props = $props();

  const isEditing = $derived(investment !== null);

  const investmentTypes: InvestmentType[] = ['crypto', 'etf', 'fund', 'stock', 'other'];

  // Form state — initialized from investment prop (form mounts fresh each time).
  // svelte-ignore state_referenced_locally
  let name = $state(investment?.name ?? '');
  // svelte-ignore state_referenced_locally
  let type = $state<InvestmentType>(investment?.type ?? 'stock');
  // svelte-ignore state_referenced_locally
  let units = $state<number | undefined>(investment?.units ?? undefined);
  // svelte-ignore state_referenced_locally
  let avgBuyPrice = $state<number | undefined>(investment?.avg_buy_price ?? undefined);
  // svelte-ignore state_referenced_locally
  let currentPrice = $state<number | undefined>(investment?.current_price ?? undefined);
  // svelte-ignore state_referenced_locally
  let currency = $state(investment?.currency ?? 'EUR');
  // svelte-ignore state_referenced_locally
  let notes = $state(investment?.notes ?? '');
  // svelte-ignore state_referenced_locally
  let lastUpdated = $state(
    investment?.last_updated?.split('T')[0] ?? new Date().toISOString().split('T')[0],
  );

  let saving = $state(false);
  let deleting = $state(false);
  let error = $state('');

  async function handleSubmit(): Promise<void> {
    if (saving || deleting) return;
    error = '';
    saving = true;

    try {
      const input: CreateInvestmentInput | UpdateInvestmentInput = {
        name,
        type,
        units: units ?? undefined,
        avg_buy_price: avgBuyPrice ?? undefined,
        current_price: currentPrice ?? undefined,
        currency,
        notes,
        last_updated: lastUpdated ? `${lastUpdated}T00:00:00Z` : undefined,
      };

      if (isEditing && investment) {
        await updateInvestment(investment.id, input);
      } else {
        await createInvestment(input);
      }
      onsave();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to save investment';
    } finally {
      saving = false;
    }
  }

  async function handleDelete(): Promise<void> {
    if (!isEditing || !investment || deleting || saving) return;

    const confirmed = window.confirm($t('finance.investments.confirmDelete'));
    if (!confirmed) return;

    error = '';
    deleting = true;
    try {
      await deleteInvestment(investment.id);
      ondelete?.();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete investment';
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
  aria-label={isEditing ? $t('finance.investments.editInvestment') : $t('finance.investments.addInvestment')}
>
  <!-- Header -->
  <div class="flex items-center justify-between border-b border-[var(--color-border)] px-6 py-4">
    <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">
      {isEditing ? $t('finance.investments.editInvestment') : $t('finance.investments.addInvestment')}
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
        {$t('finance.investments.name')}
      </span>
      <input
        type="text"
        bind:value={name}
        required
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Type selector -->
    <fieldset>
      <legend class="mb-1.5 text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.investments.type')}
      </legend>
      <div class="flex flex-wrap gap-1 rounded-[var(--radius-md)] border border-[var(--color-border)] p-1">
        {#each investmentTypes as invType (invType)}
          <button
            type="button"
            onclick={() => (type = invType)}
            class="flex-1 rounded-[var(--radius-sm)] px-3 py-1.5 text-sm font-medium transition-colors {type ===
            invType
              ? 'bg-[var(--color-brand-blue)] text-white'
              : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]'}"
          >
            {$t(`finance.investments.${invType}`)}
          </button>
        {/each}
      </div>
    </fieldset>

    <!-- Units -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.investments.units')}
      </span>
      <input
        type="number"
        bind:value={units}
        min="0"
        step="any"
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Avg Buy Price -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.investments.avgPrice')}
      </span>
      <input
        type="number"
        bind:value={avgBuyPrice}
        min="0"
        step="0.01"
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Current Price -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.investments.currentPrice')}
      </span>
      <input
        type="number"
        bind:value={currentPrice}
        min="0"
        step="0.01"
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Currency -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.investments.currency')}
      </span>
      <input
        type="text"
        bind:value={currency}
        required
        maxlength="3"
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm uppercase text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      />
    </label>

    <!-- Notes -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.investments.notes')}
      </span>
      <textarea
        bind:value={notes}
        rows="3"
        class="rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-2 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
      ></textarea>
    </label>

    <!-- Last Updated -->
    <label class="flex flex-col gap-1.5">
      <span class="text-sm font-medium text-[var(--color-text-secondary)]">
        {$t('finance.investments.lastUpdated')}
      </span>
      <input
        type="date"
        bind:value={lastUpdated}
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
