<script lang="ts">
  import { t } from 'svelte-i18n';
  import Target from 'lucide-svelte/icons/target';
  import Plus from 'lucide-svelte/icons/plus';
  import Loader2 from 'lucide-svelte/icons/loader-2';
  import { listGoals, contributeToGoal } from '../api';
  import type { SavingsGoal } from '../types';
  import EmptyState from '../shared/EmptyState.svelte';
  import GoalCard from './GoalCard.svelte';
  import GoalForm from './GoalForm.svelte';

  // State
  let goals = $state<SavingsGoal[]>([]);
  let loading = $state(true);
  let error = $state('');
  let showForm = $state(false);
  let editingGoal = $state<SavingsGoal | null>(null);

  // Separate active and completed goals
  const activeGoals = $derived(goals.filter((g) => !g.is_completed));
  const completedGoals = $derived(goals.filter((g) => g.is_completed));

  // Load on mount
  $effect(() => {
    loadGoals();
  });

  async function loadGoals(): Promise<void> {
    loading = true;
    error = '';
    try {
      goals = await listGoals();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load goals';
    } finally {
      loading = false;
    }
  }

  function handleAdd(): void {
    editingGoal = null;
    showForm = true;
  }

  function handleEdit(goal: SavingsGoal): void {
    editingGoal = goal;
    showForm = true;
  }

  async function handleContribute(id: number, amount: number): Promise<void> {
    try {
      const updated = await contributeToGoal(id, amount);
      goals = goals.map((g) => (g.id === updated.id ? updated : g));
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to contribute';
    }
  }

  function handleSave(): void {
    showForm = false;
    editingGoal = null;
    loadGoals();
  }

  function handleDelete(): void {
    showForm = false;
    editingGoal = null;
    loadGoals();
  }

  function handleCancel(): void {
    showForm = false;
    editingGoal = null;
  }
</script>

<div class="space-y-4">
  <!-- Header with Add button -->
  <div class="flex items-center justify-end">
    <button
      onclick={handleAdd}
      class="flex items-center gap-1.5 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-3 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
    >
      <Plus size={16} />
      <span class="hidden sm:inline">{$t('finance.goals.addGoal')}</span>
    </button>
  </div>

  <!-- Content area -->
  {#if loading}
    <div class="flex items-center justify-center py-16">
      <Loader2 size={24} class="animate-spin text-[var(--color-text-tertiary)]" />
    </div>
  {:else if error}
    <div
      class="flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-6 py-12 text-center"
    >
      <p class="text-sm text-[var(--color-error)]">{error}</p>
      <button
        onclick={loadGoals}
        class="mt-3 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
      >
        {$t('common.retry')}
      </button>
    </div>
  {:else if goals.length === 0}
    <EmptyState
      icon={Target}
      message={$t('finance.goals.noGoals')}
      actionLabel={$t('finance.goals.addGoal')}
      onaction={handleAdd}
    />
  {:else}
    <!-- Active goals grid -->
    {#if activeGoals.length > 0}
      <div class="grid gap-4 sm:grid-cols-2">
        {#each activeGoals as goal (goal.id)}
          <GoalCard
            {goal}
            onedit={handleEdit}
            oncontribute={handleContribute}
          />
        {/each}
      </div>
    {/if}

    <!-- Completed goals -->
    {#if completedGoals.length > 0}
      <div class="space-y-3">
        <p class="text-xs font-medium uppercase tracking-wide text-[var(--color-text-tertiary)]">
          {$t('finance.goals.completed')}
        </p>
        <div class="grid gap-4 sm:grid-cols-2">
          {#each completedGoals as goal (goal.id)}
            <GoalCard
              {goal}
              onedit={handleEdit}
              oncontribute={handleContribute}
            />
          {/each}
        </div>
      </div>
    {/if}
  {/if}
</div>

<!-- Form modal -->
{#if showForm}
  <GoalForm
    goal={editingGoal}
    onsave={handleSave}
    oncancel={handleCancel}
    ondelete={handleDelete}
  />
{/if}
