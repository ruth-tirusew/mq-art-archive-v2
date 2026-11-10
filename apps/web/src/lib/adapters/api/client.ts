import { DomainError, kindForStatus } from '$lib/core/domain/errors';
import { getApiBaseUrl } from '$lib/config/dataSource';

/** The HTTP transport's failure. Callers outside the adapters should use `DomainError`'s `kind`. */
export class ApiError extends DomainError {
  constructor(
    message: string,
    readonly status: number
  ) {
    super(message, kindForStatus(status));
    this.name = 'ApiError';
  }
}

type UnauthorizedListener = (path: string) => void;
let unauthorizedListener: UnauthorizedListener | null = null;

/**
 * Called with the request path whenever the API answers 401. The transport only reports it; the
 * application layer decides whether that means the session ended (see application/session.ts).
 */
export function onUnauthorized(listener: UnauthorizedListener | null): void {
  unauthorizedListener = listener;
}

export async function apiFetch<T>(path: string, init?: RequestInit): Promise<T> {
  const response = await apiResponse(path, init);
  if (response.status === 204) {
    return undefined as T;
  }
  return (await response.json()) as T;
}

export async function apiResponse(path: string, init?: RequestInit): Promise<Response> {
  const response = await fetch(`${getApiBaseUrl()}${path}`, {
    credentials: 'include',
    ...init
  });
  if (!response.ok) {
    let message = response.statusText;
    try {
      const body = (await response.json()) as { error?: string };
      if (body.error) message = body.error;
    } catch {
      // ignore JSON parse errors
    }
    if (response.status === 401) unauthorizedListener?.(path);
    throw new ApiError(message, response.status);
  }
  if (response.status === 204) {
    return response;
  }
  return response;
}
