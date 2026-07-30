export const IMAGE_MIME_TYPES = ['image/jpeg', 'image/png', 'image/webp'] as const;
export const MAX_IMAGE_BYTES = 10 * 1024 * 1024;

export interface MediaUploadSignature {
  timestamp: number;
  signature: string;
  cloud_name: string;
  api_key: string;
  folder: string;
  public_id: string;
  expire_at: string;
}

export interface MediaCompletion {
  public_id: string;
  secure_url: string;
  resource_type: string;
  format: string;
  width: number;
  height: number;
  bytes: number;
}

export interface MediaAsset extends MediaCompletion {
  id?: string;
}

export interface UploadedMedia {
  secure_url: string;
  id?: string;
}
