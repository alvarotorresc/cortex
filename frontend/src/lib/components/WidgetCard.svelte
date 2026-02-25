<script lang="ts">
  import { getPluginIcon } from '$lib/utils/plugin-icons';
  import type { Snippet } from 'svelte';

  interface Props {
    pluginName: string;
    pluginIcon: string;
    pluginColor: string;
    children: Snippet;
  }

  let { pluginName, pluginIcon, pluginColor, children }: Props = $props();

  const IconComponent = $derived(getPluginIcon(pluginIcon));
</script>

<div
  class="flex flex-col rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] transition-shadow hover:shadow-[var(--shadow-sm)]"
>
  <!-- Header -->
  <div
    class="flex items-center gap-3 border-b border-[var(--color-border)] px-6 py-4"
  >
    <span style="color: {pluginColor}">
      <IconComponent size={20} />
    </span>
    <span class="text-sm font-semibold text-[var(--color-text-primary)]">{pluginName}</span>
  </div>

  <!-- Body -->
  <div class="flex-1 p-6">
    {@render children()}
  </div>
</div>
