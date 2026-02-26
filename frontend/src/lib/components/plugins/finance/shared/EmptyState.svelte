<script lang="ts">
  import type { Component } from 'svelte';

  interface Props {
    icon: Component<{ size?: number }>;
    message: string;
    actionLabel?: string;
    onaction?: () => void;
  }

  const { icon: IconComponent, message, actionLabel, onaction }: Props = $props();
</script>

<div
  class="flex flex-col items-center justify-center rounded-[var(--radius-lg)] border border-[var(--color-border)] bg-[var(--color-bg-secondary)] px-6 py-12 text-center"
>
  <div class="mb-3 text-[var(--color-text-tertiary)]">
    <IconComponent size={32} />
  </div>

  <p class="text-sm text-[var(--color-text-tertiary)]">
    {message}
  </p>

  {#if actionLabel && onaction}
    <button
      onclick={onaction}
      class="mt-4 rounded-[var(--radius-md)] bg-[var(--color-brand-blue)] px-4 py-2 text-sm font-medium text-white transition-colors hover:opacity-90"
    >
      {actionLabel}
    </button>
  {/if}
</div>
