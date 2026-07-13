# mq

A wiki-style archive and community platform for Ethiopian artists. Artists share knowledge (legal advice, techniques, and more), post artwork, and build shareable public profiles. Institutions and artists are onboarded through a separate admin vetting app. A future event scraper will notify users of art events.

## Stack

| Layer | Tech |
|-------|------|
| Public app | SvelteKit (`apps/web`) |
| Admin app | SvelteKit (`apps/admin`) |
| API | Go + Gin (`backend/api`) |
| Database | PostgreSQL |

## Monorepo layout

```text
mq/
├── apps/
│   ├── web/          # Public wiki, art gallery, artist profiles
│   └── admin/        # Institution/artist vetting & onboarding
├── backend/
│   └── api/          # Hexagonal Go API
├── contracts/
│   └── openapi.yaml  # Shared API contract
├── docker-compose.yml
└── Makefile
```

## Domain contexts

Three content domains are kept separate in the backend:

| Context | Purpose |
|---------|---------|
| `profile` | Who the artist is — bio, slug, contact, social links |
| `art` | What they create — artwork posts, media, publish status |
| `content` | Community wiki — articles, revisions, contributions |

Contexts reference each other by ID only (`artist_id`, `author_id`), never by embedding foreign entities.

Other contexts: `identity` (auth), `onboarding` (approval workflow), `events` (future scraper).

## Hexagonal architecture (backend)

```text
domain/       → pure entities, no internal imports
port/         → inbound (use-case) and outbound (repository) interfaces
usecase/      → business logic
adapter/      → driving (HTTP/Gin) and driven (Postgres, auth, scraper)
```

**Dependency rule:** adapters depend on use cases and ports; the core never depends on adapters.

### Adding a feature

1. Entity in `internal/domain/<context>/`
2. Outbound port in `internal/port/outbound/`
3. Use case in `internal/usecase/<context>/`
4. Postgres adapter + migration
5. Gin handler + DTO
6. Update `contracts/openapi.yaml`
7. Mirror in the relevant SvelteKit app(s)

## Local development

### Prerequisites

- Docker & Docker Compose
- Go 1.23+
- Node.js 20+

### Quick start

```bash
cp .env.example .env
make up          # start PostgreSQL
make migrate     # run migrations (after backend scaffold)
make api         # API on :8080
make web         # public app on :5173
make admin       # admin app on :5174
```

Or run the full stack in Docker (once the backend is scaffolded):

```bash
make up-full
```

## API routes (planned)

| Context | Routes |
|---------|--------|
| `content` | `GET/POST /api/v1/articles`, `GET /api/v1/articles/:slug` |
| `profile` | `GET /api/v1/artists/:slug` |
| `art` | `GET /api/v1/artists/:slug/posts`, `GET/POST /api/v1/posts/:id` |
| `onboarding` | `GET/PUT /admin/v1/applications/:id` |
| `identity` | `GET /api/v1/auth/google`, `GET /api/v1/auth/me`, `POST /api/v1/auth/logout` |

## Authentication

Google OAuth 2.0 (Authorization Code via API BFF) with self-issued JWT session cookies.

1. Configure Google OAuth credentials in [Google Cloud Console](https://console.cloud.google.com/).
2. Set redirect URI to `http://localhost:8080/api/v1/auth/google/callback`.
3. Copy `.env.example` to `.env` and set `GOOGLE_CLIENT_ID` and `GOOGLE_CLIENT_SECRET`.
4. Run `make migrate` to apply `oauth_accounts` and seed the dev admin user (`admin@mq.local`).
5. Sign in from either SvelteKit app via **Sign in with Google**.

| Variable | Purpose |
|----------|---------|
| `JWT_SECRET` | Signs access tokens and OAuth state |
| `JWT_ACCESS_TTL` | Access token lifetime (default `1h`) |
| `AUTH_DEV_MODE` | When `true`, allows `X-User-ID` header fallback for tests |
| `OAUTH_CALLBACK_URL` | Google redirect URI registered with the API |

Admin access requires `role = admin` in the database. The seeded dev admin is linked by email only — sign in with Google using an account you promote to admin, or update the seed migration email to match yours.
