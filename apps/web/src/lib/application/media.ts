import { MediaApi } from '$lib/adapters/api/mediaApi';
import {
  IMAGE_MIME_TYPES,
  MAX_IMAGE_BYTES,
  type MediaCompletion,
  type UploadedMedia
} from '$lib/core/domain/media';

const api = new MediaApi();

export interface UploadImageOptions {
  admin?: boolean;
  signal?: AbortSignal;
  onProgress?: (progress: number) => void;
}

export function validateImageFile(file: Pick<File, 'size' | 'type'>): string | null {
  if (!IMAGE_MIME_TYPES.includes(file.type as (typeof IMAGE_MIME_TYPES)[number])) {
    return 'Choose a JPEG, PNG, or WebP image.';
  }
  if (file.size > MAX_IMAGE_BYTES) {
    return 'Image must be 10 MB or smaller.';
  }
  return null;
}

function uploadToCloudinary(
  file: File,
  signature: Awaited<ReturnType<MediaApi['sign']>>,
  options: UploadImageOptions
): Promise<MediaCompletion> {
  return new Promise((resolve, reject) => {
    const xhr = new XMLHttpRequest();
    const abort = () => xhr.abort();
    const cleanup = () => options.signal?.removeEventListener('abort', abort);

    xhr.open(
      'POST',
      `https://api.cloudinary.com/v1_1/${encodeURIComponent(signature.cloud_name)}/image/upload`
    );
    xhr.upload.onprogress = (event) => {
      if (event.lengthComputable) options.onProgress?.(Math.round((event.loaded / event.total) * 100));
    };
    xhr.onerror = () => {
      cleanup();
      reject(new Error('Image upload failed.'));
    };
    xhr.onabort = () => {
      cleanup();
      reject(new DOMException('Upload cancelled.', 'AbortError'));
    };
    xhr.onload = () => {
      cleanup();
      if (xhr.status < 200 || xhr.status >= 300) {
        reject(new Error('Cloudinary rejected the image upload.'));
        return;
      }
      try {
        const result = JSON.parse(xhr.responseText) as MediaCompletion;
        resolve({
          public_id: result.public_id,
          secure_url: result.secure_url,
          resource_type: result.resource_type,
          format: result.format,
          width: result.width,
          height: result.height,
          bytes: result.bytes
        });
      } catch {
        reject(new Error('Cloudinary returned an invalid response.'));
      }
    };

    const form = new FormData();
    form.append('file', file);
    form.append('api_key', signature.api_key);
    form.append('timestamp', String(signature.timestamp));
    form.append('signature', signature.signature);
    form.append('folder', signature.folder);
    form.append('public_id', signature.public_id);
    options.signal?.addEventListener('abort', abort, { once: true });
    if (options.signal?.aborted) {
      xhr.abort();
      return;
    }
    xhr.send(form);
  });
}

export async function uploadImage(
  file: File,
  options: UploadImageOptions = {}
): Promise<UploadedMedia> {
  const validationError = validateImageFile(file);
  if (validationError) throw new Error(validationError);

  options.onProgress?.(0);
  const signature = await api.sign(options.admin, options.signal);
  const completion = await uploadToCloudinary(file, signature, options);
  options.onProgress?.(100);
  const asset = await api.complete(completion, options.admin, options.signal);
  return { secure_url: asset.secure_url, id: asset.id };
}
