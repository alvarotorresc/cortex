<script lang="ts">
  import '../app.css';
  import '$lib/i18n';
  import { isLoading as i18nLoading } from 'svelte-i18n';
  import { page } from '$app/stores';
  import { onMount } from 'svelte';
  import Sidebar from '$lib/components/Sidebar.svelte';
  import Topbar from '$lib/components/Topbar.svelte';
  import X from 'lucide-svelte/icons/x';
  import {
    sidebarCollapsed,
    mobileSidebarOpen,
    toggleMobileSidebar,
    closeMobileSidebar,
  } from '$lib/stores/sidebar';
  import { plugins, fetchPlugins } from '$lib/stores/plugins';

  let { children } = $props();

  let pluginList = $state<import('$lib/types').PluginManifest[]>([]);
  plugins.subscribe((v) => {
    pluginList = v;
  });

  let mobileOpen = $state(false);
  mobileSidebarOpen.subscribe((v) => {
    mobileOpen = v;
  });

  let collapsed = $state(false);
  sidebarCollapsed.subscribe((v) => {
    collapsed = v;
  });

  let i18nReady = $state(false);
  i18nLoading.subscribe((v) => {
    i18nReady = !v;
  });

  let currentPath = $state('/');
  page.subscribe((p) => {
    currentPath = p.url.pathname;
  });

  onMount(() => {
    fetchPlugins();
  });

  // Close mobile sidebar on navigation
  $effect(() => {
    currentPath;
    closeMobileSidebar();
  });
</script>

{#if !i18nReady}
  <div
    class="flex h-screen items-center justify-center bg-[var(--color-bg-primary)] text-[var(--color-text-secondary)]"
  >
    <p class="text-sm">Loading...</p>
  </div>
{:else}
  <div class="flex h-screen overflow-hidden bg-[var(--color-bg-primary)]">
    <!-- Desktop sidebar -->
    <div class="hidden md:flex">
      <Sidebar plugins={pluginList} />
    </div>

    <!-- Mobile sidebar overlay -->
    {#if mobileOpen}
      <div class="fixed inset-0 z-40 flex md:hidden">
        <!-- Backdrop -->
        <button
          class="fixed inset-0 bg-black/50"
          onclick={closeMobileSidebar}
          aria-label="Close sidebar"
          tabindex="-1"
        ></button>
        <!-- Sidebar panel -->
        <div class="relative z-50 flex">
          <Sidebar plugins={pluginList} />
          <button
            onclick={closeMobileSidebar}
            class="ml-1 mt-2 flex h-8 w-8 items-center justify-center rounded-[var(--radius-md)] bg-[var(--color-bg-secondary)] text-[var(--color-text-secondary)]"
            aria-label="Close sidebar"
          >
            <X size={16} />
          </button>
        </div>
      </div>
    {/if}

    <!-- Content area -->
    <div class="flex flex-1 flex-col overflow-hidden">
      <Topbar onMenuClick={toggleMobileSidebar} />

      <main class="flex-1 overflow-y-auto">
        <div class="mx-auto max-w-[1400px] px-4 py-6 md:px-8 md:py-8">
          {@render children()}
        </div>
      </main>
    </div>
  </div>
{/if}
