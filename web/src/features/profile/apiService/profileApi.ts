import { clearSession, getToken } from '../../auth/utils/session';
import type { ProfileInput, UserProfile } from '../utils/types';
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

export async function fetchProfile(): Promise<UserProfile> {
  const response = await authFetch('/api/profile');
  return response.json() as Promise<UserProfile>;
}

export async function updateProfile(input: ProfileInput): Promise<UserProfile> {
  const response = await authFetch('/api/profile', {
    method: 'PATCH',
    body: JSON.stringify(input),
  });
  return response.json() as Promise<UserProfile>;
}
