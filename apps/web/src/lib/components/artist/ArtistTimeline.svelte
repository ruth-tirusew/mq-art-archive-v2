<script lang="ts">
	import type { TimelineEntry, Work } from '$lib/data/archive';

	type Entry = TimelineEntry & { work?: Work };

	type Props = {
		entries: Entry[];
		artistName: string;
	};

	let { entries, artistName }: Props = $props();

	const KIND_LABEL: Record<TimelineEntry['kind'], string> = {
		birth: 'Born',
		study: 'Study',
		exhibition: 'Exhibition',
		residency: 'Residency',
		work: 'Work',
		note: 'Note'
	};

	let activeIdx = $state(0);
	let progress = $state(0);
	let containerEl: HTMLElement | undefined = $state();
	let itemEls: (HTMLLIElement | null)[] = $state([]);

	const active = $derived(entries[activeIdx]);
	const firstYear = $derived(entries[0]?.year);
	const lastYear = $derived(entries[entries.length - 1]?.year);

	$effect(() => {
		const observer = new IntersectionObserver(
			(obs) => {
				obs.forEach((o) => {
					if (o.isIntersecting) {
						const idx = itemEls.findIndex((n) => n === o.target);
						if (idx >= 0) activeIdx = idx;
					}
				});
			},
			{ rootMargin: '-45% 0px -45% 0px', threshold: 0 }
		);
		itemEls.forEach((n) => n && observer.observe(n));
		return () => observer.disconnect();
	});

	$effect(() => {
		const onScroll = () => {
			const el = containerEl;
			if (!el) return;
			const rect = el.getBoundingClientRect();
			const vh = window.innerHeight;
			const total = rect.height - vh;
			const scrolled = Math.min(Math.max(-rect.top, 0), total);
			progress = total > 0 ? scrolled / total : 0;
		};
		onScroll();
		window.addEventListener('scroll', onScroll, { passive: true });
		return () => window.removeEventListener('scroll', onScroll);
	});
</script>

<section bind:this={containerEl} class="relative border-b border-border/60">
	<div class="pointer-events-none sticky top-[64px] z-10 border-b border-border/50 bg-background/85 backdrop-blur">
		<div
			class="mx-auto flex max-w-[1600px] items-center justify-between gap-6 px-6 py-3 font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground md:px-10"
		>
			<span class="text-foreground">{artistName} · Timeline</span>
			<div class="flex flex-1 items-center gap-3">
				<span>{firstYear}</span>
				<div class="relative h-px flex-1 bg-border">
					<div
						class="absolute inset-y-0 left-0 bg-accent transition-[width] duration-300"
						style:width="{progress * 100}%"
					></div>
					<div
						class="absolute -top-1 h-2 w-2 rounded-full bg-accent transition-[left] duration-300"
						style:left="calc({progress * 100}% - 4px)"
					></div>
				</div>
				<span>{lastYear}</span>
			</div>
			<span class="hidden text-accent md:inline">
				{active?.year} · {KIND_LABEL[active?.kind ?? 'note']}
			</span>
		</div>
	</div>

	<div class="mx-auto max-w-[1600px] px-6 py-16 md:px-10 md:py-24">
		<div class="grid gap-8 md:grid-cols-12">
			<aside class="hidden md:col-span-2 md:block">
				<div class="sticky top-32 space-y-1 font-mono text-[10px] uppercase tracking-[0.25em]">
					{#each entries as e, i}
						<button
							type="button"
							onclick={() => itemEls[i]?.scrollIntoView({ behavior: 'smooth', block: 'center' })}
							class="flex w-full items-center gap-3 py-1 text-left transition {i === activeIdx
								? 'text-foreground'
								: 'text-muted-foreground hover:text-foreground'}"
						>
							<span
								class="h-px transition-all {i === activeIdx ? 'w-8 bg-accent' : 'w-3 bg-border'}"
							></span>
							<span>{e.year}</span>
						</button>
					{/each}
				</div>
			</aside>

			<div class="md:col-span-10">
				<p class="font-mono text-[10px] uppercase tracking-[0.3em] text-accent">A life, chronologically</p>
				<h2 class="mt-3 font-display text-4xl text-foreground md:text-6xl">
					{firstYear}<span class="text-muted-foreground"> — </span>{lastYear}
				</h2>

				<ol class="mt-16 space-y-24 md:space-y-32">
					{#each entries as e, i}
						<li
							bind:this={itemEls[i]}
							class="scroll-mt-40"
						>
							<div
								class="grid grid-cols-12 gap-6 transition-opacity duration-500 {i === activeIdx
									? 'opacity-100'
									: 'opacity-55'}"
							>
								<div class="col-span-3 md:col-span-2">
									<p class="font-display text-4xl leading-none text-foreground md:text-6xl">
										{String(e.year).slice(-2)}
									</p>
									<p class="mt-1 font-mono text-[9px] uppercase tracking-[0.25em] text-muted-foreground">
										{String(e.year).slice(0, 2)}′
									</p>
								</div>
								<div class="col-span-9 md:col-span-6">
									<p
										class="inline-flex items-center gap-2 rounded-full bg-accent/10 px-3 py-1 font-mono text-[9px] uppercase tracking-[0.25em] text-accent"
									>
										<span class="h-1 w-1 rounded-full bg-accent"></span>
										{KIND_LABEL[e.kind]}
									</p>
									<h3 class="mt-3 font-display text-2xl leading-tight text-foreground md:text-4xl">
										{e.title}
									</h3>
									{#if e.place}
										<p class="mt-2 font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
											{e.place}
										</p>
									{/if}
									{#if e.detail}
										<p class="mt-4 max-w-md text-[15px] leading-relaxed text-foreground/80">
											{e.detail}
										</p>
									{/if}
								</div>
								<div class="col-span-12 md:col-span-4">
									{#if e.work}
										<figure class="group">
											<div class="grain aspect-[4/5] overflow-hidden rounded-sm bg-card">
												<img
													src={e.work.image}
													alt={e.work.title}
													loading="lazy"
													class="h-full w-full object-cover transition duration-700 group-hover:scale-[1.03] {i ===
													activeIdx
														? 'scale-100 grayscale-0'
														: 'scale-[1.01] grayscale'}"
												/>
											</div>
											<figcaption class="mt-3 flex items-start justify-between gap-3">
												<p class="font-display text-sm italic text-foreground">{e.work.title}</p>
												<div class="flex gap-1 pt-1">
													{#each e.work.palette.slice(0, 4) as c}
														<span class="h-4 w-1.5" style:background={c}></span>
													{/each}
												</div>
											</figcaption>
										</figure>
									{:else}
										<div class="flex h-full items-end">
											<span class="font-mono text-[10px] uppercase tracking-[0.25em] text-muted-foreground/60">
												— no work archived for this entry —
											</span>
										</div>
									{/if}
								</div>
							</div>
						</li>
					{/each}
				</ol>
			</div>
		</div>
	</div>
</section>
