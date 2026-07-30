import { apiFetch } from './client';
import type {
  MediaAsset,
  MediaCompletion,
  MediaUploadSignature
} from '$lib/core/domain/media';

export class MediaApi {
  sign(signal?: AbortSignal): Promise<MediaUploadSignature> {
    return apiFetch('/admin/v1/media/sign', { method: 'POST', signal });
  }

  complete(completion: MediaCompletion, signal?: AbortSignal): Promise<MediaAsset> {
    return apiFetch('/admin/v1/media/complete', {
      method: 'POST',
      headers: { 'content-type': 'application/json' },
      body: JSON.stringify(completion),
      signal
    });
  }
}
