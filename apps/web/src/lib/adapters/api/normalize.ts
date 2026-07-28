import type { ArtPost } from '$lib/core/domain/art';
import type { ArtistProfile } from '$lib/core/domain/profile';
import type { Article } from '$lib/core/domain/content';
import type { Event } from '$lib/core/domain/events';

type RawRecord = Record<string, unknown>;

function str(value: unknown): string | undefined {
  return typeof value === 'string' ? value : undefined;
}

function num(value: unknown): number | undefined {
  return typeof value === 'number' ? value : undefined;
}

function bool(value: unknown): boolean | undefined {
  return typeof value === 'boolean' ? value : undefined;
}

export function normalizeProfile(raw: RawRecord): ArtistProfile {
  const contact = (raw.contact ?? raw.Contact) as RawRecord | undefined;
  const social = (raw.social ?? raw.Social) as RawRecord | undefined;

  return {
    id: String(raw.id ?? raw.ID ?? ''),
    user_id: str(raw.user_id ?? raw.UserID),
    slug: String(raw.slug ?? raw.Slug ?? ''),
    handle: str(raw.handle ?? raw.Handle),
    display_name: String(raw.display_name ?? raw.DisplayName ?? ''),
    bio: str(raw.bio ?? raw.Bio),
    born: str(raw.born ?? raw.Born),
    discipline: str(raw.discipline ?? raw.Discipline),
    tagline: str(raw.tagline ?? raw.Tagline),
    years_active: str(raw.years_active ?? raw.YearsActive),
    portrait_url: str(raw.portrait_url ?? raw.PortraitURL),
    featured: bool(raw.featured ?? raw.Featured),
    influences: Array.isArray(raw.influences ?? raw.Influences)
      ? ((raw.influences ?? raw.Influences) as string[])
      : [],
    in_residence: bool(raw.in_residence ?? raw.InResidence) ?? false,
    residency_place: str(raw.residency_place ?? raw.ResidencyPlace),
    open_for_commission: bool(raw.open_for_commission ?? raw.OpenForCommission) ?? false,
    contact: contact
      ? {
          email: str(contact.email ?? contact.Email),
          phone: str(contact.phone ?? contact.Phone),
          website: str(contact.website ?? contact.Website),
          location: str(contact.location ?? contact.Location)
        }
      : undefined,
    social: social
      ? {
          instagram: str(social.instagram ?? social.Instagram),
          twitter: str(social.twitter ?? social.Twitter),
          telegram: str(social.telegram ?? social.Telegram)
        }
      : undefined,
    status: (raw.status ?? raw.Status) as ArtistProfile['status'],
    created_at: str(raw.created_at ?? raw.CreatedAt),
    updated_at: str(raw.updated_at ?? raw.UpdatedAt)
  };
}

export function normalizeProfiles(raw: unknown): ArtistProfile[] {
  if (!Array.isArray(raw)) return [];
  return raw.map((item) => normalizeProfile(item as RawRecord));
}

export function normalizeArtPost(raw: RawRecord): ArtPost {
  const mediaRaw = (raw.media ?? raw.Media) as RawRecord[] | undefined;
  const paletteRaw = raw.palette ?? raw.Palette;

  return {
    id: String(raw.id ?? raw.ID ?? ''),
    artist_id: String(raw.artist_id ?? raw.ArtistID ?? ''),
    artist_slug: str(raw.artist_slug ?? raw.ArtistSlug),
    artist_name: str(raw.artist_name ?? raw.ArtistName),
    title: String(raw.title ?? raw.Title ?? ''),
    description: str(raw.description ?? raw.Description),
    medium: str(raw.medium ?? raw.Medium),
    year: num(raw.year ?? raw.Year) ?? null,
    dimensions: str(raw.dimensions ?? raw.Dimensions),
    city: str(raw.city ?? raw.City),
    style: str(raw.style ?? raw.Style),
    residency: str(raw.residency ?? raw.Residency),
    exhibition: str(raw.exhibition ?? raw.Exhibition),
    featured_acquisition: bool(raw.featured_acquisition ?? raw.FeaturedAcquisition),
    palette: Array.isArray(paletteRaw) ? (paletteRaw as string[]) : undefined,
    status: (raw.status ?? raw.Status) as ArtPost['status'],
    published_at: str(raw.published_at ?? raw.PublishedAt) ?? null,
    created_at: str(raw.created_at ?? raw.CreatedAt),
    updated_at: str(raw.updated_at ?? raw.UpdatedAt),
    media: mediaRaw?.map((item) => ({
      id: String(item.id ?? item.ID ?? ''),
      url: String(item.url ?? item.URL ?? ''),
      mime_type: str(item.mime_type ?? item.MimeType),
      width: num(item.width ?? item.Width),
      height: num(item.height ?? item.Height),
      sort_order: num(item.sort_order ?? item.SortOrder)
    }))
  };
}

export function normalizeArtPosts(raw: unknown): ArtPost[] {
  if (!Array.isArray(raw)) return [];
  return raw.map((item) => normalizeArtPost(item as RawRecord));
}

export function normalizeArticle(raw: RawRecord): Article {
  return {
    id: String(raw.id ?? raw.ID ?? ''),
    slug: String(raw.slug ?? raw.Slug ?? ''),
    title: String(raw.title ?? raw.Title ?? ''),
    body: String(raw.body ?? raw.Body ?? ''),
    category: str(raw.category ?? raw.Category),
    excerpt: str(raw.excerpt ?? raw.Excerpt),
    reading_time: num(raw.reading_time ?? raw.ReadingTime),
    difficulty: str(raw.difficulty ?? raw.Difficulty),
    verified: bool(raw.verified ?? raw.Verified),
    contributors: num(raw.contributors ?? raw.Contributors),
    author_id: String(raw.author_id ?? raw.AuthorID ?? ''),
    status: (raw.status ?? raw.Status) as Article['status'],
    created_at: String(raw.created_at ?? raw.CreatedAt ?? ''),
    updated_at: String(raw.updated_at ?? raw.UpdatedAt ?? '')
  };
}

export function normalizeArticles(raw: unknown): Article[] {
  if (!Array.isArray(raw)) return [];
  return raw.map((item) => normalizeArticle(item as RawRecord));
}

export function normalizeEvent(raw: RawRecord): Event {
  return {
    id: String(raw.id ?? raw.ID ?? ''),
    slug: String(raw.slug ?? raw.Slug ?? ''),
    title: String(raw.title ?? raw.Title ?? ''),
    description: str(raw.description ?? raw.Description),
    image_url: str(raw.image_url ?? raw.ImageURL) ?? null,
    source_url: str(raw.source_url ?? raw.SourceURL) ?? null,
    event_type: String(raw.event_type ?? raw.EventType ?? 'Event'),
    venue: str(raw.venue ?? raw.Venue),
    city: str(raw.city ?? raw.City),
    starts_at: String(raw.starts_at ?? raw.StartsAt ?? ''),
    ends_at: str(raw.ends_at ?? raw.EndsAt) ?? null,
    status: (raw.status ?? raw.Status) as Event['status'],
    created_at: str(raw.created_at ?? raw.CreatedAt),
    updated_at: str(raw.updated_at ?? raw.UpdatedAt)
  };
}

export function normalizeEvents(raw: unknown): Event[] {
  if (!Array.isArray(raw)) return [];
  return raw.map((item) => normalizeEvent(item as RawRecord));
}
