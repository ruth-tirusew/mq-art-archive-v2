import { apiFetch, apiResponse } from '$lib/adapters/api/client';
import { normalizeProfile, normalizeProfiles } from '$lib/adapters/api/normalize';
import type { ArtistListParams, ArtistProfile } from '$lib/core/domain/profile';

function toQuery(params?: ArtistListParams): string {
  if (!params) return '';
  const q = new URLSearchParams();
  if (params.q) q.set('q', params.q);
  if (params.featured !== undefined) q.set('featured', String(params.featured));
  if (params.limit) q.set('limit', String(params.limit));
  if (params.offset) q.set('offset', String(params.offset));
  const s = q.toString();
  return s ? `?${s}` : '';
}

export class ArtistsApi {
  list(params?: ArtistListParams): Promise<ArtistProfile[]> {
    return apiFetch<unknown>(`/api/v1/artists${toQuery(params)}`).then(normalizeProfiles);
  }

  async listPage(params?: ArtistListParams): Promise<{ data: ArtistProfile[]; total: number }> {
    const response = await apiResponse(`/api/v1/artists${toQuery(params)}`);
    return {
      data: normalizeProfiles(await response.json()),
      total: Number(response.headers.get('X-Total-Count') ?? 0)
    };
  }

  getByHandle(handle: string): Promise<ArtistProfile> {
    return apiFetch<Record<string, unknown>>(`/api/v1/profiles/@${handle}`).then(normalizeProfile);
  }
}
