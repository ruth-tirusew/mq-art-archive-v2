# mq admin

SvelteKit admin console for reviewing applications, artists, posts, and events.

## Development

```bash
npm install
npm run dev
```

From the repo root, `make admin` runs on **http://localhost:5174**.

Requires `PUBLIC_API_URL` (see `.env.example`) pointing at the Go API.

## Testing

| Command | What it runs |
|---------|----------------|
| `npm run test` / `make admin-test` | Vitest unit + MSW integration |
| `npm run test:unit` | Unit only (`*.test.ts`) |
| `npm run test:integration` | MSW integration (`*.integration.test.ts`) |
| `npm run test:integration:live` | Live API (`*.live.integration.test.ts`, needs `LIVE_API=1`) |
| `npm run test:e2e:smoke` / `make admin-test-e2e` | Playwright `@smoke` |

Live API prerequisites: `make up`, `make migrate`, `make api`. Then:

```bash
make admin-test-live
```

Admin live/e2e against authenticated routes may need `AUTH_DEV_MODE` (default in `.env.example`) and the seeded admin user id via `X-User-ID` / `LIVE_USER_ID` (`00000000-0000-4000-8000-000000000001`). Smoke e2e covers `/login` and the unauthenticated redirect without a session.

### Conventions

- **Unit:** `*.test.ts`
- **MSW integration:** `*.integration.test.ts`
- **Live integration:** `*.live.integration.test.ts` (skipped unless `LIVE_API=1`)
- **E2E tags:** Playwright `@smoke`, `@auth`, `@admin`
- **Selectors:** prefer roles/labels; `data-testid` uses `admin-*` prefix where needed
