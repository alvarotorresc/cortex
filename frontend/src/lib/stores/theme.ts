import { writable } from 'svelte/store';
import { browser } from '$app/environment';

type Theme = 'light' | 'dark';

function getInitialTheme(): Theme {
  if (!browser) return 'light';
  const saved = localStorage.getItem('cortex-theme');
  if (saved === 'light' || saved === 'dark') return saved;
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

export const theme = writable<Theme>(getInitialTheme());

export function applyTheme(t: Theme) {
  if (!browser) return;

  const root = document.documentElement;
  root.classList.remove('light', 'dark');
  root.classList.add(t);

  localStorage.setItem('cortex-theme', t);
  theme.set(t);
}

export function toggleTheme() {
  let current: Theme = 'light';
  theme.subscribe((v) => (current = v))();
  applyTheme(current === 'light' ? 'dark' : 'light');
}

if (browser) {
  applyTheme(getInitialTheme());
}
