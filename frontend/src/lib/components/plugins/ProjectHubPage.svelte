<script lang="ts">
  import { t } from 'svelte-i18n';
  import FolderGit2 from 'lucide-svelte/icons/folder-git-2';
  import Plus from 'lucide-svelte/icons/plus';
  import ArrowLeft from 'lucide-svelte/icons/arrow-left';
  import Pencil from 'lucide-svelte/icons/pencil';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import X from 'lucide-svelte/icons/x';
  import ExternalLink from 'lucide-svelte/icons/external-link';
  import Github from 'lucide-svelte/icons/github';
  import Globe from 'lucide-svelte/icons/globe';
  import BookOpen from 'lucide-svelte/icons/book-open';
  import { pluginApi } from '$lib/api';

  interface Tag {
    id: number;
    name: string;
    color: string;
  }

  interface Project {
    id: number;
    name: string;
    slug: string;
    tagline: string;
    status: string;
    category: string;
    version: string | null;
    stack: string;
    icon: string;
    color: string;
    repo_url: string | null;
    web_url: string | null;
    docs_url: string | null;
    hosting: string | null;
    notes: string | null;
    sort_order: number;
    created_at: string;
    updated_at: string;
    tags: Tag[];
  }

  interface ProjectLink {
    id: number;
    project_id: number;
    label: string;
    url: string;
    sort_order: number;
  }

  interface ProjectWithTags extends Project {
    links: ProjectLink[];
  }

  const api = pluginApi('project-hub');

  // State
  let projects = $state<Project[]>([]);
  let selectedProject = $state<ProjectWithTags | null>(null);
  let allTags = $state<Tag[]>([]);
  let loading = $state(true);
  let error = $state<string | null>(null);
  let submitting = $state(false);

  // View state: 'list' | 'detail'
  let view = $state<'list' | 'detail'>('list');

  // Filters
  let filterStatus = $state('');
  let filterCategory = $state('');
  let filterTag = $state('');
  let searchQuery = $state('');

  // Modal state
  let showModal = $state(false);
  let editingSlug = $state<string | null>(null);

  // Delete confirmation
  let showDeleteConfirm = $state(false);

  // Form state
  let formName = $state('');
  let formTagline = $state('');
  let formStatus = $state('concept');
  let formCategory = $state('lab');
  let formVersion = $state('');
  let formStack = $state('');
  let formTagIds = $state<number[]>([]);
  let formIcon = $state('folder');
  let formColor = $state('#0070F3');
  let formHosting = $state('');
  let formRepoUrl = $state('');
  let formWebUrl = $state('');
  let formDocsUrl = $state('');
  let formNotes = $state('');
  let formSortOrder = $state(0);
  let formCustomLinks = $state<{ label: string; url: string }[]>([]);

  // Derived
  const flagshipProjects = $derived(
    projects.filter((p) => p.category === 'flagship'),
  );
  const labProjects = $derived(
    projects.filter((p) => p.category === 'lab'),
  );
  const filteredCount = $derived(projects.length);

  const statusConfig: Record<string, { label: string; color: string }> = {
    concept: { label: 'projectHub.statusConcept', color: 'var(--color-text-tertiary)' },
    design: { label: 'projectHub.statusDesign', color: '#6366F1' },
    development: { label: 'projectHub.statusDevelopment', color: '#0070F3' },
    active: { label: 'projectHub.statusActive', color: '#16A34A' },
    maintenance: { label: 'projectHub.statusMaintenance', color: '#D97706' },
    archived: { label: 'projectHub.statusArchived', color: 'var(--color-text-tertiary)' },
    absorbed: { label: 'projectHub.statusAbsorbed', color: 'var(--color-text-tertiary)' },
  };

  const iconOptions = [
    'folder', 'code', 'globe', 'smartphone', 'terminal', 'database', 'server',
    'rocket', 'wrench', 'gamepad-2', 'music', 'book-open', 'shopping-cart', 'heart',
    'shield', 'users', 'newspaper', 'brain', 'flame', 'paw-print', 'tag', 'wallet',
    'dumbbell', 'clipboard', 'settings', 'zap',
  ];

  // --- Data loading ---

  async function loadTags() {
    try {
      const res = await api.fetch<{ data: Tag[] }>('/tags');
      allTags = res.data ?? [];
    } catch {
      // Tags are non-critical; silently fallback to empty
      allTags = [];
    }
  }

  async function loadProjects() {
    loading = true;
    error = null;

    try {
      const params = new URLSearchParams();
      if (filterStatus) params.set('status', filterStatus);
      if (filterCategory) params.set('category', filterCategory);
      if (filterTag) params.set('tag', filterTag);
      if (searchQuery) params.set('search', searchQuery);

      const queryStr = params.toString();
      const path = queryStr ? `/projects?${queryStr}` : '/projects';
      const res = await api.fetch<{ data: Project[] }>(path);
      projects = res.data ?? [];
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load projects';
    } finally {
      loading = false;
    }
  }

  async function loadProjectDetail(slug: string) {
    try {
      const res = await api.fetch<{ data: ProjectWithTags }>(`/projects/${slug}`);
      selectedProject = res.data;
      view = 'detail';
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to load project';
    }
  }

  // --- CRUD ---

  function openCreateModal() {
    editingSlug = null;
    formName = '';
    formTagline = '';
    formStatus = 'concept';
    formCategory = 'lab';
    formVersion = '';
    formStack = '';
    formTagIds = [];
    formIcon = 'folder';
    formColor = '#0070F3';
    formHosting = '';
    formRepoUrl = '';
    formWebUrl = '';
    formDocsUrl = '';
    formNotes = '';
    formSortOrder = 0;
    formCustomLinks = [];
    showModal = true;
  }

  function openEditModal() {
    if (!selectedProject) return;
    editingSlug = selectedProject.slug;
    formName = selectedProject.name;
    formTagline = selectedProject.tagline;
    formStatus = selectedProject.status;
    formCategory = selectedProject.category;
    formVersion = selectedProject.version ?? '';
    formStack = selectedProject.stack;
    formTagIds = (selectedProject.tags ?? []).map((t) => t.id);
    formIcon = selectedProject.icon;
    formColor = selectedProject.color;
    formHosting = selectedProject.hosting ?? '';
    formRepoUrl = selectedProject.repo_url ?? '';
    formWebUrl = selectedProject.web_url ?? '';
    formDocsUrl = selectedProject.docs_url ?? '';
    formNotes = selectedProject.notes ?? '';
    formSortOrder = selectedProject.sort_order;
    formCustomLinks = selectedProject.links.map((l) => ({ label: l.label, url: l.url }));
    showModal = true;
  }

  function toggleFormTag(tagId: number) {
    if (formTagIds.includes(tagId)) {
      formTagIds = formTagIds.filter((id) => id !== tagId);
    } else {
      formTagIds = [...formTagIds, tagId];
    }
  }

  async function saveProject() {
    if (!formName.trim() || !formTagline.trim() || formTagIds.length === 0) return;

    submitting = true;
    error = null;

    // Derive stack from selected tags for backward compat
    const derivedStack =
      allTags
        .filter((t) => formTagIds.includes(t.id))
        .map((t) => t.name)
        .join(', ') || 'TBD';

    try {
      const payload: Record<string, unknown> = {
        name: formName.trim(),
        tagline: formTagline.trim(),
        status: formStatus,
        category: formCategory,
        version: formVersion.trim() || null,
        stack: derivedStack,
        icon: formIcon,
        color: formColor,
        repo_url: formRepoUrl.trim() || null,
        web_url: formWebUrl.trim() || null,
        docs_url: formDocsUrl.trim() || null,
        hosting: formHosting.trim() || null,
        notes: formNotes.trim() || null,
        sort_order: formSortOrder,
        tag_ids: formTagIds,
      };

      if (editingSlug) {
        // Update
        await api.fetch(`/projects/${editingSlug}`, {
          method: 'PUT',
          body: JSON.stringify(payload),
        });

        // Set tags
        await api.fetch(`/projects/${editingSlug}/tags`, {
          method: 'POST',
          body: JSON.stringify({ tag_ids: formTagIds }),
        });

        // Handle custom links: delete existing, recreate
        if (selectedProject) {
          for (const link of selectedProject.links) {
            await api.fetch(`/links/${link.id}`, { method: 'DELETE' });
          }
        }
        for (const link of formCustomLinks) {
          if (link.label.trim() && link.url.trim()) {
            await api.fetch(`/projects/${editingSlug}/links`, {
              method: 'POST',
              body: JSON.stringify({ label: link.label.trim(), url: link.url.trim() }),
            });
          }
        }

        // Reload detail
        await loadProjectDetail(editingSlug);
      } else {
        // Create
        const res = await api.fetch<{ data: { id: number; slug: string } }>('/projects', {
          method: 'POST',
          body: JSON.stringify(payload),
        });

        // Set tags
        const slug = res.data.slug;
        await api.fetch(`/projects/${slug}/tags`, {
          method: 'POST',
          body: JSON.stringify({ tag_ids: formTagIds }),
        });

        // Add custom links
        for (const link of formCustomLinks) {
          if (link.label.trim() && link.url.trim()) {
            await api.fetch(`/projects/${slug}/links`, {
              method: 'POST',
              body: JSON.stringify({ label: link.label.trim(), url: link.url.trim() }),
            });
          }
        }
      }

      showModal = false;
      await loadProjects();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to save project';
    } finally {
      submitting = false;
    }
  }

  async function deleteProject() {
    if (!selectedProject) return;

    submitting = true;
    try {
      await api.fetch(`/projects/${selectedProject.slug}`, { method: 'DELETE' });
      showDeleteConfirm = false;
      selectedProject = null;
      view = 'list';
      await loadProjects();
    } catch (err) {
      error = err instanceof Error ? err.message : 'Failed to delete project';
    } finally {
      submitting = false;
    }
  }

  function addCustomLink() {
    formCustomLinks = [...formCustomLinks, { label: '', url: '' }];
  }

  function removeCustomLink(index: number) {
    formCustomLinks = formCustomLinks.filter((_, i) => i !== index);
  }

  function goBackToList() {
    view = 'list';
    selectedProject = null;
  }

  function formatDate(dateStr: string): string {
    return new Date(dateStr).toLocaleDateString(undefined, {
      day: 'numeric',
      month: 'short',
      year: 'numeric',
    });
  }

  $effect(() => {
    // Load tags once on mount.
    loadTags();
  });

  $effect(() => {
    // Re-fetch when filters change (also runs on mount).
    filterStatus;
    filterCategory;
    filterTag;
    searchQuery;
    loadProjects();
  });
