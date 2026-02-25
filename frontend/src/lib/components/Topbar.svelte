<script lang="ts">
  import Menu from 'lucide-svelte/icons/menu';
  import ThemeToggle from './ThemeToggle.svelte';
  import LanguageSelector from './LanguageSelector.svelte';

  interface Props {
    onMenuClick?: () => void;
  }

  let { onMenuClick }: Props = $props();
</script>

<header
  class="flex h-14 shrink-0 items-center justify-between border-b border-[var(--color-border)] bg-[var(--color-bg-primary)] px-4 md:px-8"
>
  <div class="flex items-center gap-3">
    <!-- Hamburger (mobile only) -->
    {#if onMenuClick}
      <button
        onclick={onMenuClick}
        class="flex items-center justify-center rounded-[var(--radius-md)] p-2 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)] hover:text-[var(--color-text-primary)] md:hidden"
        aria-label="Toggle menu"
      >
        <Menu size={20} />
      </button>
    {/if}
  </div>

  <div class="flex items-center gap-1">
    <LanguageSelector />
    <ThemeToggle />
  </div>
</header>
