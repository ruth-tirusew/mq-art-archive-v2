# Artiv training video generator

Records a captioned training video for **artists** and for **the people who invite them**, by driving the real
apps in a browser. It makes the video, subtitles and a timed narration script. Nothing is saved or sent while
it runs.

```bash
# the apps must be running (API :8080, web :8082, admin :5174, see the main README)
node scripts/demo-video/record.mjs --list                # plan + runtime, records nothing
node scripts/demo-video/record.mjs                       # full cut
node scripts/demo-video/record.mjs --cut=artist          # for artists
node scripts/demo-video/record.mjs --cut=bd              # for the team that invites artists
node scripts/demo-video/record.mjs --chapters=studio,page --shots   # a few chapters + a picture per beat
```

No install step: it reuses Playwright from `apps/web`. Output goes to `scripts/demo-video/out/` (git-ignored):

| File | What it is |
|---|---|
| `demo-<cut>.webm` | the video: chapter cards, narration captions, a visible cursor, highlights |
| `captions-<cut>.vtt` | the same narration as subtitles (load it in any player or upload with the video) |
| `narration-<cut>.md` | the script with timestamps, to read as a voiceover |
| `chapters-<cut>.json` | when each chapter starts (for YouTube/Drive chapters) |

## Demo logins

Chapters that sign in need an account. Without one they are skipped and the rest still records.

| Variable | For |
|---|---|
| `DEMO_ARTIST_EMAIL`, `DEMO_ARTIST_PASSWORD` | the studio and account chapters |
| `DEMO_ARTIST_HANDLE` | whose public page to show (default `gebrekristosdesta`) |
| `DEMO_ADMIN_EMAIL`, `DEMO_ADMIN_PASSWORD` | the live admin Invites screens in the team chapter |
| `WEB_URL`, `ADMIN_URL` | where the apps are (defaults `http://localhost:8082`, `http://localhost:5174`) |

```bash
DEMO_ARTIST_EMAIL=you@example.com DEMO_ARTIST_PASSWORD='...' node scripts/demo-video/record.mjs
```

Pick an artist with a **short bio, a portrait and a few works that have a year, medium and palette**, so the
screens look good. For the best result record against a staging copy, not a laptop database.

## Changing what it says

Everything spoken and clicked is in `chapters.mjs`, one beat per step: the narration (`say`), the actions
(`do`), and the minimum time (`hold`). A caption stays up as long as it takes to read (about 2.5 words a second),
so longer sentences give a longer beat. The recorder never needs to change for a wording edit.

## What is shown live and what is a slide

It records the real app wherever that is harmless: the homepage, archive, wiki, studio, public page, share link and
account page. Typing into the bio is shown and then abandoned, and it never presses Save.

These are shown as explainer slides instead, because running them for real would email someone, upload files to
Cloudinary or create accounts: **opening an invitation, the four apply steps, and the decision emails.** The admin
invite screen is live (with the admin login) but stops before pressing Send.

## Voiceover

The video has captions but no sound, because this tool has no speech engine. Either:

- record a voiceover while reading `narration-<cut>.md` (the timestamps match the video), or
- feed the `.vtt` or the narration text to a text-to-speech service, and mix the audio over the video in your
  editor. Each caption lasts about as long as its sentence takes to read aloud, so the pacing already fits.

For an **Amharic** version: translate `say` lines in `chapters.mjs` (keep sentences short), record it as another
cut, and the subtitles will follow. The cards need an Ethiopic font installed (for example Noto Sans Ethiopic).

## Notes

- The video is WebM (what the browser records). Convert for WhatsApp or Telegram with
  `ffmpeg -i out/demo-full.webm -c:v libx264 -pix_fmt yuv420p out/demo-full.mp4`.
- The cards use the fonts on this machine. Install Fraunces and Inter to match the product.
- Both apps share the browser's `localhost` cookies, so the script signs in again between the team chapter and the
  artist chapters. That is normal.
