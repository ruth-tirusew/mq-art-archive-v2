import { getApiBaseUrl } from '$lib/config/dataSource';

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
    throw new ApiError(message, response.status);
  }
  if (response.status === 204) {
    return response;
  }
  return response;
}
