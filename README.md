# mq

A wiki-style archive and community platform for Ethiopian artists. Artists share knowledge (legal advice, techniques, and more), post artwork, and build shareable public profiles. Institutions and artists are onboarded through a separate admin vetting app. An event scraper can ingest RSS/JSON feeds and optional Telegram channels.

## Stack

| Layer | Tech |
|-------|------|
| Public app | SvelteKit (`apps/web`) |
| Admin app | SvelteKit (`apps/admin`) |
| API | Go + Gin (`backend/api`) |
| Database | PostgreSQL |
| Media (pilot) | Cloudinary signed uploads |

## Monorepo layout

```text
mq/
├── apps/
│   ├── web/          # Public wiki, art gallery, artist profiles, studio
│   └── admin/        # Vetting, moderation, settings
├── backend/
│   └── api/          # Hexagonal Go API + scraper workers
├── contracts/
│   └── openapi.yaml  # Shared API contract
├── docs/
│   └── operations.md # Backup, migrate, production smoke
├── scripts/          # openapi + migrate checks
├── docker-compose.yml
└── Makefile
```

**Note:** sibling repos such as `mq-art-archive-svelte` / `mq-art-archive-main` are retired references. This monorepo is the source of truth. The public app loads all product content from the API (no hardcoded catalogs).

## Domain contexts

| Context | Purpose |
|---------|---------|
| `profile` | Artist identity — bio, handle, contact, social links |
| `art` | Artwork posts, media assets, publish status |
| `content` | Wiki articles, revisions, community submissions |
| `identity` | Auth (Google OAuth, email/password, reset, verify) |
| `onboarding` | Artist/institution applications and provisioning |
| `events` | Curated + scraped art events |
| `media` | Cloudinary signed upload + asset metadata |
| `analytics` | Privacy-conscious daily page-view aggregates |
| `settings` | Scrape configuration |

Contexts reference each other by ID only (`artist_id`, `author_id`), never by embedding foreign entities.

## Hexagonal architecture (backend)

```text
domain/       → pure entities, no internal imports
port/         → inbound (use-case) and outbound (repository) interfaces
usecase/      → business logic
adapter/      → driving (HTTP/Gin) and driven (Postgres, auth, scraper, Cloudinary)
```

**Dependency rule:** adapters depend on use cases and ports; the core never depends on adapters.

### Adding a feature (test-first)

1. Update `contracts/openapi.yaml` and write a failing test
2. Entity in `internal/domain/<context>/`
3. Outbound port in `internal/port/outbound/`
4. Use case in `internal/usecase/<context>/`
5. Postgres adapter + migration
6. Gin handler + DTO
7. Mirror in the relevant SvelteKit app(s)
8. Run `make verify` (and `make ci` before merge)

## Local development

### Prerequisites

- Docker & Docker Compose
- Go 1.25+ (or use the Makefile Docker fallback)
- Node.js 20+

### Quick start

```bash
cp .env.example .env
make up          # start PostgreSQL
make migrate     # run database migrations
make api         # API on :8080
make web         # public app on :5173
make admin       # admin app on :5174
```

Full stack in Docker:

```bash
make up-full
```

Set `PUBLIC_API_URL=http://localhost:8080` in `apps/web/.env` and `apps/admin/.env` (copied from examples via `make env`).

## API surfaces (implemented)

| Context | Routes |
|---------|--------|
| `content` | Public articles; admin CRUD + revisions; community `/me/wiki/submissions` |
| `profile` | Artists by slug/handle; studio `/me/profile`; admin artist moderation |
| `art` | Posts list/detail; studio post lifecycle; admin post moderation |
| `onboarding` | `POST /applications` (artist/institution + requested handle); admin review |
| `identity` | Google OAuth, register/login, password reset, notification prefs |
| `events` | Public list/detail; admin review + sync |
| `search` | `GET /api/v1/search?q=` — articles, events, artists, posts |
| `media` | `/me/media/sign|complete` and admin equivalents (Cloudinary) |
| `analytics` | `POST /analytics/view`; studio/admin query |
| `settings` | Admin scrape settings |

### Event scraper

Configure RSS/JSON feeds and optional Telegram channels in `.env` (`SCRAPE_*`, `TELEGRAM_*`). Off by default.

```bash
cd backend/api && go run ./cmd/telegram-login
cd backend/api && go run ./cmd/scraper
```

### Cloudinary (pilot)

Enable with `CLOUDINARY_ENABLED=true` and set cloud name, API key/secret, and folder. Studio and admin use signed direct uploads (JPEG/PNG/WebP, 10 MB max).

## Testing & quality gates

```bash
make verify              # unit + arch + openapi + frontend tests/checks
make ci                  # verify + integration + migrate-check
make test                # Go unit
make test-integration    # Go integration (Docker)
make openapi-check
make migrate-check
make web-test / admin-test
make web-test-e2e / admin-test-e2e
```

Operations: [`docs/operations.md`](docs/operations.md)

## Authentication

- Google OAuth 2.0 (Authorization Code via API BFF) with JWT session cookies
- Email/password register + login
- Forgot/reset password (Resend or log mailer)

1. Configure Google OAuth credentials and redirect URI `http://localhost:8080/api/v1/auth/google/callback`.
2. Copy `.env.example` → `.env` and set secrets.
3. `make migrate` seeds dev admin `admin@mq.local` / `admin123`.

| Variable | Purpose |
|----------|---------|
| `APP_ENV` | `production` enables strict config validation |
| `JWT_SECRET` | Signs access tokens and OAuth state |
| `AUTH_DEV_MODE` | When `true`, allows `X-User-ID` header fallback for tests |
| `CLOUDINARY_*` | Pilot media uploads |
| `PUBLIC_API_URL` | Required by web/admin for all product content |
