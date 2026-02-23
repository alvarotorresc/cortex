<script lang="ts">
  import GripVertical from 'lucide-svelte/icons/grip-vertical';
  import { getPluginIcon } from '$lib/utils/plugin-icons';
  import type { Snippet } from 'svelte';

  interface Props {
    pluginName: string;
    pluginIcon: string;
    pluginColor: string;
    editMode?: boolean;
    children: Snippet;
  }

  let { pluginName, pluginIcon, pluginColor, editMode = false, children }: Props = $props();

  const IconComponent = $derived(getPluginIcon(pluginIcon));
</script>

<div
  class="flex flex-col rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] transition-shadow hover:shadow-[var(--shadow-sm)]"
>
  <!-- Header -->
  <div
    class="flex items-center gap-3 border-b border-[var(--color-border)] px-6 py-4"
  >
    {#if editMode}
      <span class="cursor-grab text-[var(--color-text-tertiary)]">
        <GripVertical size={16} />
      </span>
    {/if}
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
