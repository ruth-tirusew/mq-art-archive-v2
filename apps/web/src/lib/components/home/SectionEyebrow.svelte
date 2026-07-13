<script lang="ts">
	type Props = { number: string; label: string };
	let { number, label }: Props = $props();
</script>

<p class="flex items-center gap-3 font-mono text-[11px] uppercase tracking-[0.3em] text-muted-foreground">
	<span class="text-accent">{number}</span>
	<span class="text-muted-foreground/50">—</span>
	<span>{label}</span>
</p>
