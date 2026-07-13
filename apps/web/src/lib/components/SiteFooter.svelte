<script lang="ts">
	const footerLinks = [
		{ href: '/archive', label: 'Archive' },
		{ href: '/artists', label: 'Artists' },
		{ href: '/wiki', label: 'Wiki' },
		{ href: '/events', label: 'Events' },
		{ href: '/portfolio', label: 'Portfolio' },
		{ href: '/about', label: 'About' }
	];

	const year = new Date().getFullYear();
</script>

<footer class="border-t border-border/60">
	<div class="mx-auto grid max-w-[1600px] gap-8 px-6 py-12 md:grid-cols-3 md:px-10">
		<div>
			<p class="font-display text-3xl text-foreground">mäkdäs</p>
			<p class="mt-2 max-w-xs text-sm text-muted-foreground">
				Modern Ethiopian art, held in common.
			</p>
		</div>
		<div class="font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground">
			<p class="mb-3 text-foreground">Index</p>
			<ul class="space-y-2">
				{#each footerLinks as l}
					<li>
						<a href={l.href} class="transition hover:text-foreground">{l.label}</a>
					</li>
				{/each}
			</ul>
		</div>
		<div class="font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground md:text-right">
			<p>© {year} Mäkdäs Archive</p>
		</div>
	</div>
</footer>
