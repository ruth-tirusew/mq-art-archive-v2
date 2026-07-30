export type ArtStatus = 'draft' | 'published' | 'archived';

export interface MediaAsset {
  id?: string;
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
  palette?: string[];
  status?: ArtStatus;
  featured_acquisition?: boolean;
  media?: MediaAsset[];
  created_at?: string;
  updated_at?: string;
  published_at?: string | null;
}

export interface ArtPostWrite {
  title: string;
  description?: string;
  medium?: string;
  year?: number | null;
  dimensions?: string;
  city?: string;
  style?: string;
  palette?: string[];
  media_urls?: string[];
}

export interface ArtPostCreate extends ArtPostWrite {
  artist_id: string;
  status?: ArtStatus;
}
