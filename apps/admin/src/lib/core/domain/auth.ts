import { PUBLIC_API_URL } from '$env/static/public';

export interface User {
  id: string;
  email: string;
  role: string;
}

export function googleLoginUrl(): string {
  const returnTo = `${window.location.origin}/auth/callback`;
  return `${PUBLIC_API_URL}/api/v1/auth/google?return_to=${encodeURIComponent(returnTo)}`;
}