</script>

<div class="space-y-6">
  {#if view === 'list'}
    <!-- ============ LIST VIEW ============ -->

    <!-- Header -->
    <div class="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
      <div class="flex items-center gap-3">
        <FolderGit2 size={20} style="color: #8B5CF6" />
        <h2 class="text-2xl font-semibold text-[var(--color-text-primary)]">
          {$t('projectHub.title')}
        </h2>
      </div>

      <button
        onclick={openCreateModal}
        class="flex items-center gap-2 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
      >
        <Plus size={16} />
        {$t('projectHub.addProject')}
      </button>
    </div>

    <!-- Filters -->
    <div class="flex flex-col gap-3 sm:flex-row sm:items-center">
      <select
        bind:value={filterStatus}
        class="rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
      >
        <option value="">{$t('projectHub.allStatuses')}</option>
        {#each Object.entries(statusConfig) as [value, config]}
          <option {value}>{$t(config.label)}</option>
        {/each}
      </select>

      <select
        bind:value={filterCategory}
        class="rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
      >
        <option value="">{$t('projectHub.allCategories')}</option>
        <option value="flagship">{$t('projectHub.categoryFlagship')}</option>
        <option value="lab">{$t('projectHub.categoryLab')}</option>
      </select>

      <input
        type="text"
        bind:value={searchQuery}
        placeholder={$t('projectHub.search')}
        class="flex-1 rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
      />

      {#if filterTag}
        <button
          onclick={() => { filterTag = ''; }}
          class="flex items-center gap-1 rounded-[var(--radius-full)] bg-[var(--color-brand-blue)]/10 px-3 py-1.5 text-xs font-medium text-[var(--color-brand-blue)]"
        >
          {filterTag}
          <X size={12} />
        </button>
      {/if}

      <span class="text-xs text-[var(--color-text-tertiary)]">
        {filteredCount} {$t('projectHub.projects')}
      </span>
    </div>

    <!-- Content -->
    {#if loading}
      <div class="space-y-4">
        {#each [1, 2, 3] as _}
          <div
            class="h-16 animate-pulse rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]"
          ></div>
        {/each}
      </div>
    {:else if error}
      <div
        class="rounded-[var(--radius-lg)] border border-[var(--color-error)]/20 bg-[var(--color-error)]/5 p-4"
      >
        <p class="text-sm text-[var(--color-error)]">{error}</p>
        <button
          onclick={loadProjects}
          class="mt-2 text-sm font-medium text-[var(--color-brand-blue)] hover:underline"
        >
          {$t('common.retry')}
        </button>
      </div>
    {:else if projects.length === 0}
      <div
        class="rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-6 py-12 text-center"
      >
        <p class="text-sm text-[var(--color-text-tertiary)]">
          {filterStatus || filterCategory || filterTag || searchQuery
            ? $t('projectHub.noResults')
            : $t('projectHub.noProjects')}
        </p>
      </div>
    {:else}
      <!-- Flagship section -->
      {#if flagshipProjects.length > 0}
        <div class="space-y-2">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-tertiary)]">
            {$t('projectHub.categoryFlagship')} ({flagshipProjects.length})
          </h3>
          <div
            class="divide-y divide-[var(--color-border)] rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]"
          >
            {#each flagshipProjects as project}
              <button
                onclick={() => loadProjectDetail(project.slug)}
                class="flex w-full items-center gap-4 px-6 py-4 text-left transition-colors hover:bg-[var(--color-bg-tertiary)]"
              >
                <span
                  class="inline-block h-3 w-3 shrink-0 rounded-[var(--radius-full)]"
                  style="background-color: {project.color}"
                ></span>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="truncate font-medium text-[var(--color-text-primary)]">
                      {project.name}
                    </span>
                    {#if project.version}
                      <span
                        class="shrink-0 rounded-[var(--radius-full)] bg-[var(--color-bg-tertiary)] px-2 py-0.5 text-xs font-medium text-[var(--color-text-secondary)]"
                      >
                        {project.version}
                      </span>
                    {/if}
                  </div>
                  <div class="mt-0.5 flex flex-wrap gap-1">
                    {#each project.tags ?? [] as tag}
                      <button
                        onclick={(e) => { e.stopPropagation(); filterTag = tag.name; }}
                        class="rounded-[var(--radius-full)] px-2 py-0.5 text-xs font-medium transition-opacity hover:opacity-80"
                        style="background-color: {tag.color}20; color: {tag.color}; border: 1px solid {tag.color}40"
                      >
                        {tag.name}
                      </button>
                    {/each}
                  </div>
                </div>
                <span
                  class="shrink-0 rounded-[var(--radius-full)] px-2.5 py-1 text-xs font-medium"
                  style="background-color: {statusConfig[project.status]?.color}20; color: {statusConfig[project.status]?.color}"
                >
                  {$t(statusConfig[project.status]?.label ?? 'projectHub.statusConcept')}
                </span>
                {#if project.web_url}
                  <span class="shrink-0 text-xs text-[var(--color-text-tertiary)]">
                    {new URL(project.web_url).hostname}
                  </span>
                {/if}
              </button>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Lab section -->
      {#if labProjects.length > 0}
        <div class="space-y-2">
          <h3 class="text-xs font-semibold uppercase tracking-wider text-[var(--color-text-tertiary)]">
            {$t('projectHub.categoryLab')} ({labProjects.length})
          </h3>
          <div
            class="divide-y divide-[var(--color-border)] rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]"
          >
            {#each labProjects as project}
              <button
                onclick={() => loadProjectDetail(project.slug)}
                class="flex w-full items-center gap-4 px-6 py-4 text-left transition-colors hover:bg-[var(--color-bg-tertiary)]"
              >
                <span
                  class="inline-block h-3 w-3 shrink-0 rounded-[var(--radius-full)]"
                  style="background-color: {project.color}"
                ></span>
                <div class="min-w-0 flex-1">
                  <div class="flex items-center gap-2">
                    <span class="truncate font-medium text-[var(--color-text-primary)]"
                      class:line-through={project.status === 'absorbed'}
                      class:opacity-50={project.status === 'absorbed'}
                    >
                      {project.name}
                    </span>
                    {#if project.version}
                      <span
                        class="shrink-0 rounded-[var(--radius-full)] bg-[var(--color-bg-tertiary)] px-2 py-0.5 text-xs font-medium text-[var(--color-text-secondary)]"
                      >
                        {project.version}
                      </span>
                    {/if}
                  </div>
                  <div class="mt-0.5 flex flex-wrap gap-1">
                    {#each project.tags ?? [] as tag}
                      <button
                        onclick={(e) => { e.stopPropagation(); filterTag = tag.name; }}
                        class="rounded-[var(--radius-full)] px-2 py-0.5 text-xs font-medium transition-opacity hover:opacity-80"
                        style="background-color: {tag.color}20; color: {tag.color}; border: 1px solid {tag.color}40"
                      >
                        {tag.name}
                      </button>
                    {/each}
                  </div>
                </div>
                <span
                  class="shrink-0 rounded-[var(--radius-full)] px-2.5 py-1 text-xs font-medium"
                  style="background-color: {statusConfig[project.status]?.color}20; color: {statusConfig[project.status]?.color}"
                >
                  {$t(statusConfig[project.status]?.label ?? 'projectHub.statusConcept')}
                </span>
                {#if project.web_url}
                  <span class="shrink-0 text-xs text-[var(--color-text-tertiary)]">
                    {new URL(project.web_url).hostname}
                  </span>
                {/if}
              </button>
            {/each}
          </div>
        </div>
      {/if}
    {/if}

  {:else if view === 'detail' && selectedProject}
    <!-- ============ DETAIL VIEW ============ -->

    <!-- Header -->
    <div class="flex items-center justify-between">
      <button
        onclick={goBackToList}
        class="flex items-center gap-2 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:text-[var(--color-text-primary)]"
      >
        <ArrowLeft size={16} />
        {$t('projectHub.back')}
      </button>

      <div class="flex items-center gap-2">
        <button
          onclick={openEditModal}
          class="flex items-center gap-2 rounded-[var(--radius-md)] border border-[var(--color-border)] px-3 py-1.5 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
        >
          <Pencil size={14} />
          {$t('projectHub.edit')}
        </button>
        <button
          onclick={() => (showDeleteConfirm = true)}
          class="flex items-center gap-2 rounded-[var(--radius-md)] border border-[var(--color-error)]/20 px-3 py-1.5 text-sm font-medium text-[var(--color-error)] transition-colors hover:bg-[var(--color-error)]/5"
        >
          <Trash2 size={14} />
          {$t('projectHub.delete')}
        </button>
      </div>
    </div>

    <!-- Project detail card -->
    <div
      class="rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] p-8"
    >
      <!-- Title row -->
      <div class="flex items-start gap-4">
        <span
          class="mt-1 inline-block h-4 w-4 shrink-0 rounded-[var(--radius-full)]"
          style="background-color: {selectedProject.color}"
        ></span>
        <div class="flex-1">
          <div class="flex items-center gap-3">
            <h2 class="text-2xl font-bold text-[var(--color-text-primary)]"
              class:line-through={selectedProject.status === 'absorbed'}
            >
              {selectedProject.name}
            </h2>
            {#if selectedProject.version}
              <span
                class="rounded-[var(--radius-full)] bg-[var(--color-bg-tertiary)] px-3 py-1 text-sm font-medium text-[var(--color-text-secondary)]"
              >
                {selectedProject.version}
              </span>
            {/if}
          </div>
          <p class="mt-1 text-[var(--color-text-secondary)]">{selectedProject.tagline}</p>
        </div>
      </div>

      <!-- Status & Category -->
      <div class="mt-6 flex flex-wrap gap-4">
        <div>
          <span class="text-xs font-medium text-[var(--color-text-tertiary)]">
            {$t('projectHub.status')}
          </span>
          <div class="mt-1">
            <span
              class="rounded-[var(--radius-full)] px-3 py-1 text-sm font-medium"
              style="background-color: {statusConfig[selectedProject.status]?.color}20; color: {statusConfig[selectedProject.status]?.color}"
            >
              {$t(statusConfig[selectedProject.status]?.label ?? 'projectHub.statusConcept')}
            </span>
          </div>
        </div>
        <div>
          <span class="text-xs font-medium text-[var(--color-text-tertiary)]">
            {$t('projectHub.category')}
          </span>
          <div class="mt-1">
            <span
              class="rounded-[var(--radius-full)] bg-[var(--color-bg-tertiary)] px-3 py-1 text-sm font-medium text-[var(--color-text-secondary)]"
            >
              {selectedProject.category === 'flagship'
                ? $t('projectHub.categoryFlagship')
                : $t('projectHub.categoryLab')}
            </span>
          </div>
        </div>
      </div>

      <!-- Tags -->
      {#if (selectedProject.tags ?? []).length > 0}
        <div class="mt-6">
          <span class="text-xs font-medium text-[var(--color-text-tertiary)]">
            {$t('projectHub.tags')}
          </span>
          <div class="mt-1 flex flex-wrap gap-1.5">
            {#each selectedProject.tags as tag}
              <span
                class="rounded-[var(--radius-full)] px-2.5 py-1 text-xs font-medium"
                style="background-color: {tag.color}20; color: {tag.color}; border: 1px solid {tag.color}40"
              >
                {tag.name}
              </span>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Hosting -->
      {#if selectedProject.hosting}
        <div class="mt-4">
          <span class="text-xs font-medium text-[var(--color-text-tertiary)]">
            {$t('projectHub.hosting')}
          </span>
          <p class="mt-1 text-sm text-[var(--color-text-primary)]">{selectedProject.hosting}</p>
        </div>
      {/if}

      <!-- Links -->
      {#if selectedProject.repo_url || selectedProject.web_url || selectedProject.docs_url || selectedProject.links.length > 0}
        <div class="mt-6">
          <span class="text-xs font-medium text-[var(--color-text-tertiary)]">
            {$t('projectHub.links')}
          </span>
          <div class="mt-2 space-y-2">
            {#if selectedProject.repo_url}
              <a
                href={selectedProject.repo_url}
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-2 text-sm text-[var(--color-brand-blue)] hover:underline"
              >
                <Github size={14} />
                {selectedProject.repo_url}
                <ExternalLink size={12} />
              </a>
            {/if}
            {#if selectedProject.web_url}
              <a
                href={selectedProject.web_url}
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-2 text-sm text-[var(--color-brand-blue)] hover:underline"
              >
                <Globe size={14} />
                {selectedProject.web_url}
                <ExternalLink size={12} />
              </a>
            {/if}
            {#if selectedProject.docs_url}
              <a
                href={selectedProject.docs_url}
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-2 text-sm text-[var(--color-brand-blue)] hover:underline"
              >
                <BookOpen size={14} />
                {selectedProject.docs_url}
                <ExternalLink size={12} />
              </a>
            {/if}
            {#each selectedProject.links as link}
              <a
                href={link.url}
                target="_blank"
                rel="noopener noreferrer"
                class="flex items-center gap-2 text-sm text-[var(--color-brand-blue)] hover:underline"
              >
                <ExternalLink size={14} />
                {link.label}: {link.url}
              </a>
            {/each}
          </div>
        </div>
      {/if}

      <!-- Notes -->
      {#if selectedProject.notes}
        <div class="mt-6">
          <span class="text-xs font-medium text-[var(--color-text-tertiary)]">
            {$t('projectHub.notes')}
          </span>
          <p class="mt-1 whitespace-pre-wrap text-sm text-[var(--color-text-primary)]">
            {selectedProject.notes}
          </p>
        </div>
      {/if}

      <!-- Timestamps -->
      <div class="mt-8 flex gap-6 border-t border-[var(--color-border)] pt-4">
        <span class="text-xs text-[var(--color-text-tertiary)]">
          {$t('projectHub.created')}: {formatDate(selectedProject.created_at)}
        </span>
        <span class="text-xs text-[var(--color-text-tertiary)]">
          {$t('projectHub.updated')}: {formatDate(selectedProject.updated_at)}
        </span>
      </div>
    </div>

    <!-- Delete confirmation -->
    {#if showDeleteConfirm}
      <div
        class="fixed inset-0 z-50 flex items-center justify-center bg-black/50"
        role="dialog"
        aria-modal="true"
      >
        <div
          class="w-full max-w-md rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-primary)] p-6 shadow-lg"
        >
          <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">
            {$t('projectHub.deleteProject')}
          </h3>
          <p class="mt-2 text-sm text-[var(--color-text-secondary)]">
            {$t('projectHub.confirmDelete', { values: { name: selectedProject.name } })}
          </p>
          <div class="mt-6 flex justify-end gap-2">
            <button
              onclick={() => (showDeleteConfirm = false)}
              class="rounded-[var(--radius-md)] border border-[var(--color-border)] px-4 py-2 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
            >
              {$t('projectHub.cancel')}
            </button>
            <button
              onclick={deleteProject}
              disabled={submitting}
              class="rounded-[var(--radius-md)] bg-[var(--color-error)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90 disabled:opacity-50"
            >
              {$t('projectHub.confirm')}
            </button>
          </div>
        </div>
      </div>
    {/if}
  {/if}

  <!-- ============ CREATE/EDIT MODAL ============ -->
  {#if showModal}
    <div
      class="fixed inset-0 z-50 flex items-start justify-center overflow-y-auto bg-black/50 p-4 pt-12"
      role="dialog"
      aria-modal="true"
    >
      <div
        class="w-full max-w-2xl rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-primary)] shadow-lg"
      >
        <!-- Modal header -->
        <div class="flex items-center justify-between border-b border-[var(--color-border)] px-6 py-4">
          <h3 class="text-lg font-semibold text-[var(--color-text-primary)]">
            {editingSlug ? $t('projectHub.editProject') : $t('projectHub.addProject')}
          </h3>
          <button
            onclick={() => (showModal = false)}
            class="rounded-[var(--radius-sm)] p-1.5 text-[var(--color-text-tertiary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
          >
            <X size={18} />
          </button>
        </div>

        <!-- Modal body -->
        <form
          onsubmit={(e) => {
            e.preventDefault();
            saveProject();
          }}
          class="space-y-4 p-6"
        >
          <div class="grid grid-cols-1 gap-4 sm:grid-cols-2">
            <!-- Name -->
            <div>
              <label for="ph-name" class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]">
                {$t('projectHub.name')} *
              </label>
              <input
                id="ph-name"
                type="text"
                bind:value={formName}
                required
                maxlength="100"
                class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
              />
            </div>

            <!-- Tagline -->
            <div>
              <label for="ph-tagline" class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]">
                {$t('projectHub.tagline')} *
              </label>
              <input
                id="ph-tagline"
                type="text"
                bind:value={formTagline}
                required
                maxlength="200"
                class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
              />
            </div>

            <!-- Status -->
            <div>
              <label for="ph-status" class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]">
                {$t('projectHub.status')} *
              </label>
              <select
                id="ph-status"
                bind:value={formStatus}
                required
                class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
              >
                {#each Object.entries(statusConfig) as [value, config]}
                  <option {value}>{$t(config.label)}</option>
                {/each}
              </select>
            </div>

            <!-- Category -->
            <div>
              <label for="ph-category" class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]">
                {$t('projectHub.category')} *
              </label>
              <select
                id="ph-category"
                bind:value={formCategory}
                required
                class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
              >
                <option value="flagship">{$t('projectHub.categoryFlagship')}</option>
                <option value="lab">{$t('projectHub.categoryLab')}</option>
              </select>
            </div>

            <!-- Version -->
            <div>
              <label for="ph-version" class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]">
                {$t('projectHub.version')}
              </label>
              <input
                id="ph-version"
                type="text"
                bind:value={formVersion}
                class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
                placeholder="v0.1.0"
              />
            </div>

            <!-- Tags -->
            <div class="sm:col-span-2">
              <div class="mb-1 text-sm font-medium text-[var(--color-text-secondary)]">
                {$t('projectHub.tags')} *
              </div>
              <div class="flex flex-wrap gap-2 rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-primary)] p-3 min-h-[42px]">
                {#each allTags as tag}
                  <button
                    type="button"
                    onclick={() => toggleFormTag(tag.id)}
                    class="rounded-[var(--radius-full)] px-2.5 py-1 text-xs font-medium transition-all"
                    style={formTagIds.includes(tag.id)
                      ? `background-color: ${tag.color}; color: white; border: 1px solid ${tag.color}`
                      : `background-color: transparent; color: ${tag.color}; border: 1px solid ${tag.color}40`}
                  >
                    {tag.name}
                  </button>
                {/each}
                {#if allTags.length === 0}
                  <span class="text-xs text-[var(--color-text-tertiary)]">
                    {$t('projectHub.noTags')}
                  </span>
                {/if}
              </div>
            </div>

            <!-- Icon -->
            <div>
              <label for="ph-icon" class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]">
                {$t('projectHub.icon')}
              </label>
              <select
                id="ph-icon"
                bind:value={formIcon}
                class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
              >
                {#each iconOptions as icon}
                  <option value={icon}>{icon}</option>
                {/each}
              </select>
            </div>

            <!-- Color -->
            <div>
              <label for="ph-color" class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]">
                {$t('projectHub.color')}
              </label>
              <div class="flex items-center gap-2">
                <input
                  id="ph-color"
                  type="color"
                  bind:value={formColor}
                  class="h-9 w-12 cursor-pointer rounded-[var(--radius-sm)] border border-[var(--color-border)]"
                />
                <input
                  type="text"
                  bind:value={formColor}
                  pattern="#[0-9A-Fa-f]{'{'}6{'}'}"
                  class="flex-1 rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
                />
              </div>
            </div>

            <!-- Hosting -->
            <div>
              <label for="ph-hosting" class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]">
                {$t('projectHub.hosting')}
              </label>
              <input
                id="ph-hosting"
                type="text"
                bind:value={formHosting}
                class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
                placeholder="Vercel, Railway, Self-hosted..."
              />
            </div>

            <!-- Sort Order -->
            <div>
              <label for="ph-sort" class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]">
                {$t('projectHub.sortOrder')}
              </label>
              <input
                id="ph-sort"
                type="number"
                min="0"
                bind:value={formSortOrder}
                class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
              />
            </div>
          </div>

          <!-- Links section -->
          <div class="space-y-3 border-t border-[var(--color-border)] pt-4">
            <span class="text-sm font-medium text-[var(--color-text-secondary)]">
              {$t('projectHub.links')}
            </span>

            <div class="grid grid-cols-1 gap-4 sm:grid-cols-3">
              <div>
                <label for="ph-repo" class="mb-1 block text-xs text-[var(--color-text-tertiary)]">
                  {$t('projectHub.repoUrl')}
                </label>
                <input
                  id="ph-repo"
                  type="url"
                  bind:value={formRepoUrl}
                  class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
                  placeholder="https://github.com/..."
                />
              </div>
              <div>
                <label for="ph-web" class="mb-1 block text-xs text-[var(--color-text-tertiary)]">
                  {$t('projectHub.webUrl')}
                </label>
                <input
                  id="ph-web"
                  type="url"
                  bind:value={formWebUrl}
                  class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
                  placeholder="https://..."
                />
              </div>
              <div>
                <label for="ph-docs" class="mb-1 block text-xs text-[var(--color-text-tertiary)]">
                  {$t('projectHub.docsUrl')}
                </label>
                <input
                  id="ph-docs"
                  type="url"
                  bind:value={formDocsUrl}
                  class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
                  placeholder="https://docs...."
                />
              </div>
            </div>

            <!-- Custom links -->
            {#each formCustomLinks as link, i}
              <div class="flex items-end gap-2">
                <div class="flex-1">
                  <label for="ph-link-label-{i}" class="mb-1 block text-xs text-[var(--color-text-tertiary)]">
                    {$t('projectHub.linkLabel')}
                  </label>
                  <input
                    id="ph-link-label-{i}"
                    type="text"
                    bind:value={link.label}
                    class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
                    placeholder="Play Store, npm..."
                  />
                </div>
                <div class="flex-1">
                  <label for="ph-link-url-{i}" class="mb-1 block text-xs text-[var(--color-text-tertiary)]">
                    {$t('projectHub.linkUrl')}
                  </label>
                  <input
                    id="ph-link-url-{i}"
                    type="url"
                    bind:value={link.url}
                    class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
                    placeholder="https://..."
                  />
                </div>
                <button
                  type="button"
                  onclick={() => removeCustomLink(i)}
                  class="mb-0.5 rounded-[var(--radius-sm)] p-2 text-[var(--color-text-tertiary)] transition-colors hover:bg-[var(--color-error)]/10 hover:text-[var(--color-error)]"
                >
                  <Trash2 size={14} />
                </button>
              </div>
            {/each}

            <button
              type="button"
              onclick={addCustomLink}
              class="flex items-center gap-1 text-sm font-medium text-[var(--color-brand-blue)] hover:underline"
            >
              <Plus size={14} />
              {$t('projectHub.addLink')}
            </button>
          </div>

          <!-- Notes -->
          <div>
            <label for="ph-notes" class="mb-1 block text-sm font-medium text-[var(--color-text-secondary)]">
              {$t('projectHub.notes')}
            </label>
            <textarea
              id="ph-notes"
              bind:value={formNotes}
              rows="3"
              class="w-full rounded-[var(--radius-sm)] border border-[var(--color-border)] bg-[var(--color-bg-tertiary)] px-3 py-2 text-sm text-[var(--color-text-primary)] placeholder:text-[var(--color-text-tertiary)] focus:border-[var(--color-brand-blue)] focus:outline-none focus:ring-1 focus:ring-[var(--color-brand-blue)]"
            ></textarea>
          </div>

          <!-- Error message -->
          {#if error}
            <div class="rounded-[var(--radius-sm)] border border-[var(--color-error)]/20 bg-[var(--color-error)]/5 p-3">
              <p class="text-sm text-[var(--color-error)]">{error}</p>
            </div>
          {/if}

          <!-- Actions -->
          <div class="flex justify-end gap-2 border-t border-[var(--color-border)] pt-4">
            <button
              type="button"
              onclick={() => (showModal = false)}
              class="rounded-[var(--radius-md)] border border-[var(--color-border)] px-4 py-2 text-sm font-medium text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)]"
            >
              {$t('projectHub.cancel')}
            </button>
            <button
              type="submit"
              disabled={submitting}
              class="rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90 disabled:opacity-50"
            >
              {$t('projectHub.save')}
            </button>
          </div>
        </form>
      </div>
    </div>
  {/if}
</div>
