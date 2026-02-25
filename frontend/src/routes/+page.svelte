<script lang="ts">
  import { t } from 'svelte-i18n';
  import { onMount } from 'svelte';
  import Home from 'lucide-svelte/icons/home';
  import Puzzle from 'lucide-svelte/icons/puzzle';
  import TrendingUp from 'lucide-svelte/icons/trending-up';
  import TrendingDown from 'lucide-svelte/icons/trending-down';
  import WidgetCard from '$lib/components/WidgetCard.svelte';
  import ProjectHubWidget from '$lib/components/plugins/ProjectHubWidget.svelte';
  import { plugins } from '$lib/stores/plugins';
  import { pluginApi } from '$lib/api';
  import type { PluginManifest } from '$lib/types';

  interface FinanceWidgetData {
    income: number;
    expense: number;
    balance: number;
    currency: string;
  }

  interface NotePreview {
    id: string;
    title: string;
    content: string;
    pinned: boolean;
  }

  interface NotesWidgetData {
    latest: NotePreview[];
    pinned_count: number;
  }

  interface ProjectHubWidgetData {
    total: number;
    by_status: Record<string, number>;
  }

  let pluginList = $state<PluginManifest[]>([]);
  plugins.subscribe((v) => {
    pluginList = v;
  });

  let financeData = $state<FinanceWidgetData | null>(null);
  let notesData = $state<NotesWidgetData | null>(null);
  let projectHubData = $state<ProjectHubWidgetData | null>(null);
  let widgetLoading = $state(true);

  const hasFinance = $derived(pluginList.some((p) => p.id === 'finance-tracker'));
  const hasNotes = $derived(pluginList.some((p) => p.id === 'quick-notes'));
  const hasProjectHub = $derived(pluginList.some((p) => p.id === 'project-hub'));
  const hasPlugins = $derived(pluginList.length > 0);

  const financePlugin = $derived(pluginList.find((p) => p.id === 'finance-tracker'));
  const notesPlugin = $derived(pluginList.find((p) => p.id === 'quick-notes'));
  const projectHubPlugin = $derived(pluginList.find((p) => p.id === 'project-hub'));

  async function loadWidgets() {
    widgetLoading = true;

    try {
      const promises: Promise<void>[] = [];

      if (hasFinance) {
        promises.push(
          pluginApi('finance-tracker')
            .widget<{ data: FinanceWidgetData }>('dashboard-widget')
            .then((res) => {
              financeData = res.data;
            })
            .catch(() => {
              financeData = null;
            }),
        );
      }

      if (hasNotes) {
        promises.push(
          pluginApi('quick-notes')
            .widget<{ data: NotesWidgetData }>('dashboard-widget')
            .then((res) => {
              notesData = res.data;
            })
            .catch(() => {
              notesData = null;
            }),
        );
      }

      if (hasProjectHub) {
        promises.push(
          pluginApi('project-hub')
            .widget<{ data: ProjectHubWidgetData }>('dashboard-widget')
            .then((res) => {
              projectHubData = res.data;
            })
            .catch(() => {
              projectHubData = null;
            }),
        );
      }

      await Promise.allSettled(promises);
    } finally {
      widgetLoading = false;
    }
  }

  function formatCurrency(amount: number, currency: string = 'EUR'): string {
    return new Intl.NumberFormat(undefined, { style: 'currency', currency }).format(amount);
  }

  onMount(() => {
    // Wait a tick for plugins store to be populated
    const unsub = plugins.subscribe((list) => {
      if (list.length > 0) {
        loadWidgets();
      } else {
        widgetLoading = false;
      }
    });
    return unsub;
  });
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div class="flex items-center gap-3">
      <Home size={20} class="text-[var(--color-text-secondary)]" />
      <h2 class="text-2xl font-semibold text-[var(--color-text-primary)]">
        {$t('dashboard.title')}
      </h2>
    </div>
  </div>

  <!-- Widget Grid -->
  {#if widgetLoading}
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
      {#each [1, 2] as _}
        <div
          class="h-48 animate-pulse rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)]"
        ></div>
      {/each}
    </div>
  {:else if !hasPlugins}
    <!-- Empty state -->
    <div
      class="flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-8 py-16"
    >
      <Puzzle size={48} class="mb-4 text-[var(--color-text-tertiary)]" />
      <h3 class="mb-2 text-xl font-semibold text-[var(--color-text-primary)]">
        {$t('dashboard.empty.title')}
      </h3>
      <p class="text-center text-sm text-[var(--color-text-secondary)]">
        {$t('dashboard.empty.description')}
      </p>
    </div>
  {:else}
    <div class="grid grid-cols-1 gap-4 lg:grid-cols-2">
      <!-- Finance widget -->
      {#if hasFinance && financePlugin}
        <WidgetCard
          pluginName={financePlugin.name}
          pluginIcon={financePlugin.icon}
          pluginColor={financePlugin.color}
        >
          {#if financeData}
            <div class="space-y-4">
              <div class="grid grid-cols-3 gap-4">
                <div>
                  <p class="text-xs font-medium text-[var(--color-text-secondary)]">
                    {$t('finance.income')}
                  </p>
                  <p class="flex items-center gap-1 text-lg font-semibold text-[var(--color-success)]">
                    <TrendingUp size={16} />
                    {formatCurrency(financeData.income, financeData.currency)}
                  </p>
                </div>
                <div>
                  <p class="text-xs font-medium text-[var(--color-text-secondary)]">
                    {$t('finance.expense')}
                  </p>
                  <p class="flex items-center gap-1 text-lg font-semibold text-[var(--color-error)]">
                    <TrendingDown size={16} />
                    {formatCurrency(financeData.expense, financeData.currency)}
                  </p>
                </div>
                <div>
                  <p class="text-xs font-medium text-[var(--color-text-secondary)]">
                    {$t('finance.balance')}
                  </p>
                  <p
                    class="text-lg font-semibold {financeData.balance >= 0
                      ? 'text-[var(--color-success)]'
                      : 'text-[var(--color-error)]'}"
                  >
                    {formatCurrency(financeData.balance, financeData.currency)}
                  </p>
                </div>
              </div>
            </div>
          {:else}
            <p class="text-sm text-[var(--color-text-tertiary)]">
              {$t('finance.noTransactions')}
            </p>
          {/if}
        </WidgetCard>
      {/if}

      <!-- Notes widget -->
      {#if hasNotes && notesPlugin}
        <WidgetCard
          pluginName={notesPlugin.name}
          pluginIcon={notesPlugin.icon}
          pluginColor={notesPlugin.color}
        >
          {#if notesData && notesData.latest.length > 0}
            <div class="space-y-3">
              {#each notesData.latest.slice(0, 3) as note}
                <a
                  href="/plugins/quick-notes?note={note.id}"
                  class="flex items-start gap-3 rounded-[var(--radius-sm)] p-2 transition-colors hover:bg-[var(--color-bg-tertiary)]"
                >
                  <div class="flex-1 overflow-hidden">
                    <div class="flex items-center gap-2">
                      <p class="truncate text-sm font-medium text-[var(--color-text-primary)]">
                        {note.title}
                      </p>
                      {#if note.pinned}
                        <span
                          class="shrink-0 rounded-[var(--radius-full)] bg-[var(--color-plugin-notes)]/10 px-2 py-0.5 text-xs font-medium text-[var(--color-plugin-notes)]"
                        >
                          {$t('notes.pin')}
                        </span>
                      {/if}
                    </div>
                    <p class="mt-0.5 truncate text-xs text-[var(--color-text-tertiary)]">
                      {note.content}
                    </p>
                  </div>
                </a>
              {/each}
            </div>
          {:else}
            <p class="text-sm text-[var(--color-text-tertiary)]">{$t('notes.noNotes')}</p>
          {/if}
        </WidgetCard>
      {/if}

      <!-- Project Hub widget -->
      {#if hasProjectHub && projectHubPlugin}
        <a href="/plugins/project-hub" class="block">
          <WidgetCard
            pluginName={projectHubPlugin.name}
            pluginIcon={projectHubPlugin.icon}
            pluginColor={projectHubPlugin.color}
          >
            <ProjectHubWidget data={projectHubData} />
          </WidgetCard>
        </a>
      {/if}
    </div>
  {/if}
</div>
