<script lang="ts">
  import { t } from 'svelte-i18n';
  import Plus from 'lucide-svelte/icons/plus';
  import Pencil from 'lucide-svelte/icons/pencil';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import LayoutList from 'lucide-svelte/icons/layout-list';
  import Loader2 from 'lucide-svelte/icons/loader-2';
  import Check from 'lucide-svelte/icons/check';
  import X from 'lucide-svelte/icons/x';
  import { listCategories, createCategory, updateCategory, deleteCategory } from '../api';
  import type { Category, CategoryType, CreateCategoryInput, UpdateCategoryInput } from '../types';
  import EmptyState from '../shared/EmptyState.svelte';
  import LucideIcon from '../shared/LucideIcon.svelte';

  // State
  let categories = $state<Category[]>([]);
  let loading = $state(true);
  let error = $state('');

  // Form state
  let showForm = $state(false);
  let editingId = $state<number | null>(null);
  let formName = $state('');
  let formType = $state<CategoryType>('expense');
  let formIcon = $state('');
  let formColor = $state('#6366f1');
  let saving = $state(false);
  let formError = $state('');

  const DEFAULT_COLORS = [
    '#ef4444', '#f97316', '#eab308', '#22c55e', '#06b6d4',
    '#3b82f6', '#6366f1', '#8b5cf6', '#ec4899', '#64748b',
  ];

  $effect(() => {
    loadCategories();
  });

  async function loadCategories(): Promise<void> {
    loading = true;
    error = '';
    try {
      categories = await listCategories();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load categories';
    } finally {
      loading = false;
    }
  }

  function openAddForm(): void {
    editingId = null;
    formName = '';
    formType = 'expense';
    formIcon = '';
    formColor = '#6366f1';
    formError = '';
    showForm = true;
  }

  function openEditForm(cat: Category): void {
    editingId = cat.id;
    formName = cat.name;
    formType = cat.type;
    formIcon = cat.icon;
    formColor = cat.color;
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
      const input: CreateCategoryInput | UpdateCategoryInput = {
        name: formName.trim(),
        type: formType,
        icon: formIcon,
        color: formColor,
      };

      if (editingId !== null) {
        await updateCategory(editingId, input);
      } else {
        await createCategory(input);
      }

      showForm = false;
      editingId = null;
      await loadCategories();
    } catch (err) {
      formError = err instanceof Error ? err.message : 'Failed to save category';
    } finally {
      saving = false;
    }
  }

  async function handleDelete(cat: Category): Promise<void> {
    if (cat.is_default) return;
    const confirmed = window.confirm($t('finance.settingsPanel.confirmDeleteCategory'));
    if (!confirmed) return;

    try {
      await deleteCategory(cat.id);
      await loadCategories();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete category';
    }
  }

  function getTypeBadgeClasses(type: CategoryType): string {
    switch (type) {
      case 'income':
        return 'bg-[var(--color-success)]/10 text-[var(--color-success)]';
      case 'expense':
        return 'bg-[var(--color-error)]/10 text-[var(--color-error)]';
      case 'both':
        return 'bg-[var(--color-brand-blue)]/10 text-[var(--color-brand-blue)]';
    }
  }

  function getTypeLabel(type: CategoryType): string {
    switch (type) {
      case 'income':
        return $t('finance.income');
      case 'expense':
        return $t('finance.expense');
      case 'both':
        return $t('finance.settingsPanel.both');
    }
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
      <span class="hidden sm:inline">{$t('finance.settingsPanel.addCategory')}</span>
    </button>
  </div>

  <!-- Inline form (add mode, shown at top) -->
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
            <option value="expense">{$t('finance.expense')}</option>
            <option value="income">{$t('finance.income')}</option>
            <option value="both">{$t('finance.settingsPanel.both')}</option>
          </select>
        </label>

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
        onclick={loadCategories}
        class="mt-3 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
      >
        {$t('common.retry')}
      </button>
    </div>
  {:else if categories.length === 0}
    <EmptyState
      icon={LayoutList}
      message={$t('finance.settingsPanel.noCategories')}
      actionLabel={$t('finance.settingsPanel.addCategory')}
      onaction={openAddForm}
    />
  {:else}
    <div class="divide-y divide-[var(--color-border)] rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]">
      {#each categories as cat (cat.id)}
        {#if showForm && editingId === cat.id}
          <!-- Inline edit form replaces the row -->
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
                  <option value="expense">{$t('finance.expense')}</option>
                  <option value="income">{$t('finance.income')}</option>
                  <option value="both">{$t('finance.settingsPanel.both')}</option>
                </select>
              </label>

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
          <!-- Category row -->
          <div class="flex items-center gap-3 px-4 py-3">
            <!-- Icon -->
            {#if cat.icon}
              <span
                class="flex h-8 w-8 shrink-0 items-center justify-center rounded-[var(--radius-md)] text-sm"
                style:background-color="{cat.color}20"
                aria-hidden="true"
              >
                <LucideIcon name={cat.icon} />
              </span>
            {:else}
              <span
                class="h-3 w-3 shrink-0 rounded-full"
                style:background-color={cat.color}
                aria-hidden="true"
              ></span>
            {/if}

            <!-- Name -->
            <span class="min-w-0 flex-1 truncate text-sm font-medium text-[var(--color-text-primary)]">
              {cat.name}
            </span>

            <!-- Type badge -->
            <span class="shrink-0 rounded-full px-2 py-0.5 text-xs font-medium {getTypeBadgeClasses(cat.type)}">
              {getTypeLabel(cat.type)}
            </span>

            <!-- Color dot -->
            <span
              class="h-3 w-3 shrink-0 rounded-full"
              style:background-color={cat.color}
              aria-hidden="true"
            ></span>

            <!-- Actions -->
            <div class="flex shrink-0 items-center gap-1">
              <button
                onclick={() => openEditForm(cat)}
                class="rounded-[var(--radius-sm)] p-1.5 text-[var(--color-text-tertiary)] transition-colors hover:bg-[var(--color-bg-tertiary)] hover:text-[var(--color-text-secondary)]"
                aria-label="Edit {cat.name}"
              >
                <Pencil size={14} />
              </button>
              {#if !cat.is_default}
                <button
                  onclick={() => handleDelete(cat)}
                  class="rounded-[var(--radius-sm)] p-1.5 text-[var(--color-text-tertiary)] transition-colors hover:bg-[var(--color-error)]/10 hover:text-[var(--color-error)]"
                  aria-label="Delete {cat.name}"
                >
                  <Trash2 size={14} />
                </button>
              {/if}
            </div>
          </div>
        {/if}
      {/each}
    </div>
  {/if}
</div>
