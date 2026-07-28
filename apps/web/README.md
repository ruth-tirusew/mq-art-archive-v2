# Artiv — public web (`apps/web`)

SvelteKit public site for the Ethiopian artists wiki/archive. All product content (artists, posts, wiki, events, search) is loaded from the Go API via `PUBLIC_API_URL`. There are no hardcoded product catalogs in the client.

## Stack

- **SvelteKit 2** + **Svelte 5**
- **Tailwind CSS v4**
- **Vitest** + **MSW** + **Playwright**

## Routes

| Path | Description |
|------|-------------|
| `/` | Homepage (API-backed featured/acquisitions) |
| `/archive` | Filterable works catalog |
| `/artists` | Artist roster |
| `/artists/[slug]` | Artist profile + timeline |
| `/events`, `/events/[slug]` | Events |
| `/wiki`, `/wiki/[slug]` | Community handbook |
| `/portfolio` | Marketing + featured API profile |
| `/@[handle]` | Shareable link-in-bio |
| `/studio/*` | Artist studio (profile, posts, wiki submissions) |
| `/apply` | Artist or institution application (+ requested `@handle`) |
| `/login`, `/forgot-password`, `/reset-password` | Auth |
| `/about` | About |

## Development

```bash
# from repo root
make env
make web   # http://localhost:5173
```

Requires `PUBLIC_API_URL` (see `.env.example`). API failures render explicit unavailable/empty states — never fabricated content.

## Testing

| Command | What it runs |
|---------|----------------|
| `npm run test` / `make web-test` | Vitest unit + MSW integration |
| `npm run test:unit` | Unit only |
| `npm run test:integration` | MSW integration |
| `npm run test:integration:live` | Live API (`LIVE_API=1`) |
| `npm run test:e2e:smoke` / `make web-test-e2e` | Playwright `@smoke` |
| `npm run check` | `svelte-check` |

### Conventions

- **Unit:** `*.test.ts`
- **MSW integration:** `*.integration.test.ts`
- **Live integration:** `*.live.integration.test.ts` (skipped unless `LIVE_API=1`)
- **E2E tags:** `@smoke`, `@auth`
- **Selectors:** `data-testid` uses `web-*` prefix

## Relation to other projects

`mq-art-archive-main` / `mq-art-archive-svelte` are retired prototypes. Use this app + `backend/api` + Postgres seeds.
