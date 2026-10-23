# Product Requirements Document — artiv

## Document info

| | |
|---|---|
| Product | artiv  |
| Author | Ruth Tirusew |
| Status | Living document — reflects what has been built |
| Last updated | 2026-09-28 |

## 1. Overview

artiv is a wiki-style archive and community platform for Ethiopian artists. Artists and institutions build public profiles, publish artwork, and contribute wiki-style knowledge articles (legal advice, techniques, and more). Content is community-authored and moderated: contributors submit or revise articles, admins review and publish them, and readers engage with published content through favorites, highlights, and comments. A separate admin app handles vetting, moderation, and platform settings, including an optional event scraper (RSS/JSON/Telegram) and a weekly digest (email + Telegram).

This PRD describes the product as currently implemented across the monorepo (`apps/web`, `apps/admin`, `backend/api`), not a forward-looking proposal.

## 2. Problem statement

Ethiopian artists lack a shared, trustworthy place to publish verified profiles, showcase work, and pool practical knowledge (legal, technical, career) that is normally scattered across informal channels. Institutions vetting or platforming artists need a moderated, structured source of truth instead of ad-hoc documents. artiv addresses this by combining a verified-identity layer with a collaboratively edited, moderated wiki and a lightweight engagement layer that rewards readers for contributing and returning.

## 3. Users and roles

| Role | Description |
|---|---|
| Reader (anonymous or signed-in) | Browses public wiki articles, artist profiles, artwork, and events; can search across all public content |
| Contributor (signed-in artist/institution) | Submits new wiki articles or revisions to existing ones from their studio; favorites, highlights, and comments on articles; manages their own artwork and profile |
| Admin reviewer | Reviews onboarding applications, moderates wiki submissions (approve/reject with diff view), moderates artwork/artist profiles, manages users/roles, configures scrape settings |
| System (scraper / digest) | Long-running background services that ingest events and send scheduled email/Telegram digests |

## 4. Domain contexts (implemented)

The backend is organized as hexagonal domain contexts, each independently addressable and referencing others by ID only:

| Context | Responsibility |
|---|---|
| `identity` | Registration/login (email+password, Google OAuth), password reset/verify, notification preferences, roles |
| `onboarding` | Artist/institution applications, handle uniqueness validation, admin approve/reject, provisioning on approval |
| `profile` | Artist identity — bio, handle, contact, social links; public and studio views |
| `art` | Artwork posts, media assets, publish status, studio lifecycle, admin moderation |
| `content` | Wiki articles, revisions, community submissions, engagement (favorites, highlights, comments) |
| `events` | Curated and scraped art events, admin review/sync |
| `media` | Cloudinary signed direct uploads (JPEG/PNG/WebP, 10 MB max) |
| `analytics` | Privacy-conscious daily page-view aggregates, deduplicated per visitor ID per day, no IP/user-agent retained |
| `settings` | Admin scrape configuration (interval, channels, keywords) |
| `digest` | Weekly email/Telegram digest of new articles/events/artists; Telegram account linking |

## 5. Feature areas

### 5.1 Identity and onboarding

- Account creation via email/password or Google OAuth, with email verification.
- Onboarding application: account type (artist/institution), display name, requested handle, profile notes.
- Server-side handle uniqueness validation; duplicate or malformed handles are rejected before persistence.
- Admin queue lists pending applications; admin approves or rejects with optional notes.
- Approval activates the profile and grants studio/content-management access; rejection is communicated with a reason and, where policy allows, the user may revise and resubmit.
- Full functional spec and test matrix: `docs/frs-onboarding-test.md`. User-facing flow: `docs/user-onboarding.md`.

### 5.2 Wiki content and moderation

- Public wiki: browsable, searchable articles (`GET /api/v1/search?q=`) covering articles, events, artists, and posts.
- Contributors submit new articles or revisions from their studio (`/studio/wiki`); each submission is diffable against the current published version (`DiffView` component, `diffLines` utility) so reviewers see exactly what changed.
- Admin wiki queue (`apps/admin/routes/wiki/submissions`) lists pending submissions; approving publishes or revises the article, rejecting returns it to the contributor.
- Article revisions are tracked as first-class history, not overwritten in place.
- E2E coverage: `apps/web/e2e/wiki-collaboration.spec.ts`, `apps/admin/e2e/wiki-submissions.spec.ts`.

### 5.3 Reader engagement — favorites, highlights, comments

Built as an `engagement` domain shared across the `content` context:

