import { writable } from 'svelte/store';
import { browser } from '$app/environment';

export const sidebarCollapsed = writable<boolean>(false);
export const mobileSidebarOpen = writable<boolean>(false);

export function toggleSidebar() {
  sidebarCollapsed.update((v) => !v);
}

export function toggleMobileSidebar() {
  mobileSidebarOpen.update((v) => !v);
}

export function closeMobileSidebar() {
  mobileSidebarOpen.set(false);
}

if (browser) {
  const mql = window.matchMedia('(min-width: 1024px)');
  if (!mql.matches) {
    sidebarCollapsed.set(true);
  }
}
