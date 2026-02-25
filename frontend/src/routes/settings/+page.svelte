<script lang="ts">
  import { t } from 'svelte-i18n';
  import Settings from 'lucide-svelte/icons/settings';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import Circle from 'lucide-svelte/icons/circle';
  import Puzzle from 'lucide-svelte/icons/puzzle';
  import LoaderCircle from 'lucide-svelte/icons/loader-circle';
  import { plugins, uninstallPlugin } from '$lib/stores/plugins';
  import { getPluginIcon } from '$lib/utils/plugin-icons';
  import type { PluginManifest } from '$lib/types';

  let pluginList = $state<PluginManifest[]>([]);
  plugins.subscribe((v) => {
    pluginList = v;
  });

  let uninstallConfirmId = $state<string | null>(null);
  let uninstalling = $state(false);

  async function handleUninstall(pluginId: string): Promise<void> {
    uninstalling = true;
    try {
      await uninstallPlugin(pluginId);
    } catch {
      // fetchPlugins inside uninstallPlugin already handles refresh
    } finally {
      uninstalling = false;
      uninstallConfirmId = null;
    }
  }
</script>

<div class="space-y-6">
  <!-- Header -->
  <div class="flex items-center gap-3">
    <Settings size={20} class="text-[var(--color-text-secondary)]" />
    <h2 class="text-2xl font-semibold text-[var(--color-text-primary)]">
      {$t('settings.title')}
    </h2>
  </div>

  <!-- Installed Plugins section -->
  <div>
    <h3 class="mb-4 text-lg font-semibold text-[var(--color-text-primary)]">
      {$t('settings.plugins')}
    </h3>

    {#if pluginList.length === 0}
      <!-- Empty state -->
      <div
        class="flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-8 py-16"
      >
        <Puzzle size={48} class="mb-4 text-[var(--color-text-tertiary)]" />
        <h4 class="mb-2 text-xl font-semibold text-[var(--color-text-primary)]">
          {$t('settings.noPlugins.title')}
        </h4>
        <p class="text-center text-sm text-[var(--color-text-secondary)]">
          {$t('settings.noPlugins.description')}
        </p>
      </div>
    {:else}
      <div class="space-y-3">
        {#each pluginList as plugin}
          {@const IconComponent = getPluginIcon(plugin.icon)}
          <div
            class="flex items-center justify-between rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-6 py-4 transition-colors hover:bg-[var(--color-bg-tertiary)]"
          >
            <div class="flex items-center gap-4">
              <div
                class="flex h-10 w-10 items-center justify-center rounded-[var(--radius-md)]"
                style="background-color: {plugin.color}15; color: {plugin.color}"
              >
                <IconComponent size={20} />
              </div>
              <div>
                <div class="flex items-center gap-2">
                  <span class="text-sm font-semibold text-[var(--color-text-primary)]">
                    {plugin.name}
                  </span>
                  <span
                    class="rounded-[var(--radius-sm)] bg-[var(--color-bg-tertiary)] px-2 py-0.5 font-mono text-xs text-[var(--color-text-secondary)]"
                  >
                    v{plugin.version}
                  </span>
                  <span
                    class="flex items-center gap-1 rounded-[var(--radius-full)] bg-[var(--color-success)]/10 px-2 py-0.5 text-xs font-medium text-[var(--color-success)]"
                  >
                    <Circle size={8} fill="currentColor" />
                    {$t('settings.status.running')}
                  </span>
                </div>
                <p class="mt-0.5 text-sm text-[var(--color-text-secondary)]">
                  {plugin.description}
                </p>
              </div>
            </div>

            {#if uninstallConfirmId === plugin.id}
              <div class="flex items-center gap-2">
                <span class="text-sm text-[var(--color-text-secondary)]">
                  {$t('settings.confirmUninstall')}
                </span>
                <button
                  class="flex items-center gap-1.5 rounded-[var(--radius-md)] bg-[var(--color-error)] px-3 py-2 text-sm font-medium text-white transition-colors hover:bg-[var(--color-error)]/90 disabled:opacity-50"
                  aria-label="{$t('settings.confirm')} {$t('settings.uninstall')} {plugin.name}"
                  disabled={uninstalling}
                  onclick={() => handleUninstall(plugin.id)}
                >
                  {#if uninstalling}
                    <LoaderCircle size={16} class="animate-spin" />
                  {/if}
                  {$t('settings.confirm')}
                </button>
                <button
                  class="rounded-[var(--radius-md)] px-3 py-2 text-sm text-[var(--color-text-secondary)] transition-colors hover:text-[var(--color-text-primary)]"
                  aria-label="{$t('dashboard.cancel')}"
                  disabled={uninstalling}
                  onclick={() => (uninstallConfirmId = null)}
                >
                  {$t('dashboard.cancel')}
                </button>
              </div>
            {:else}
              <button
                class="flex items-center gap-2 rounded-[var(--radius-md)] border border-[var(--color-error)]/20 px-3 py-2 text-sm font-medium text-[var(--color-error)] transition-colors hover:bg-[var(--color-error)]/10"
                aria-label="{$t('settings.uninstall')} {plugin.name}"
                onclick={() => (uninstallConfirmId = plugin.id)}
              >
                <Trash2 size={16} />
                <span class="hidden sm:inline">{$t('settings.uninstall')}</span>
              </button>
            {/if}
          </div>
        {/each}
      </div>
    {/if}
  </div>
</div>
