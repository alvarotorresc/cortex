import { writable } from 'svelte/store';
import { browser } from '$app/environment';

type Theme = 'light' | 'dark' | 'system';

function getInitialTheme(): Theme {
  if (!browser) return 'system';
  return (localStorage.getItem('cortex-theme') as Theme) ?? 'system';
}

function getResolvedTheme(theme: Theme): 'light' | 'dark' {
  if (theme !== 'system') return theme;
  if (!browser) return 'light';
  return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
}

export const theme = writable<Theme>(getInitialTheme());
export const resolvedTheme = writable<'light' | 'dark'>(getResolvedTheme(getInitialTheme()));

export function applyTheme(t: Theme) {
  if (!browser) return;

  const resolved = getResolvedTheme(t);
  const root = document.documentElement;
  root.classList.remove('light', 'dark');
  root.classList.add(resolved);

  localStorage.setItem('cortex-theme', t);
  theme.set(t);
  resolvedTheme.set(resolved);
}

export function toggleTheme() {
  let current = 'system' as string;
  theme.subscribe((v) => (current = v))();

  const order: Theme[] = ['light', 'dark', 'system'];
  const idx = order.indexOf(current as Theme);
  const next = order[(idx + 1) % order.length];

  applyTheme(next);
}

if (browser) {
  applyTheme(getInitialTheme());

  window.matchMedia('(prefers-color-scheme: dark)').addEventListener('change', () => {
    let current: Theme = 'system';
    theme.subscribe((v) => (current = v))();
    if (current === 'system') {
      applyTheme('system');
    }
  });
}
