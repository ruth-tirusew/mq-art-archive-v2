import { apiFetch } from './client';
import type {
  MediaAsset,
  MediaCompletion,
  MediaUploadSignature
} from '$lib/core/domain/media';

export class MediaApi {
  sign(admin = false, signal?: AbortSignal): Promise<MediaUploadSignature> {
    return apiFetch(admin ? '/admin/v1/media/sign' : '/api/v1/me/media/sign', {
      method: 'POST',
      signal
    });
  }

  complete(
    completion: MediaCompletion,
    admin = false,
    signal?: AbortSignal
  ): Promise<MediaAsset> {
    return apiFetch(admin ? '/admin/v1/media/complete' : '/api/v1/me/media/complete', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(completion),
      signal
    });
  }
}
