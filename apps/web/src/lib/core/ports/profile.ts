import type { ArtistProfile, ContactInfo, ProfileStatus, SocialLinks } from '$lib/core/domain/profile';

export interface ProfilePort {
  getArtistBySlug(slug: string): Promise<ArtistProfile>;
  getMyProfile(): Promise<ArtistProfile>;
  updateMyProfile(input: UpdateMyProfileInput): Promise<ArtistProfile>;
}

export interface UpdateMyProfileInput {
  display_name?: string;
  slug?: string;
  handle?: string;
  bio?: string;
  born?: string;
  discipline?: string;
  tagline?: string;
  years_active?: string;
  portrait_url?: string;
  influences?: string[];
  in_residence?: boolean;
  residency_place?: string;
  open_for_commission?: boolean;
  contact?: ContactInfo;
  social?: SocialLinks;
  status?: ProfileStatus;
}
