<script lang="ts">
  import Globe from 'lucide-svelte/icons/globe';
  import { locale } from 'svelte-i18n';

  let open = $state(false);

  let currentLocale = $state('en');
  locale.subscribe((v) => {
    currentLocale = v ?? 'en';
  });

  const languages = [
    { code: 'en', label: 'EN' },
    { code: 'es', label: 'ES' },
  ] as const;

  function selectLocale(code: string) {
    locale.set(code);
    localStorage.setItem('cortex-locale', code);
    open = false;
  }

  function handleClickOutside(event: MouseEvent) {
    const target = event.target as HTMLElement;
    if (!target.closest('.language-selector')) {
      open = false;
    }
  }
</script>

<svelte:window onclick={handleClickOutside} />

<div class="language-selector relative">
  <button
    onclick={() => (open = !open)}
    class="flex items-center gap-1.5 rounded-[var(--radius-md)] p-2 text-[var(--color-text-secondary)] transition-colors hover:bg-[var(--color-bg-tertiary)] hover:text-[var(--color-text-primary)]"
    aria-label="Select language"
    aria-expanded={open}
  >
    <Globe size={20} />
    <span class="text-sm font-medium">{currentLocale.toUpperCase()}</span>
  </button>

  {#if open}
    <div
      class="absolute right-0 top-full z-50 mt-1 min-w-[80px] overflow-hidden rounded-[var(--radius-md)] border border-[var(--color-border)] bg-[var(--color-bg-primary)] shadow-[var(--shadow-md)]"
    >
      {#each languages as lang}
        <button
          onclick={() => selectLocale(lang.code)}
          class="flex w-full items-center px-3 py-2 text-sm transition-colors {currentLocale ===
          lang.code
            ? 'bg-[var(--color-bg-tertiary)] font-medium text-[var(--color-text-primary)]'
            : 'text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)]'}"
        >
          {lang.label}
        </button>
      {/each}
    </div>
  {/if}
</div>
