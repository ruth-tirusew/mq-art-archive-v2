export type ArtStatus = 'draft' | 'published' | 'archived';

export interface MediaAsset {
  id: string;
  url: string;
  mime_type?: string;
  width?: number;
  height?: number;
  sort_order?: number;
}

export interface ArtPost {
  id: string;
  artist_id: string;
  artist_slug?: string;
  artist_name?: string;
  title: string;
  description?: string;
  medium?: string;
  year?: number | null;
  dimensions?: string;
  city?: string;
  style?: string;
  residency?: string;
  exhibition?: string;
  featured_acquisition?: boolean;
  palette?: string[];
  media?: MediaAsset[];
  status?: ArtStatus;
  published_at?: string | null;
  created_at?: string;
  updated_at?: string;
}

export interface ArtListParams {
  city?: string;
  medium?: string;
  year?: number;
  style?: string;
  featured?: boolean;
  q?: string;
  limit?: number;
  offset?: number;
}
