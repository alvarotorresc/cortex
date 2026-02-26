<script lang="ts">
  import { t } from 'svelte-i18n';
  import Plus from 'lucide-svelte/icons/plus';
  import Pencil from 'lucide-svelte/icons/pencil';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import TagIcon from 'lucide-svelte/icons/tag';
  import Loader2 from 'lucide-svelte/icons/loader-2';
  import Check from 'lucide-svelte/icons/check';
  import X from 'lucide-svelte/icons/x';
  import { listTags, createTag, updateTag, deleteTag } from '../api';
  import type { Tag, CreateTagInput, UpdateTagInput } from '../types';
  import EmptyState from '../shared/EmptyState.svelte';

  // State
  let tags = $state<Tag[]>([]);
  let loading = $state(true);
  let error = $state('');

  // Form state
  let showForm = $state(false);
  let editingId = $state<number | null>(null);
  let formName = $state('');
  let formColor = $state('#3b82f6');
  let saving = $state(false);
  let formError = $state('');

  const DEFAULT_COLORS = [
    '#ef4444', '#f97316', '#eab308', '#22c55e', '#06b6d4',
    '#3b82f6', '#6366f1', '#8b5cf6', '#ec4899', '#64748b',
  ];

  $effect(() => {
    loadTags();
  });

  async function loadTags(): Promise<void> {
    loading = true;
    error = '';
    try {
      tags = await listTags();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load tags';
    } finally {
      loading = false;
    }
  }

  function openAddForm(): void {
    editingId = null;
    formName = '';
    formColor = '#3b82f6';
    formError = '';
    showForm = true;
  }

  function openEditForm(tag: Tag): void {
    editingId = tag.id;
    formName = tag.name;
    formColor = tag.color;
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
      const input: CreateTagInput | UpdateTagInput = {
        name: formName.trim(),
        color: formColor,
      };

      if (editingId !== null) {
        await updateTag(editingId, input);
      } else {
        await createTag(input);
      }

      showForm = false;
      editingId = null;
      await loadTags();
    } catch (err) {
      formError = err instanceof Error ? err.message : 'Failed to save tag';
    } finally {
      saving = false;
    }
  }

  async function handleDelete(tag: Tag): Promise<void> {
    const confirmed = window.confirm($t('finance.settingsPanel.confirmDeleteTag'));
    if (!confirmed) return;

    try {
      await deleteTag(tag.id);
      await loadTags();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete tag';
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
      <span class="hidden sm:inline">{$t('finance.settingsPanel.addTag')}</span>
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
        onclick={loadTags}
        class="mt-3 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
      >
        {$t('common.retry')}
      </button>
    </div>
  {:else if tags.length === 0}
    <EmptyState
      icon={TagIcon}
      message={$t('finance.settingsPanel.noTags')}
      actionLabel={$t('finance.settingsPanel.addTag')}
      onaction={openAddForm}
    />
  {:else}
    <div class="divide-y divide-[var(--color-border)] rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]">
      {#each tags as tag (tag.id)}
        {#if showForm && editingId === tag.id}
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
          <!-- Tag row -->
          <div class="flex items-center gap-3 px-4 py-3">
            <!-- Color dot -->
            <span
              class="h-3 w-3 shrink-0 rounded-full"
              style:background-color={tag.color}
              aria-hidden="true"
            ></span>

            <!-- Name -->
            <span class="min-w-0 flex-1 truncate text-sm font-medium text-[var(--color-text-primary)]">
              {tag.name}
            </span>

            <!-- Actions -->
            <div class="flex shrink-0 items-center gap-1">
              <button
                onclick={() => openEditForm(tag)}
                class="rounded-[var(--radius-sm)] p-1.5 text-[var(--color-text-tertiary)] transition-colors hover:bg-[var(--color-bg-tertiary)] hover:text-[var(--color-text-secondary)]"
                aria-label="Edit {tag.name}"
              >
                <Pencil size={14} />
              </button>
              <button
                onclick={() => handleDelete(tag)}
                class="rounded-[var(--radius-sm)] p-1.5 text-[var(--color-text-tertiary)] transition-colors hover:bg-[var(--color-error)]/10 hover:text-[var(--color-error)]"
                aria-label="Delete {tag.name}"
              >
                <Trash2 size={14} />
              </button>
            </div>
          </div>
        {/if}
      {/each}
    </div>
  {/if}
</div>
