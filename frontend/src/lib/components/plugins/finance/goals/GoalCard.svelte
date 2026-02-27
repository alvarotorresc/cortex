<script lang="ts">
  import { t } from 'svelte-i18n';
  import Check from 'lucide-svelte/icons/check';
  import Plus from 'lucide-svelte/icons/plus';
  import CalendarDays from 'lucide-svelte/icons/calendar-days';
  import type { SavingsGoal } from '../types';
  import ProgressBar from '../shared/ProgressBar.svelte';
  import AmountDisplay from '../shared/AmountDisplay.svelte';

  interface Props {
    goal: SavingsGoal;
    onedit: (goal: SavingsGoal) => void;
    oncontribute: (id: number, amount: number) => void;
  }

  const { goal, onedit, oncontribute }: Props = $props();

  let showContributeInput = $state(false);
  let contributeAmount = $state(0);
  let contributing = $state(false);

  const percentage = $derived(
    goal.target_amount > 0
      ? (goal.current_amount / goal.target_amount) * 100
      : 0,
  );

  const daysRemaining = $derived(() => {
    if (!goal.target_date) return null;
    const target = new Date(goal.target_date);
    const now = new Date();
    const diff = Math.ceil((target.getTime() - now.getTime()) / (1000 * 60 * 60 * 24));
    return diff;
  });

  const formattedTargetDate = $derived(() => {
    if (!goal.target_date) return null;
    return new Date(goal.target_date).toLocaleDateString(undefined, {
      month: 'short',
      day: 'numeric',
      year: 'numeric',
    });
  });

  function handleCardClick(): void {
    if (!showContributeInput) {
      onedit(goal);
    }
  }

  function handleContributeToggle(e: MouseEvent): void {
    e.stopPropagation();
    showContributeInput = !showContributeInput;
    contributeAmount = 0;
  }

  async function handleContributeSubmit(e: MouseEvent): Promise<void> {
    e.stopPropagation();
    if (contributing || contributeAmount <= 0) return;
    contributing = true;
    try {
      oncontribute(goal.id, contributeAmount);
      showContributeInput = false;
      contributeAmount = 0;
    } finally {
      contributing = false;
    }
  }

  function handleContributeCancel(e: MouseEvent): void {
    e.stopPropagation();
    showContributeInput = false;
    contributeAmount = 0;
  }

  function handleInputClick(e: Event): void {
    e.stopPropagation();
  }

  function handleCardKeydown(e: KeyboardEvent): void {
    if (e.key === 'Enter' || e.key === ' ') {
      e.preventDefault();
      handleCardClick();
    }
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div
  role="button"
  tabindex="0"
  onclick={handleCardClick}
  onkeydown={handleCardKeydown}
  class="flex w-full cursor-pointer flex-col gap-3 rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-4 text-left transition-colors hover:bg-[var(--color-bg-tertiary)] {goal.is_completed ? 'opacity-60' : ''}"
>
  <!-- Header: Icon + Name + Completed badge -->
  <div class="flex items-center gap-3">
    <span
      class="flex h-9 w-9 shrink-0 items-center justify-center rounded-[var(--radius-md)] text-base"
      style:background-color="{goal.color}20"
      aria-hidden="true"
    >
      {goal.icon}
    </span>

    <div class="min-w-0 flex-1">
      <p class="truncate text-sm font-semibold text-[var(--color-text-primary)] {goal.is_completed ? 'line-through' : ''}">
        {goal.name}
      </p>
      {#if goal.target_date}
        <div class="flex items-center gap-1 text-xs text-[var(--color-text-tertiary)]">
          <CalendarDays size={12} />
          <span>{formattedTargetDate()}</span>
          {#if !goal.is_completed}
            {@const days = daysRemaining()}
            {#if days !== null}
              <span class="ml-1 {days < 0 ? 'text-[var(--color-error)]' : ''}">
                ({days > 0 ? `${days} ${$t('finance.goals.daysLeft')}` : $t('finance.goals.completed')})
              </span>
            {/if}
          {/if}
        </div>
      {/if}
    </div>

    {#if goal.is_completed}
      <span class="flex h-6 w-6 shrink-0 items-center justify-center rounded-full bg-[var(--color-success)]" aria-label={$t('finance.goals.completed')}>
        <Check size={14} class="text-white" />
      </span>
    {/if}
  </div>

  <!-- Progress bar -->
  <ProgressBar percentage={percentage} />

  <!-- Amounts -->
  <div class="flex items-center justify-between">
    <span class="text-sm font-medium">
      <AmountDisplay amount={goal.current_amount} />
    </span>
    <span class="text-xs text-[var(--color-text-tertiary)]">
      / <AmountDisplay amount={goal.target_amount} />
    </span>
  </div>

  <!-- Contribute section -->
  {#if !goal.is_completed}
    {#if showContributeInput}
      <!-- svelte-ignore a11y_no_static_element_interactions -->
      <div
        class="flex items-center gap-2 border-t border-[var(--color-border)] pt-3"
        onclick={handleInputClick}
        onkeydown={handleInputClick}
      >
        <input
          type="number"
          bind:value={contributeAmount}
          min="0.01"
          step="0.01"
          placeholder={$t('finance.amount')}
          class="min-w-0 flex-1 rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-sm text-[var(--color-text-primary)] outline-none transition-colors focus:border-[var(--color-brand-blue)]"
          aria-label={$t('finance.amount')}
        />
        <button
          type="button"
          onclick={handleContributeSubmit}
          disabled={contributing || contributeAmount <= 0}
          class="rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-3 py-1.5 text-sm font-medium text-white transition-colors hover:opacity-90 disabled:opacity-50"
        >
          {$t('finance.save')}
        </button>
        <button
          type="button"
          onclick={handleContributeCancel}
          class="rounded-[var(--radius-md)] px-3 py-1.5 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
        >
          {$t('finance.cancel')}
        </button>
      </div>
    {:else}
      <div class="border-t border-[var(--color-border)] pt-3">
        <button
          type="button"
          onclick={handleContributeToggle}
          class="flex items-center gap-1.5 rounded-[var(--radius-md)] px-3 py-1.5 text-sm font-medium text-[var(--color-brand-blue)] transition-colors hover:bg-[var(--color-brand-blue)]/10"
        >
          <Plus size={14} />
          {$t('finance.goals.contribute')}
        </button>
      </div>
    {/if}
  {/if}
</div>
