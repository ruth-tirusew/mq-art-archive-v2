<script lang="ts">
	import { cn } from '$lib/utils';

	let { class: className }: { class?: string } = $props();
</script>

<div class={cn('rounded-sm border border-warning/40 bg-warning/10 p-6', className)} role="status" data-testid="moderation-hold">
	<p class="font-display text-xl text-foreground">An editor has taken your profile offline</p>
	<p class="mt-2 text-sm leading-relaxed text-muted-foreground">
		Visitors can't see your profile or your works right now, and you can't make it public again yourself. Nothing has been deleted: your
		profile, works and answers are all still here.
	</p>
	<p class="mt-3 text-sm leading-relaxed text-foreground">
		To have it restored, reply to any email you've had from the Artiv team, or contact your Artiv contact. They can tell you why it was taken
		down and what to change.
	</p>
</div>
