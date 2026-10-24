<script lang="ts">
	let { class: className = '' }: { class?: string } = $props();
</script>

<div class="animate-pulse rounded-sm bg-border/50 {className}" aria-hidden="true"></div>
