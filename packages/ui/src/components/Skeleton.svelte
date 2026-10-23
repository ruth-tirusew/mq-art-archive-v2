<script lang="ts">
	import { cn } from '../cn';

	let { class: className }: { class?: string } = $props();
</script>

<div class={cn('animate-pulse rounded-sm bg-muted', className)} aria-hidden="true"></div>
