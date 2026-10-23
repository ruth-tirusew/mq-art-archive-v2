# Backend operations

## Production configuration

Set `APP_ENV=production`, disable `AUTH_DEV_MODE`, use a non-default `JWT_SECRET`, a production `DATABASE_URL`, and HTTPS-only CORS origins. If `CLOUDINARY_ENABLED=true`, all `CLOUDINARY_*` variables and a fixed `CLOUDINARY_FOLDER` are mandatory. Startup fails before connecting to the database when these checks fail.

`ERROR_MONITOR_DSN` is optional. When set, recovered HTTP panics are captured by the log monitor with the request ID; when empty, the no-op monitor is used. Replace the log adapter with a hosted provider adapter without changing middleware or use cases.

## Event scraper

The `scraper` service runs a long-lived process (`cmd/scraper`) that syncs events on an interval, not a one-shot job. It's defined under the `full` Compose profile (`docker compose --profile full up`, or `make up-full`); local dev is expected to leave it off by default since it needs a real Telegram session most developers won't have.

In any environment where stale event data is unacceptable (staging, production), set in that environment's `.env`:

- `SCRAPE_ENABLED=true` and, if Telegram is in use, `TELEGRAM_ENABLED=true` — these seed the initial `scrape_settings` row on first boot only; after that, the admin settings UI is the source of truth.
- `SCRAPE_INTERVAL` sets the *default* interval used to seed that row. Once seeded, changing this env var has no further effect — the running scraper polls `scrape_settings.scrape_interval_seconds` from Postgres (checked roughly once a minute) and re-reads the full settings row before every sync, so admin changes to the interval, channels, or keywords take effect without a restart.
- The admin settings response includes `last_run_at`, `last_success_at`, and `last_error` — check these to confirm the scraper is actually running rather than inferring it from event freshness.

The Telegram session (`mq_telegram_session` Docker volume, used by both `api` and `scraper`) is populated by running `cmd/telegram-login` once per environment/session rotation — it is not created automatically. If that volume is ever recreated, scraping will fail until `telegram-login` is re-run; `last_error` on the scrape settings surfaces this. Never commit the resulting `.session` file — it's a live MTProto login credential for the Telegram account, equivalent to a password. `backend/api/data/*.session` is gitignored for this reason.

## Digest Telegram bot

Account linking and Telegram delivery for the weekly digest use a **separate Telegram credential** from the scraper above: a Bot API token (`TELEGRAM_BOT_TOKEN`, from [@BotFather](https://t.me/BotFather)), not the MTProto user session. Mixing these up doesn't work — a bot token can't read channels for scraping, and the scraper's user session can't be the target of a `t.me/<bot>?start=` deep link.

- `TELEGRAM_BOT_TOKEN` — if set, `cmd/api` starts a long-polling loop (`getUpdates`) in a background goroutine on startup; no public HTTPS endpoint or webhook is needed. If unset, the bot is simply off — `POST /api/v1/auth/me/telegram-link` returns 503.
- `TELEGRAM_BOT_USERNAME` — the bot's `@handle` (without `@`), used to build the `t.me/<username>?start=<token>` links returned by that endpoint.
- Linking is one-time-token based: the settings page requests a link, the token is stored in `telegram_link_tokens` with a 15-minute expiry, and the bot consumes it on `/start <token>`. `/stop` unlinks. Both are handled by `internal/usecase/digest.Service`, not the HTTP layer.
- The digest's own unsubscribe link (`GET /unsubscribe?token=...`) is unrelated to Telegram — it's a stateless, non-expiring HMAC token signed with `JWT_SECRET`, since an email's unsubscribe link has to keep working no matter how long the email sits unread.

## Weekly digest

The `digest` service (`cmd/digest`) is a long-lived process, like `scraper` — it computes the next scheduled send time and sleeps until then, rather than polling on a short interval. It's under the `full` Compose profile alongside `scraper`.

- `DIGEST_WEEKDAY` (default `monday`) and `DIGEST_HOUR` (default `9`, UTC, 0–23) set the fixed weekly send time. This is deliberately not a `SCRAPE_INTERVAL`-style duration: a plain ticker fires immediately on every process restart and drifts from a wall-clock target over time, both wrong for a "every Monday at 9am" expectation. Changing these env vars takes effect on the next process restart, computed fresh each cycle — there's no admin-UI override the way scrape settings have one.
- A send is idempotent per recipient per channel, tracked in `digest_runs`/`digest_deliveries`: if the process crashes mid-send, the next scheduled trigger resumes the same incomplete run (same run ID) instead of starting a new one, so recipients already delivered to aren't sent to twice. This only self-heals on the *next* scheduled trigger, not immediately on restart — an interruption shortly after send time means the affected recipients wait up to a week for the retry.
- Uses the same `RESEND_API_KEY`/`MAIL_FROM` as the rest of the API, and `TELEGRAM_BOT_TOKEN` from the section above for the Telegram side. Both are optional independently: with no bot token, Telegram-linked recipients are simply skipped; with no `RESEND_API_KEY`, email falls back to the log mailer (nothing actually sent — fine for local dev, not for production).
- `WEB_APP_URL` and `PUBLIC_API_URL` are used to build the links inside the digest (event/article/artist pages point at `WEB_APP_URL`; the unsubscribe link points at `PUBLIC_API_URL`, since that's served by the API, not the frontend).

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
- `/api/v1/auth/me/telegram-link` returns a `t.me` deep link, and sending `/start <token>` from that chat links it (check `user_notification_preferences.telegram_chat_id`); `/stop` unlinks it.
- `GET /unsubscribe?token=...` with a valid token turns off `newsletter_enabled`, and with a garbage token still returns the confirmation page rather than an error.
- `cmd/digest` logs its next scheduled send time on startup at the configured `DIGEST_WEEKDAY`/`DIGEST_HOUR`; a manual `SendWeekly` run (or waiting for the schedule) produces a completed row in `digest_runs` with matching `digest_deliveries`.
- Database backup completed and restore procedure was last tested within the required recovery window.
