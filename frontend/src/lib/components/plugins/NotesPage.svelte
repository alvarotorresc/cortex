<script lang="ts">
  import { t } from 'svelte-i18n';
  import { onMount } from 'svelte';
  import Plus from 'lucide-svelte/icons/plus';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import Pin from 'lucide-svelte/icons/pin';
  import PinOff from 'lucide-svelte/icons/pin-off';
  import NotebookPen from 'lucide-svelte/icons/notebook-pen';
  import FileText from 'lucide-svelte/icons/file-text';
  import ArrowUpDown from 'lucide-svelte/icons/arrow-up-down';
  import { pluginApi } from '$lib/api';

  type SortOption =
    | 'created-newest'
    | 'created-oldest'
    | 'updated-newest'
    | 'updated-oldest'
    | 'title-az';

  interface Note {
    id: string;
    title: string;
    content: string;
    pinned: boolean;
    created_at: string;
    updated_at: string;
  }

  const api = pluginApi('quick-notes');

  let notes = $state<Note[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let saving = $state(false);
  let isDirty = $state(false);

  let selectedNoteId = $state<string | null>(null);
  let editTitle = $state('');
  let editContent = $state('');
  let deleteConfirmId = $state<string | null>(null);
  let sortBy = $state<SortOption>('created-newest');

  function compareBySortOption(a: Note, b: Note): number {
    switch (sortBy) {
      case 'created-newest':
        return new Date(b.created_at).getTime() - new Date(a.created_at).getTime();
      case 'created-oldest':
        return new Date(a.created_at).getTime() - new Date(b.created_at).getTime();
      case 'updated-newest':
        return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
      case 'updated-oldest':
        return new Date(a.updated_at).getTime() - new Date(b.updated_at).getTime();
      case 'title-az':
        return a.title.localeCompare(b.title);
    }
  }

  const sortedNotes = $derived(
    [...notes].sort((a, b) => {
      if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
      return compareBySortOption(a, b);
    }),
  );

  const selectedNote = $derived(notes.find((n) => n.id === selectedNoteId) ?? null);

  async function loadNotes() {
    loading = true;
    error = null;

    try {
      const res = await api.fetch<{ data: Note[] }>('/notes');
      notes = res.data ?? [];
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load notes';
    } finally {
      loading = false;
    }
  }

  function selectNote(note: Note) {
    flushSave();
    selectedNoteId = note.id;
    editTitle = note.title;
    editContent = note.content;
    deleteConfirmId = null;
  }

  async function createNote() {
    saving = true;
    try {
      const res = await api.fetch<{ data: Note }>('/notes', {
        method: 'POST',
        body: JSON.stringify({ title: $t('notes.newNote'), content: '' }),
      });
      await loadNotes();
      if (res.data) {
        selectNote(res.data);
      }
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to create note';
    } finally {
      saving = false;
    }
  }

  async function saveNote() {
    if (!selectedNoteId) return;
    if (!isDirty) return;

    isDirty = false;
    saving = true;
    try {
      await api.fetch(`/notes/${selectedNoteId}`, {
        method: 'PUT',
        body: JSON.stringify({ title: editTitle, content: editContent }),
      });
      const now = new Date().toISOString();
      notes = notes.map((n) =>
        n.id === selectedNoteId
          ? { ...n, title: editTitle, content: editContent, updated_at: now }
          : n,
      );
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to save note';
    } finally {
      saving = false;
    }
  }

  async function togglePin(noteId: string) {
    try {
      await api.fetch(`/notes/${noteId}/pin`, { method: 'PUT' });
      await loadNotes();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to toggle pin';
    }
  }

  async function deleteNote(noteId: string) {
    try {
      await api.fetch(`/notes/${noteId}`, { method: 'DELETE' });
      if (selectedNoteId === noteId) {
        selectedNoteId = null;
        editTitle = '';
        editContent = '';
      }
      deleteConfirmId = null;
      await loadNotes();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete note';
    }
  }

  let saveTimeout: ReturnType<typeof setTimeout> | null = null;

  function handleEditorInput() {
    isDirty = true;
    if (saveTimeout) clearTimeout(saveTimeout);
    saveTimeout = setTimeout(() => {
      saveNote();
    }, 1000);
  }

  function flushSave() {
    if (saveTimeout) {
      clearTimeout(saveTimeout);
      saveTimeout = null;
    }
    if (isDirty) {
      saveNote();
    }
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString(undefined, {
      day: 'numeric',
      month: 'short',
      hour: '2-digit',
      minute: '2-digit',
    });
  }

  onMount(() => {
    loadNotes();
  });
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-3">
      <NotebookPen size={20} class="text-[var(--color-plugin-notes)]" />
      <h2 class="text-2xl font-semibold text-[var(--color-text-primary)]">
        {$t('notes.title')}
      </h2>
    </div>

    <button
      onclick={createNote}
      disabled={saving}
      class="flex items-center gap-2 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90 disabled:opacity-50"
    >
      <Plus size={16} />
      {$t('notes.newNote')}
    </button>
  </div>

  {#if loading}
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <div
        class="h-96 animate-pulse rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]"
      ></div>
      <div
        class="col-span-2 h-96 animate-pulse rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]"
      ></div>
    </div>
  {:else if error}
    <div
      class="rounded-[var(--radius-lg)] border border-[var(--color-error)]/20 bg-[var(--color-error)]/5 p-4"
    >
      <p class="text-sm text-[var(--color-error)]">{error}</p>
      <button
        onclick={loadNotes}
        class="mt-2 text-sm font-medium text-[var(--color-brand-blue)] hover:underline"
      >
        {$t('common.retry')}
      </button>
    </div>
  {:else if notes.length === 0}
    <div
      class="flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-8 py-16"
    >
      <FileText size={48} class="mb-4 text-[var(--color-text-tertiary)]" />
      <p class="text-sm text-[var(--color-text-secondary)]">{$t('notes.noNotes')}</p>
    </div>
  {:else}
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-3">
      <!-- Notes list -->
      <div class="flex flex-col gap-2">
        <!-- Sort dropdown -->
        <div class="flex items-center gap-2">
          <ArrowUpDown size={14} class="shrink-0 text-[var(--color-text-tertiary)]" />
          <label for="notes-sort" class="sr-only">{$t('notes.sortBy')}</label>
          <select
            id="notes-sort"
            bind:value={sortBy}
            class="w-full cursor-pointer rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-3 py-1.5 text-xs text-[var(--color-text-secondary)] transition-colors hover:border-[var(--color-text-tertiary)] focus:border-[var(--color-brand-blue)] focus:outline-none"
          >
            <option value="created-newest">{$t('notes.sortCreatedNewest')}</option>
            <option value="created-oldest">{$t('notes.sortCreatedOldest')}</option>
            <option value="updated-newest">{$t('notes.sortUpdatedNewest')}</option>
            <option value="updated-oldest">{$t('notes.sortUpdatedOldest')}</option>
            <option value="title-az">{$t('notes.sortTitleAz')}</option>
          </select>
        </div>

        <div
          class="divide-y divide-[var(--color-border)] overflow-y-auto rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] lg:max-h-[calc(100vh-280px)]"
        >
        {#each sortedNotes as note}
          <button
            onclick={() => selectNote(note)}
            class="flex w-full flex-col gap-1 px-4 py-3 text-left transition-colors {selectedNoteId ===
            note.id
              ? 'bg-[var(--color-bg-tertiary)]'
              : 'hover:bg-[var(--color-bg-tertiary)]'}"
          >
            <div class="flex items-center gap-2">
              <span class="flex-1 truncate text-sm font-medium text-[var(--color-text-primary)]">
                {note.title || $t('notes.newNote')}
              </span>
              {#if note.pinned}
                <span
                  class="shrink-0 rounded-[var(--radius-full)] bg-[var(--color-plugin-notes)]/10 px-2 py-0.5 text-xs font-medium text-[var(--color-plugin-notes)]"
                >
                  {$t('notes.pin')}
                </span>
              {/if}
            </div>
            <span class="truncate text-xs text-[var(--color-text-tertiary)]">
              {note.content
                ? note.content.slice(0, 80) + (note.content.length > 80 ? '...' : '')
                : ''}
            </span>
            <span class="text-xs text-[var(--color-text-tertiary)]">
              {formatDate(note.updated_at)}
            </span>
          </button>
        {/each}
        </div>
      </div>

      <!-- Editor -->
      <div class="lg:col-span-2">
        {#if selectedNote}
          <div
            class="flex h-full flex-col rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]"
          >
            <!-- Editor toolbar -->
            <div
              class="flex items-center justify-between border-b border-[var(--color-border)] px-4 py-3"
            >
              <div class="flex items-center gap-2">
                <button
                  onclick={() => togglePin(selectedNote.id)}
                  class="rounded-[var(--radius-sm)] p-1.5 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)] {selectedNote.pinned
                    ? 'text-[var(--color-plugin-notes)]'
                    : ''}"
                  title={selectedNote.pinned ? $t('notes.unpin') : $t('notes.pin')}
                >
                  {#if selectedNote.pinned}
                    <PinOff size={16} />
                  {:else}
                    <Pin size={16} />
                  {/if}
                </button>

                {#if deleteConfirmId === selectedNote.id}
                  <span class="text-xs text-[var(--color-error)]">
                    {$t('notes.delete')}?
                  </span>
                  <button
                    onclick={() => deleteNote(selectedNote.id)}
                    class="rounded-[var(--radius-sm)] bg-[var(--color-error)] px-2 py-1 text-xs font-medium text-white"
                  >
                    {$t('notes.delete')}
                  </button>
                  <button
                    onclick={() => (deleteConfirmId = null)}
                    class="text-xs text-[var(--color-text-secondary)] hover:underline"
                  >
                    {$t('dashboard.cancel')}
                  </button>
                {:else}
                  <button
                    onclick={() => (deleteConfirmId = selectedNote.id)}
                    class="rounded-[var(--radius-sm)] p-1.5 text-[var(--color-text-tertiary)] transition-colors hover:bg-[var(--color-error)]/10 hover:text-[var(--color-error)]"
                    title={$t('notes.delete')}
                  >
                    <Trash2 size={16} />
                  </button>
                {/if}
              </div>

              {#if saving}
                <span class="flex items-center gap-1.5 text-xs text-[var(--color-text-tertiary)]">
                  <span class="saving-dot inline-block h-1.5 w-1.5 rounded-[var(--radius-full)] bg-[var(--color-text-tertiary)]"></span>
                  {$t('notes.saving')}
                </span>
              {/if}
            </div>

            <!-- Title input -->
            <input
              type="text"
              bind:value={editTitle}
              oninput={handleEditorInput}
              onblur={flushSave}
              class="border-b border-[var(--color-border)] bg-transparent px-6 py-3 text-lg font-semibold text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:outline-none"
              placeholder={$t('notes.titleField')}
            />

            <!-- Content textarea -->
            <textarea
              bind:value={editContent}
              oninput={handleEditorInput}
              onblur={flushSave}
              class="min-h-[300px] flex-1 resize-none bg-transparent px-6 py-4 text-sm leading-relaxed text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:outline-none"
              placeholder={$t('notes.content')}
            ></textarea>
          </div>
        {:else}
          <div
            class="flex h-full min-h-[400px] items-center justify-center rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]"
          >
            <p class="text-sm text-[var(--color-text-tertiary)]">
              Select a note to edit
            </p>
          </div>
        {/if}
      </div>
    </div>
  {/if}
</div>

<style>
  .saving-dot {
    animation: saving-pulse 1s ease-in-out infinite;
  }

  @keyframes saving-pulse {
    0%,
    100% {
      opacity: 0.3;
    }
    50% {
      opacity: 1;
    }
  }
</style>
