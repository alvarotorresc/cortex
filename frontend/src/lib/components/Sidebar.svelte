<script lang="ts">
  import { page } from '$app/stores';
  import { t } from 'svelte-i18n';
  import Home from 'lucide-svelte/icons/home';
  import Settings from 'lucide-svelte/icons/settings';
  import CortexLogo from './CortexLogo.svelte';
  import { sidebarCollapsed } from '$lib/stores/sidebar';
  import { getPluginIcon } from '$lib/utils/plugin-icons';
  import type { PluginManifest } from '$lib/types';

  interface Props {
    plugins: PluginManifest[];
  }

  let { plugins }: Props = $props();

  let collapsed = $state(false);
  sidebarCollapsed.subscribe((v) => {
    collapsed = v;
  });

  let currentPath = $state('/');
  page.subscribe((p) => {
    currentPath = p.url.pathname;
  });

  interface NavItem {
    href: string;
    label: string;
    // eslint-disable-next-line @typescript-eslint/no-explicit-any
    icon: any;
    color?: string;
  }

  const staticTop: NavItem[] = [{ href: '/', label: 'nav.home', icon: Home }];

  const pluginItems = $derived<NavItem[]>(
    plugins.map((p) => ({
      href: `/plugins/${p.id}`,
      label: p.name,
      icon: getPluginIcon(p.icon),
      color: p.color,
    })),
  );

  const staticBottom: NavItem[] = [{ href: '/settings', label: 'nav.settings', icon: Settings }];

  function isActive(href: string): boolean {
    if (href === '/') return currentPath === '/';
    return currentPath.startsWith(href);
  }
</script>

<aside
  class="flex h-full flex-col border-r border-[var(--color-border)] bg-[var(--color-bg-secondary)] transition-[width] duration-200"
  style="width: {collapsed ? '64px' : '240px'}"
>
  <!-- Logo -->
  <div class="flex h-14 items-center gap-3 border-b border-[var(--color-border)] px-4">
    <CortexLogo size={24} color="var(--color-cortex-emerald)" />
    {#if !collapsed}
      <span class="text-lg font-semibold text-[var(--color-text-primary)]">Cortex</span>
    {/if}
  </div>

  <!-- Navigation -->
  <nav class="flex flex-1 flex-col gap-1 overflow-y-auto px-3 py-3">
    <!-- Top items -->
    {#each staticTop as item}
      {@const active = isActive(item.href)}
      <a
        href={item.href}
        class="flex items-center gap-3 rounded-[var(--radius-md)] px-3 py-2 text-sm font-medium transition-colors
          {active
          ? 'border-l-2 border-l-[var(--color-cortex-emerald)] bg-[var(--color-bg-tertiary)] text-[var(--color-text-primary)]'
          : 'border-l-2 border-l-transparent text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)] hover:text-[var(--color-text-primary)]'}"
        title={collapsed ? $t(item.label) : undefined}
      >
        <item.icon size={20} />
        {#if !collapsed}
          <span>{$t(item.label)}</span>
        {/if}
      </a>
    {/each}

    <!-- Plugin items -->
    {#if plugins.length > 0}
      <div class="mx-3 my-2 border-t border-[var(--color-border)]"></div>
    {/if}
    {#each pluginItems as item}
      {@const active = isActive(item.href)}
      <a
        href={item.href}
        class="flex items-center gap-3 rounded-[var(--radius-md)] px-3 py-2 text-sm font-medium transition-colors
          {active
          ? 'border-l-2 border-l-[var(--color-cortex-emerald)] bg-[var(--color-bg-tertiary)] text-[var(--color-text-primary)]'
          : 'border-l-2 border-l-transparent text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)] hover:text-[var(--color-text-primary)]'}"
        title={collapsed ? item.label : undefined}
      >
        <span style="color: {item.color ?? 'currentColor'}">
          <item.icon size={20} />
        </span>
        {#if !collapsed}
          <span>{item.label}</span>
        {/if}
      </a>
    {/each}

    <!-- Spacer -->
    <div class="flex-1"></div>

    <!-- Bottom items -->
    {#each staticBottom as item}
      {@const active = isActive(item.href)}
      <a
        href={item.href}
        class="flex items-center gap-3 rounded-[var(--radius-md)] px-3 py-2 text-sm font-medium transition-colors
          {active
          ? 'border-l-2 border-l-[var(--color-cortex-emerald)] bg-[var(--color-bg-tertiary)] text-[var(--color-text-primary)]'
          : 'border-l-2 border-l-transparent text-[var(--color-text-secondary)] hover:bg-[var(--color-bg-tertiary)] hover:text-[var(--color-text-primary)]'}"
        title={collapsed ? $t(item.label) : undefined}
      >
        <item.icon size={20} />
        {#if !collapsed}
          <span>{$t(item.label)}</span>
        {/if}
      </a>
    {/each}
  </nav>

  <!-- Footer signature -->
  {#if !collapsed}
    <div class="border-t border-[var(--color-border)] px-4 py-4">
      <p class="text-center text-sm text-[var(--color-text-secondary)]">
        {$t('footer.madeWith')} &#129504; {$t('footer.by')}
        <a
          href="https://alvarotc.com"
          target="_blank"
          rel="noopener noreferrer"
          class="transition-colors hover:text-[var(--color-brand-blue)]"
        >
          Alvaro Torres
        </a>
      </p>
    </div>
  {/if}
</aside>
