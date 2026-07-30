# Backend operations

## Production configuration

Set `APP_ENV=production`, disable `AUTH_DEV_MODE`, use a non-default `JWT_SECRET`, a production `DATABASE_URL`, and HTTPS-only CORS origins. If `CLOUDINARY_ENABLED=true`, all `CLOUDINARY_*` variables and a fixed `CLOUDINARY_FOLDER` are mandatory. Startup fails before connecting to the database when these checks fail.

`ERROR_MONITOR_DSN` is optional. When set, recovered HTTP panics are captured by the log monitor with the request ID; when empty, the no-op monitor is used. Replace the log adapter with a hosted provider adapter without changing middleware or use cases.

## Migrations

Back up the database before applying migrations. From `backend/api`, run `goose -dir migrations postgres "$DATABASE_URL" status`, then `goose -dir migrations postgres "$DATABASE_URL" up`. Apply migrations once per release, before starting the new API instances. Inspect and resolve a failed migration rather than forcing its version.

## Backup and restore

Create a daily encrypted custom-format backup:

```sh
pg_dump --format=custom --no-owner --file=mq-$(date +%F).dump "$DATABASE_URL"
```

Test restores regularly in an isolated database:

```sh
createdb mq_restore_test
pg_restore --clean --if-exists --no-owner --dbname=mq_restore_test mq-YYYY-MM-DD.dump
```

Confirm migration status, row counts, application login, and representative article/profile reads after restore. Cloudinary assets are external; retain Cloudinary backups/versioning separately and keep database backups because they contain asset ownership and public IDs.

## Uptime

Point an external uptime check at `GET /health` (expect HTTP 200). Alert if the check fails for more than one interval. Keep CORS and cookie domains aligned with the public/admin origins.

## Dependency scanning

- Go: `make govulncheck` (or CI `govulncheck ./...`)
- npm: `npm audit --audit-level=high` in `apps/web` and `apps/admin` (also run in CI)

## Quality gates

```sh
make verify   # unit, arch, openapi, frontend tests + svelte-check
make ci       # verify + Go integration + migrate-check
```

## Release smoke checklist

- `/health` returns 200 and a request ID header.
- Registration/login and password-reset rate limits return 429 after repeated requests.
- Public articles, artists, posts, events, and expanded search return successfully.
- Artist media signing expires quickly, uses the configured folder, and rejects non-image or over-10 MB completion metadata.
- Artist onboarding accepts a valid requested handle and rejects an active duplicate.
- Contributor submission appears in the admin wiki queue; approval publishes or revises the article.
- Repeated analytics views from one `mq_vid` count once per entity per UTC day; no IP address or user agent is persisted.
- Email verification link from registration marks the user verified.
- Admin can list users and change roles without demoting the last admin.
- Event sync sends a summary email when notification prefs allow it.
- Database backup completed and restore procedure was last tested within the required recovery window.
