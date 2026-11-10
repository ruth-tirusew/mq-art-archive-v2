<script lang="ts">
	import { onMount } from 'svelte';
	import { questionnaireService } from '$lib/composition';
	import type { QuestionnaireNudge } from '$lib/core/domain/questionnaire';

	let nudge = $state<QuestionnaireNudge | null>(null);

	onMount(async () => {
		nudge = await questionnaireService.getNudge();
	});

	async function dismiss() {
		if (!nudge) return;
		const revisionId = nudge.revision_id;
		nudge = null; // disappears immediately; a failed request just means it shows again next time
		try {
			await questionnaireService.dismissNudge(revisionId);
		} catch {
			// not worth interrupting anyone over
		}
	}
</script>

{#if nudge}
	<div class="border-b border-border/60 bg-accent/5" data-testid="nudge-banner">
		<div class="mx-auto flex max-w-[1600px] flex-wrap items-center gap-x-6 gap-y-2 px-6 py-3 md:px-10">
			<p class="text-sm text-foreground">
				We added {nudge.questions.length === 1 ? 'a new question' : `${nudge.questions.length} new questions`} to artist
				profiles. Answer {nudge.questions.length === 1 ? 'it' : 'them'} whenever you like; nothing changes if you don't.
			</p>
			<div class="flex items-center gap-4 font-mono text-[11px] uppercase tracking-[0.2em]">
				<a href="/studio/answers" class="text-accent underline underline-offset-4">Answer</a>
				<button type="button" class="text-muted-foreground hover:text-foreground" onclick={dismiss}>Not now</button>
			</div>
		</div>
	</div>
{/if}
