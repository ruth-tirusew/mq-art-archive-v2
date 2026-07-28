import { browser } from '$app/environment';
import { PUBLIC_API_URL } from '$env/static/public';

export class ApiError extends Error {
  constructor(
    message: string,
    readonly status: number
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const headers = new Headers(init?.headers);

  // Universal +page.ts loaders run on the server without browser cookie jar.
  // Forward the incoming request cookies so HttpOnly auth reaches the API.
  if (!browser && !headers.has('cookie')) {
    try {
      const { getRequestEvent } = await import('$app/server');
      const cookie = getRequestEvent().request.headers.get('cookie');
      if (cookie) headers.set('cookie', cookie);
    } catch {
      // Outside a request context (e.g. build/prerender).
    }
  }

  const response = await fetch(`${PUBLIC_API_URL}${path}`, {
    credentials: 'include',
    ...init,
    headers
  });
  if (!response.ok) {
    let message = response.statusText;
    try {
      const body = (await response.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // ignore JSON parse errors
    }
    throw new ApiError(message, response.status);
  }
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}
