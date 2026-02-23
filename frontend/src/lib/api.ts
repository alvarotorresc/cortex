const BASE = '/api';

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await fetch(`${BASE}${path}`, {
    headers: {
      'Content-Type': 'application/json',
      ...init?.headers,
    },
    ...init,
  });

  if (!response.ok) {
    const error = await response.json().catch(() => ({ error: { message: 'Request failed' } }));
    throw new Error(error.error?.message ?? `HTTP ${response.status}`);
  }

  return response.json();
}

export function pluginApi(pluginId: string) {
  return {
    fetch<T>(path: string, init?: RequestInit): Promise<T> {
      return apiFetch<T>(`/plugins/${pluginId}${path}`, init);
    },
    widget<T>(slot: string): Promise<T> {
      return apiFetch<T>(`/plugins/${pluginId}/widget/${slot}`);
    },
  };
}
