<script lang="ts">
	import { TELEGRAM_DIGEST_URL } from '$lib/config/site';
	import type { Event } from '$lib/core/domain/events';
	import { weekStats } from './eventFormat';

	let { events }: { events: Event[] } = $props();

	const stats = $derived(weekStats(events));
</script>

<section class="border-b border-border/70">
	<div class="mx-auto grid max-w-[1600px] gap-10 px-6 py-14 md:grid-cols-12 md:px-10 md:py-16">
		<div class="md:col-span-7 lg:col-span-8">
			<p class="font-mono text-[11px] uppercase tracking-[0.3em] text-accent">
				Pillar 03 · The discovery hub
			</p>
			<h1 class="mt-4 max-w-3xl font-display text-4xl leading-[1.05] text-foreground md:text-6xl">
				Every opening, reading and pop-up — <em class="italic">in one place</em>.
			</h1>
			<p class="mt-6 max-w-xl text-base leading-relaxed text-muted-foreground md:text-lg">
				We scan Telegram channels, Facebook pages and gallery sites every morning. De-duplicated,
				tagged and organized so you can discover what's happening across Addis and beyond.
			</p>
			<p class="mt-4 font-mono text-[10px] uppercase tracking-[0.22em] text-muted-foreground">
				Auto-synced this morning · 06:02 EAT
			</p>

			<div
				class="mt-10 grid grid-cols-2 gap-x-6 gap-y-5 border-t border-border/70 pt-6 sm:grid-cols-5"
			>
				<div>
					<p class="font-display text-3xl text-accent md:text-4xl">{stats.total}</p>
					<p class="mt-1 font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground">
						Events this week
					</p>
				</div>
				<div>
					<p class="font-display text-3xl text-accent md:text-4xl">{stats.openings}</p>
					<p class="mt-1 font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground">
						Openings
					</p>
				</div>
				<div>
					<p class="font-display text-3xl text-accent md:text-4xl">{stats.workshops}</p>
					<p class="mt-1 font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground">
						Workshops
					</p>
				</div>
				<div>
					<p class="font-display text-3xl text-accent md:text-4xl">{stats.exhibitions}</p>
					<p class="mt-1 font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground">
						Exhibitions
					</p>
				</div>
				<div>
					<p class="font-display text-3xl text-accent md:text-4xl">{stats.talks}</p>
					<p class="mt-1 font-mono text-[9px] uppercase tracking-[0.2em] text-muted-foreground">
						Talks & readings
					</p>
				</div>
			</div>
		</div>

		<div class="md:col-span-5 lg:col-span-4">
			<div class="rounded-sm border border-border bg-card/80 p-7 shadow-sm">
				<div
					class="flex h-11 w-11 items-center justify-center rounded-full bg-foreground text-background"
					aria-hidden="true"
				>
					<svg class="h-5 w-5" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6">
						<path d="M22 2 11 13" stroke-linecap="round" />
						<path d="M22 2 15 22 11 13 2 9l20-7Z" stroke-linejoin="round" />
					</svg>
				</div>
				<h2 class="mt-5 font-display text-2xl leading-tight text-foreground md:text-3xl">
					The weekly digest, in your Telegram.
				</h2>
				<p class="mt-3 text-sm leading-relaxed text-muted-foreground">
					Every Sunday we send a curated rundown of the week's openings, talks and pop-ups —
					straight to where the community already is.
				</p>
				<div class="mt-6 flex flex-col gap-3 sm:flex-row sm:items-center">
					<input
						placeholder="@your-telegram"
						class="flex-1 rounded-full border border-border bg-background px-4 py-2.5 font-mono text-[11px] text-foreground placeholder:text-muted-foreground/60 focus:border-foreground focus:outline-none"
						readonly
						aria-label="Telegram username"
					/>
					<a
						href={TELEGRAM_DIGEST_URL}
						target="_blank"
						rel="noopener noreferrer"
						class="rounded-full bg-foreground px-4 py-2.5 text-center font-mono text-[10px] uppercase tracking-[0.2em] text-background transition hover:bg-accent"
					>
						Subscribe →
					</a>
				</div>
				<p class="mt-4 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground/70">
					1,247 subscribers · last digest sent Sun, Jun 21
				</p>
			</div>
		</div>
	</div>
</section>
