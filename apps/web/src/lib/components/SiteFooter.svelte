<script lang="ts">
	const year = new Date().getFullYear();
</script>

<footer class="overflow-hidden border-t border-border/60">
	<div class="mx-auto flex max-w-[1600px] flex-col gap-10 px-6 pt-12 md:px-10">
		<div class="flex flex-wrap items-baseline justify-between gap-4">
			<p class="text-sm text-muted-foreground">Modern Ethiopian art, held in common.</p>
			<p class="t-eyebrow text-muted-foreground">© {year} Artiv · Addis Ababa</p>
		</div>

		<!-- Signature wordmark: the logo at page width, closing every page. "artiv." in Fraunces is
		     ~1.62em wide, so these sizes fill the padded container at each breakpoint. -->
		<p
			class="font-display text-center text-[54vw] leading-[0.82] tracking-[-0.045em] text-foreground select-none md:text-[min(55vw,58rem)]"
			aria-hidden="true"
		>
			artiv.
		</p>
	</div>
</footer>
