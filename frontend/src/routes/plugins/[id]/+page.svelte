<script lang="ts">
  import { t } from 'svelte-i18n';
  import Puzzle from 'lucide-svelte/icons/puzzle';
  import FinancePage from '$lib/components/plugins/FinancePage.svelte';
  import NotesPage from '$lib/components/plugins/NotesPage.svelte';
  import ProjectHubPage from '$lib/components/plugins/ProjectHubPage.svelte';

  let { data } = $props();

  const pluginId = $derived(data.pluginId);
</script>

{#if pluginId === 'finance-tracker'}
  <FinancePage />
{:else if pluginId === 'quick-notes'}
  <NotesPage />
{:else if pluginId === 'project-hub'}
  <ProjectHubPage />
{:else}
  <div
    class="flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-8 py-16"
  >
    <Puzzle size={48} class="mb-4 text-[var(--color-text-tertiary)]" />
    <h3 class="mb-2 text-xl font-semibold text-[var(--color-text-primary)]">
      Plugin page not available
    </h3>
    <p class="text-center text-sm text-[var(--color-text-secondary)]">
      No UI is available for plugin "{pluginId}".
    </p>
  </div>
{/if}
