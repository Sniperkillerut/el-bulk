import { apiFetch } from './api';
import { Theme, ThemeInput } from './types';

export async function fetchThemes(): Promise<Theme[]> {
  return apiFetch<Theme[]>('/api/themes');
}

export async function createTheme(input: ThemeInput): Promise<Theme> {
  return apiFetch<Theme>('/api/admin/themes', {
    method: 'POST',
    body: JSON.stringify(input),
  });
}

export async function updateTheme(id: string, input: ThemeInput): Promise<Theme> {
  return apiFetch<Theme>(`/api/admin/themes/${id}`, {
    method: 'PUT',
    body: JSON.stringify(input),
  });
}

export async function deleteTheme(id: string): Promise<void> {
  return apiFetch<void>(`/api/admin/themes/${id}`, {
    method: 'DELETE',
  });
}
