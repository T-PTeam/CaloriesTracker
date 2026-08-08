import { clearSession, getToken, saveSession, saveStoredUser } from '../utils/session';
import type { AuthResult, PublicUser } from '../utils/types';
import { apiUrl } from '../../../utils/apiUrl';
import type { Locale } from '../../../i18n/messages';

async function parseError(response: Response): Promise<string> {
  try {
    const body = (await response.json()) as { error?: string };
    return body.error || `Request failed with ${response.status}`;
  } catch {
    return `Request failed with ${response.status}`;
  }
}

async function authRequest(path: string, body: Record<string, string>): Promise<AuthResult> {
  const response = await fetch(apiUrl(path), {
    method: 'POST',
    headers: {
      Accept: 'application/json',
      'Content-Type': 'application/json',
    },
    body: JSON.stringify(body),
  });

  if (!response.ok) {
    throw new Error(await parseError(response));
  }

  const result = (await response.json()) as AuthResult;
  saveSession(result);
  return result;
}

export async function register(
  email: string,
  password: string,
  linkCode: string,
): Promise<AuthResult> {
  return authRequest('/api/auth/register', {
    email,
    password,
    link_code: linkCode,
  });
}

export async function login(email: string, password: string): Promise<AuthResult> {
  return authRequest('/api/auth/login', { email, password });
}

export async function fetchMe(): Promise<PublicUser> {
  const token = getToken();
  if (!token) {
    throw new Error('unauthorized');
  }

  const response = await fetch(apiUrl('/api/auth/me'), {
    headers: {
      Accept: 'application/json',
      Authorization: `Bearer ${token}`,
    },
  });

  if (!response.ok) {
    if (response.status === 401) {
      clearSession();
    }
    throw new Error(await parseError(response));
  }

  const user = (await response.json()) as PublicUser;
  saveStoredUser(user);
  return user;
}

export async function updateLanguage(language: Locale): Promise<PublicUser> {
  const token = getToken();
  if (!token) {
    throw new Error('unauthorized');
  }

  const response = await fetch(apiUrl('/api/auth/language'), {
    method: 'PATCH',
    headers: {
      Accept: 'application/json',
      Authorization: `Bearer ${token}`,
      'Content-Type': 'application/json',
    },
    body: JSON.stringify({ language }),
  });

  if (!response.ok) {
    if (response.status === 401) {
      clearSession();
    }
    throw new Error(await parseError(response));
  }

  const user = (await response.json()) as PublicUser;
  saveStoredUser(user);
  return user;
}

export function logout(): void {
  clearSession();
}
