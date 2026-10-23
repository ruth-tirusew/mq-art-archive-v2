# @artiv/ui — design system

Source-only Svelte 5 + Tailwind 4 kit shared by `apps/web` and `apps/admin`. No build or install step:
apps import it through the `$ui` alias. Live reference: run web and open `/design` (dev only).

## Principles (taken from the landing page)
- **Paper and ink.** Cream background, ink text. Oxblood (`accent`) is the one accent; ochre appears only on ink surfaces.
- **Three voices.** Fraunces for headings (emphasise with *italic*, never bold), Inter for reading, JetBrains Mono for labels and metadata.
- **Hairlines, not shadows.** 1px `border-border` dividers, `rounded-sm` surfaces, pill-shaped buttons. Shadows only on hover and overlays.
- **Every state is designed.** Loading → `Skeleton` / `NavProgress`. Empty → `EmptyState` with a next step. Error → `ErrorState` with a real retry. Success → `toast`. Public-facing or irreversible actions → `ConfirmDialog`.

## Use it
```css
/* app.css */
@import 'tailwindcss' source(none);
@source "../src";
@source "../../../packages/ui/src";
@import 'tw-animate-css';
@import '../../../packages/ui/src/styles/theme.css';
```
```js
// svelte.config.js kit.alias (see apps/web) — $ui plus the three bare-import aliases
// vite.config.ts: resolve.dedupe ['svelte'], server.fs.allow ['../..']
```
```svelte
<script>
	import { Button, Field, Input, toast } from '$ui';
</script>
```
Mount `<NavProgress />` and `<Toaster />` once in the root layout.

## Rules
- Type roles (`t-h1`, `t-eyebrow`, …) and `container-page` live in `theme.css`. Don't hand-write `font-mono text-[10px] tracking-[…]`.
- New component needed twice? Add it here and to `/design` first.
- Hosting: the app's Vercel project must allow source files outside its root directory.
