<script lang="ts">
  import { t } from 'svelte-i18n';

  interface ProjectHubWidgetData {
    total: number;
    by_status: Record<string, number>;
  }

  interface Props {
    data: ProjectHubWidgetData | null;
  }

  let { data }: Props = $props();

  const statusConfig: Record<string, { label: string; color: string }> = {
    active: { label: 'projectHub.statusActive', color: '#16A34A' },
    development: { label: 'projectHub.statusDevelopment', color: '#0070F3' },
    design: { label: 'projectHub.statusDesign', color: '#6366F1' },
    concept: { label: 'projectHub.statusConcept', color: 'var(--color-text-tertiary)' },
    maintenance: { label: 'projectHub.statusMaintenance', color: '#D97706' },
    archived: { label: 'projectHub.statusArchived', color: 'var(--color-text-tertiary)' },
    absorbed: { label: 'projectHub.statusAbsorbed', color: 'var(--color-text-tertiary)' },
  };

  const statusOrder = ['active', 'development', 'design', 'concept', 'maintenance', 'archived', 'absorbed'];

  const visibleStatuses = $derived(
    data ? statusOrder.filter((s) => (data.by_status[s] ?? 0) > 0) : [],
  );
</script>

{#if data}
  <div class="space-y-3">
    <p class="text-lg font-semibold text-[var(--color-text-primary)]">
      {data.total}
      <span class="text-sm font-normal text-[var(--color-text-secondary)]">
        {$t('projectHub.projects')}
      </span>
    </p>

    <div class="space-y-1.5">
      {#each visibleStatuses as status}
        {@const config = statusConfig[status]}
        {@const count = data.by_status[status] ?? 0}
        <div class="flex items-center gap-2">
          <span
            class="inline-block h-2 w-2 shrink-0 rounded-[var(--radius-full)]"
            style="background-color: {config.color}"
          ></span>
          <span class="text-xs text-[var(--color-text-secondary)]">
            {count} {$t(config.label)}
          </span>
        </div>
      {/each}
    </div>
  </div>
{:else}
  <p class="text-sm text-[var(--color-text-tertiary)]">{$t('projectHub.noProjects')}</p>
{/if}