- **Favorites** — readers save articles for later; a dedicated "saved" view (`/wiki/saved`) lists a user's favorited articles. Favoriting is idempotent per user/article pair (migration `00035_article_favorites.sql`).
- **Highlights** — readers select article text and highlight a passage; highlight positions are anchored to stable text offsets (`textAnchor.ts`) so they survive minor article edits, and a heatmap aggregates highlight density across an article (`engagement/heatmap.go`).
- **Comments** — threaded comments anchored to a specific highlighted passage (`CommentThread.svelte`, `SelectionToolbar.svelte`), backed by an anchor model shared with highlights (`domain/engagement/anchor.go`).
- **Activity tracking** — a signed-in user's own account activity (favorites, highlights, comments they've made) is surfaced under `/account/activity`, with a profile dropdown and account layout added to support it.
- Data model: migration `00037_article_highlights_and_comments.sql`; API surface added to `contracts/openapi.yaml` across three commits (favorites, highlights/comments, activity).

### 5.4 Artist profiles and artwork

- Public artist profiles by slug/handle; studio view for the artist to manage their own profile (`/me/profile`).
- Artwork posts with a publish lifecycle (draft → published), managed from studio, moderated by admin.
- Cloudinary-backed signed direct uploads for media, gated behind `CLOUDINARY_ENABLED`.

### 5.5 Events

- Public event listing/detail, sourced from curated admin entries and/or an optional scraper (RSS/JSON feeds, Telegram channels), off by default and enabled per environment via `SCRAPE_ENABLED`/`TELEGRAM_ENABLED`.
- Admin reviews and syncs scraped events before they go public.
- Scrape interval, channels, and keywords are runtime-configurable from the admin settings UI (source of truth in `scrape_settings`, not env vars, after first boot).

### 5.6 Weekly digest and Telegram bot

- `cmd/digest` sends a weekly summary (new articles/events/artists) by email (Resend) and/or Telegram, on a fixed weekday/hour rather than a polling interval.
- Sends are idempotent per recipient per channel (`digest_runs`/`digest_deliveries`), so a crash mid-send resumes rather than duplicates on the next scheduled trigger.
- Telegram account linking uses a separate Bot API credential from the scraper's MTProto session; one-time linking tokens expire after 15 minutes.
- Email unsubscribe uses a stateless, non-expiring HMAC token so the link never breaks.

### 5.7 Analytics

- Privacy-conscious page-view tracking: one view per entity per visitor ID (`mq_vid`) per UTC day; no IP address or user agent is persisted.
- Studio and admin can query aggregated view counts.

## 6. Non-functional requirements

- **Architecture** — hexagonal (domain / port / usecase / adapter) in the Go API; adapters depend on use cases, never the reverse. Feature work follows a test-first flow: update `contracts/openapi.yaml` and write a failing test before implementation, then mirror the change in the relevant SvelteKit app(s).
- **Security** — `APP_ENV=production` enables strict startup config validation (fails fast rather than connecting to the database with bad config); `AUTH_DEV_MODE` must be disabled in production; JWT-signed sessions; rate limiting on auth endpoints (429 after repeated attempts).
- **Data protection** — Telegram MTProto session files are never committed (gitignored); Cloudinary assets are backed up separately from the database; database backups are encrypted, daily, custom-format `pg_dump`, with restores tested regularly in an isolated database.
- **Observability** — `GET /health` for uptime checks; optional error monitor (`ERROR_MONITOR_DSN`) captures recovered HTTP panics with request IDs; scrape/digest health is checked via `last_run_at`/`last_success_at`/`last_error`, not inferred from data freshness.
- **Testing** — `make verify` (unit + arch + OpenAPI + frontend) and `make ci` (adds Go integration + migrate-check) gate merges; Playwright e2e for both web and admin apps.

## 7. Success criteria (by feature area)

- Onboarding: a user can register, verify, submit an application, and — once approved — gain studio access; duplicate handles are always rejected; rejections are clear and, where allowed, resubmittable. Full matrix in `docs/frs-onboarding-test.md`.
- Wiki: a contributor's submission reaches the admin queue with a correct diff; approval publishes or revises the live article without data loss.
- Engagement: favorites, highlights, and comments persist correctly per user/article, survive minor article edits (anchor stability), and are visible in a user's own activity view.
- Digest/Telegram: linking, unlinking, and unsubscribing all behave correctly per the release smoke checklist in `docs/operations.md`; a scheduled send produces a completed `digest_runs` row with matching deliveries and never double-sends on retry.
- Analytics: repeated views from one visitor count once per entity per UTC day, with no PII retained.
- Operations: release smoke checklist (`docs/operations.md`) passes before every production release; backup/restore is tested within the required recovery window.

## 8. Out of scope / not yet built

- Public API rate limiting beyond auth endpoints.
- Article versioning UI for readers (revision history is tracked but not yet exposed publicly).
- Real-time (websocket) updates for comments/highlights — current model is request/response.
- Multi-language content.

