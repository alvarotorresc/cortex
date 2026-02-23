import { writable } from 'svelte/store';
import { apiFetch } from '$lib/api';
import type { WidgetLayout } from '$lib/types';

export const widgetLayouts = writable<WidgetLayout[]>([]);
export const editMode = writable<boolean>(false);
export const dashboardLoading = writable<boolean>(false);
export const dashboardError = writable<string | null>(null);

export async function fetchLayout(): Promise<void> {
  dashboardLoading.set(true);
  dashboardError.set(null);

  try {
    const data = await apiFetch<{ data: WidgetLayout[] }>('/dashboard/layout');
    widgetLayouts.set(data.data ?? []);
  } catch (err) {
    dashboardError.set(err instanceof Error ? err.message : 'Failed to load layout');
    widgetLayouts.set([]);
  } finally {
    dashboardLoading.set(false);
  }
}

export async function saveLayout(layouts: WidgetLayout[]): Promise<void> {
  try {
    await apiFetch('/dashboard/layout', {
      method: 'PUT',
      body: JSON.stringify({ layouts }),
    });
    widgetLayouts.set(layouts);
    editMode.set(false);
  } catch (err) {
    throw err;
  }
}

export function toggleEditMode(): void {
  editMode.update((v) => !v);
}
