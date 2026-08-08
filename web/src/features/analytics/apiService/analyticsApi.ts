import type {
  CreateMealPayload,
  MealsResponse,
  StatsSummary,
  UpdateMealPayload,
} from '../utils/types';
import { clearSession, getToken } from '../../auth/utils/session';
import { apiUrl } from '../../../utils/apiUrl';

async function parseError(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { error?: string };
    return body.error || `Request failed with ${response.status}`;
  } catch {
    return `Request failed with ${response.status}`;
  }
}

async function authFetch(path: string, init?: RequestInit): Promise<Response> {
  const token = getToken();
  if (!token) {
    throw new Error('unauthorized');
  }

  const headers = new Headers(init?.headers);
  headers.set('Accept', 'application/json');
  headers.set('Authorization', `Bearer ${token}`);
  if (init?.body && !headers.has('Content-Type')) {
    headers.set('Content-Type', 'application/json');
  }

  const response = await fetch(apiUrl(path), {
    ...init,
    headers,
  });

  if (!response.ok) {
    if (response.status === 401) {
      clearSession();
    }
    throw new Error(await parseError(response));
  }

  return response;
}

async function request<T>(path: string, query: Record<string, string>): Promise<T> {
  const params = new URLSearchParams(query);
  const response = await authFetch(`${path}?${params.toString()}`);
  return response.json() as Promise<T>;
}

export async function fetchStatsSummary(from: string, to: string): Promise<StatsSummary> {
  return request<StatsSummary>('/api/stats/summary', { from, to });
}

export async function fetchMeals(from: string, to: string, limit = 100): Promise<MealsResponse> {
  return request<MealsResponse>('/api/meals', {
    from,
    to,
    limit: String(limit),
  });
}

export async function createMeal(payload: CreateMealPayload) {
  const response = await authFetch('/api/meals', {
    method: 'POST',
    body: JSON.stringify(payload),
  });
  return response.json();
}

export async function updateMeal(id: number, payload: UpdateMealPayload) {
  const response = await authFetch(`/api/meals/${id}`, {
    method: 'PATCH',
    body: JSON.stringify(payload),
  });
  return response.json();
}

export async function deleteMeal(id: number): Promise<void> {
  await authFetch(`/api/meals/${id}`, { method: 'DELETE' });
}

export async function reorderMeals(date: string, mealIds: number[]): Promise<void> {
  await authFetch('/api/meals/reorder', {
    method: 'PATCH',
    body: JSON.stringify({ date, meal_ids: mealIds }),
  });
}
