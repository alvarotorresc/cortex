import { writable } from 'svelte/store';
import { apiFetch } from '$lib/api';
import type { PluginManifest } from '$lib/types';

export const plugins = writable<PluginManifest[]>([]);
export const pluginsLoading = writable<boolean>(false);
export const pluginsError = writable<string | null>(null);

export async function fetchPlugins(): Promise<void> {
  pluginsLoading.set(true);
  pluginsError.set(null);

  try {
    const data = await apiFetch<{ data: PluginManifest[] }>('/plugins');
    plugins.set(data.data ?? []);
  } catch (err) {
    pluginsError.set(err instanceof Error ? err.message : 'Failed to load plugins');
    plugins.set([]);
  } finally {
    pluginsLoading.set(false);
  }
}

export async function uninstallPlugin(pluginId: string): Promise<void> {
  await apiFetch(`/plugins/${pluginId}`, { method: 'DELETE' });
  await fetchPlugins();
}

export async function installPlugin(pluginId: string): Promise<void> {
  await apiFetch(`/plugins/${pluginId}/install`, { method: 'POST' });
  await fetchPlugins();
}

export async function reloadPlugin(pluginId: string): Promise<void> {
  await apiFetch(`/plugins/${pluginId}/reload`, { method: 'POST' });
  await fetchPlugins();
}
