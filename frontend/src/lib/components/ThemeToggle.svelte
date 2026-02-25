<script lang="ts">
  import Sun from 'lucide-svelte/icons/sun';
  import Moon from 'lucide-svelte/icons/moon';
  import { theme, toggleTheme } from '$lib/stores/theme';

  let currentTheme = $state<'light' | 'dark'>('light');

  theme.subscribe((v) => {
    currentTheme = v;
  });

  const label = $derived(currentTheme === 'light' ? 'Light' : 'Dark');
</script>

<button
  onclick={toggleTheme}
  class="flex items-center justify-center rounded-[var(--radius-md)] p-2 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)] hover:text-[var(--color-text-primary)]"
  aria-label="Toggle theme: {label}"
  title="Theme: {label}"
>
  {#if currentTheme === 'light'}
    <Sun size={20} />
  {:else}
    <Moon size={20} />
  {/if}
</button>
