import { apiFetch } from './client';
import type {
  MediaAsset,
  MediaCompletion,
  MediaScope,
  MediaUploadSignature
} from '$lib/core/domain/media';

// Artists, admins and applicants (who aren't artists yet) each have their own sign/complete pair.
const BASE: Record<MediaScope, string> = {
  artist: '/api/v1/me/media',
  admin: '/admin/v1/media',
  applicant: '/api/v1/me/application/media'
};

export class MediaApi {
  sign(scope: MediaScope = 'artist', signal?: AbortSignal): Promise<MediaUploadSignature> {
    return apiFetch(`${BASE[scope]}/sign`, { method: 'POST', signal });
  }

  complete(
    completion: MediaCompletion,
    scope: MediaScope = 'artist',
    signal?: AbortSignal
  ): Promise<MediaAsset> {
    return apiFetch(`${BASE[scope]}/complete`, {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(completion),
      signal
    });
  }
}
