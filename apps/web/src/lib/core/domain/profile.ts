export type ProfileStatus = 'draft' | 'pending' | 'approved';

export interface ContactInfo {
  email?: string;
  phone?: string;
  website?: string;
  location?: string;
}

export interface SocialLinks {
  instagram?: string;
  twitter?: string;
  telegram?: string;
}

export interface ArtistProfile {
  id: string;
  user_id?: string;
  slug: string;
  handle?: string;
  display_name: string;
  bio?: string;
  born?: string;
  discipline?: string;
  tagline?: string;
  years_active?: string;
  portrait_url?: string;
  featured?: boolean;
  influences?: string[];
  in_residence?: boolean;
  residency_place?: string;
  open_for_commission?: boolean;
  contact?: ContactInfo;
  social?: SocialLinks;
  status?: ProfileStatus;
  approved_at?: string | null;
  created_at?: string;
  updated_at?: string;
}

export interface ArtistListParams {
  q?: string;
  featured?: boolean;
  limit?: number;
  offset?: number;
}
