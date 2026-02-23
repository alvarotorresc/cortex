<script lang="ts">
  import { t } from 'svelte-i18n';
  import { onMount } from 'svelte';
  import Plus from 'lucide-svelte/icons/plus';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import Pin from 'lucide-svelte/icons/pin';
  import PinOff from 'lucide-svelte/icons/pin-off';
  import StickyNote from 'lucide-svelte/icons/sticky-note';
  import FileText from 'lucide-svelte/icons/file-text';
  import { pluginApi } from '$lib/api';

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

  let selectedNoteId = $state<string | null>(null);
  let editTitle = $state('');
  let editContent = $state('');
  let deleteConfirmId = $state<string | null>(null);

  const sortedNotes = $derived(
    [...notes].sort((a, b) => {
      if (a.pinned !== b.pinned) return a.pinned ? -1 : 1;
      return new Date(b.updated_at).getTime() - new Date(a.updated_at).getTime();
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

    saving = true;
    try {
      await api.fetch(`/notes/${selectedNoteId}`, {
        method: 'PUT',
        body: JSON.stringify({ title: editTitle, content: editContent }),
      });
      await loadNotes();
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
    if (saveTimeout) clearTimeout(saveTimeout);
    saveTimeout = setTimeout(() => {
      saveNote();
    }, 1000);
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
      <StickyNote size={20} class="text-[var(--color-plugin-notes)]" />
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
      <div
        class="divide-y divide-[var(--color-border)] overflow-y-auto rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] lg:max-h-[calc(100vh-240px)]"
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
                <span class="text-xs text-[var(--color-text-tertiary)]">
                  {$t('common.loading')}
                </span>
              {/if}
            </div>

            <!-- Title input -->
            <input
              type="text"
              bind:value={editTitle}
              oninput={handleEditorInput}
              class="border-b border-[var(--color-border)] bg-transparent px-6 py-3 text-lg font-semibold text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:outline-none"
              placeholder={$t('notes.titleField')}
            />

            <!-- Content textarea -->
            <textarea
              bind:value={editContent}
              oninput={handleEditorInput}
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
