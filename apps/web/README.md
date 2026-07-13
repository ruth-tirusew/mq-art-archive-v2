# Mäkdäs — SvelteKit

A SvelteKit port of the TanStack Start art archive prototype (`mq-art-archive-main`).

## Stack

- **SvelteKit 2** + **Svelte 5**
- **Tailwind CSS v4** (same design tokens as the React app)
- Static TypeScript data (no API required)

## Routes

| Path | Description |
|------|-------------|
| `/` | Homepage with featured artist, acquisitions, wiki/events CTAs |
| `/archive` | Filterable works catalog |
| `/artists` | Artist roster |
| `/artists/[slug]` | Artist profile + scroll timeline |
| `/events` | Events calendar |
| `/events/[slug]` | Event detail |
| `/wiki` | Community handbook index |
| `/wiki/[slug]` | Wiki article |
| `/portfolio` | Portfolio builder marketing |
| `/@[handle]` | Shareable link-in-bio profile |
| `/about` | About page |

## Development

```bash
npm install
npm run dev
```

Runs on **http://localhost:8082** by default.

## Relation to other projects

- **TanStack prototype**: `../mq-art-archive-main` (retired)
- **Production SvelteKit + API**: `../../mq/apps/web` (live API integration)

This project is a standalone static-data conversion for reference and experimentation.
