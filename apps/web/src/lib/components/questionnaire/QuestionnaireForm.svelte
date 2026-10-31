<script lang="ts">
	import { onDestroy, onMount } from 'svelte';
	import { Alert, Input } from '$ui';
	import RichTextEditor from '$lib/components/RichTextEditor.svelte';
	import { questionnaireService } from '$lib/application/questionnaire';
	import type { FormQuestion, QuestionnaireForm } from '$lib/core/domain/questionnaire';
	import { AnswerAutosave, type SaveStatus } from '$lib/utils/answerAutosave';
	import {
		answerLength,
		answerProblem,
		detectLang,
		formatFor,
		htmlText,
		isBlankAnswer,
		missingRequired
	} from '$lib/utils/answers';

	type Props = {
		form: QuestionnaireForm;
		/** Show answers but don't allow edits (also forced when the server says the form is locked). */
		readonly?: boolean;
		/**
		 * "apply" enforces required questions. "studio" is for approved artists: nothing is required
		 * any more (questions added later are always optional for them), so no required marks or errors.
		 */
		mode?: 'apply' | 'studio';
		/** Question keys to flag as new, e.g. those a nudge is about. */
		highlight?: string[];
	};

	let { form, readonly = false, mode = 'apply', highlight = [] }: Props = $props();
	const enforceRequired = $derived(mode === 'apply');

	const locked = $derived(readonly || form.locked);
	const questions = $derived([...form.revision.questions].sort((a, b) => a.position - b.position));

	// Editing state is seeded once from the saved answers; the form owns it from then on.
	// svelte-ignore state_referenced_locally
	let values = $state<Record<string, string>>(Object.fromEntries(form.answers.map((a) => [a.key, a.body])));
	let touched = $state<Record<string, boolean>>({});
	let showAll = $state(false);
	let status = $state<SaveStatus>('idle');
	let statusError = $state('');

	const questionByKey = $derived(new Map(questions.map((q) => [q.key, q])));

	const autosave = new AnswerAutosave({
		save: (answers, context) => questionnaireService.saveAnswers(answers, context),
		onStatus: (s, err) => {
			status = s;
			statusError = err ?? '';
		}
	});

	function problemFor(q: FormQuestion): string | null {
		return answerProblem(q, values[q.key] ?? '');
	}

	function errorFor(q: FormQuestion): string | undefined {
		const shown = showAll || touched[q.key];
		if (!shown) return undefined;
		const problem = problemFor(q);
		if (problem) return problem;
		if (enforceRequired && q.required && isBlankAnswer(values[q.key] ?? '', formatFor(q))) return 'This question is required.';
		return undefined;
	}

	function change(q: FormQuestion, body: string) {
		values[q.key] = body;
		touched[q.key] = true;
		// The server rejects a whole batch if one answer is invalid, so a bad answer stays local
		// (and flagged) rather than stopping everything else from saving.
		if (problemFor(q)) return;
		const format = formatFor(q);
		const blank = isBlankAnswer(body, format);
		autosave.set({
			key: q.key,
			body: blank ? '' : body,
			format,
			lang: q.type === 'long_text' || q.type === 'short_text' ? detectLang(format === 'rich' ? htmlText(body) : body) : 'en'
		});
	}

	/** Save everything now. True when nothing is left unsaved. */
	export function flush(): Promise<boolean> {
		return autosave.flush();
	}

	/** Show every problem and focus the first one. True when the form is ready to submit. */
	export function validate(): boolean {
		showAll = true;
		const missing = new Set(enforceRequired ? missingRequired(questions, values) : []);
		const firstBad = questions.find((q) => missing.has(q.key) || problemFor(q));
		if (firstBad) {
			queueMicrotask(() =>
				document.getElementById(`q-${firstBad.key}`)?.scrollIntoView({ behavior: 'smooth', block: 'center' })
			);
			return false;
		}
		return true;
	}

	/** Current answer text per question key, for the review step. */
	export function current(): Record<string, string> {
		return { ...values };
	}

	function warnUnsaved(event: BeforeUnloadEvent) {
		if (!autosave.hasUnsaved) return;
		// Try to save in the background first. Only interrupt with the browser's "leave site?" prompt
		// when saving has already been failing (e.g. offline), the case where leaving really loses work.
		void autosave.flush({ keepalive: true });
		if (status === 'error') event.preventDefault();
	}

	// Switching tabs/apps or closing the page within the quiet period must not lose the last edit,
	// so flush with keepalive when the page is hidden. beforeunload additionally warns if something
	// still couldn't be saved.
	function flushForExit() {
		if (autosave.hasUnsaved) void autosave.flush({ keepalive: true });
	}
	function onVisibility() {
		if (document.visibilityState === 'hidden') flushForExit();
	}

	onMount(() => {
		window.addEventListener('beforeunload', warnUnsaved);
		window.addEventListener('pagehide', flushForExit);
		document.addEventListener('visibilitychange', onVisibility);
		return () => {
			window.removeEventListener('beforeunload', warnUnsaved);
			window.removeEventListener('pagehide', flushForExit);
			document.removeEventListener('visibilitychange', onVisibility);
		};
	});
	// Leaving the page within the quiet period still saves: the request continues in the SPA.
	onDestroy(() => void autosave.flush());

	const statusText = $derived(
		status === 'saving' ? 'Saving…' : status === 'saved' ? 'All changes saved' : status === 'error' ? '' : ''
	);

	function near(q: FormQuestion): number | null {
		const used = answerLength(values[q.key] ?? '', formatFor(q));
		return used >= q.max_length * 0.8 ? used : null;
	}
</script>

{#snippet meta(q: FormQuestion)}
	<p class="font-mono text-[10px] uppercase tracking-[0.2em] {q.is_public ? 'text-muted-foreground' : 'text-accent'}">
		{q.is_public ? 'Shown on your public profile' : 'Only our team sees this'}
	</p>
{/snippet}

{#snippet newBadge(q: FormQuestion)}
	{#if highlight.includes(q.key)}
		<span class="ml-2 rounded-full border border-accent/50 px-2 py-0.5 align-middle font-mono text-[9px] tracking-[0.2em] text-accent">NEW</span>
	{/if}
{/snippet}

{#snippet heading(q: FormQuestion, forId: string | null)}
	<div class="space-y-1">
		{#if forId}
			<label for={forId} class="t-eyebrow block text-foreground">
				{q.prompt_en}{#if q.required && enforceRequired}<span class="text-accent" aria-hidden="true"> *</span>{/if}{@render newBadge(q)}
			</label>
		{:else}
			<p class="t-eyebrow text-foreground">
				{q.prompt_en}{#if q.required && enforceRequired}<span class="text-accent" aria-hidden="true"> *</span>{/if}{@render newBadge(q)}
			</p>
		{/if}
		{#if q.prompt_am}<p class="text-sm text-muted-foreground" lang="am">{q.prompt_am}</p>{/if}
		{#if q.help_en}<p class="text-[13px] text-muted-foreground">{q.help_en}</p>{/if}
		{#if q.help_am}<p class="text-[13px] text-muted-foreground" lang="am">{q.help_am}</p>{/if}
	</div>
{/snippet}

{#snippet footer(q: FormQuestion, error: string | undefined)}
	<div class="flex flex-wrap items-baseline justify-between gap-x-4 gap-y-1">
		<div>
			{#if error}<p id="q-{q.key}-error" role="alert" class="text-[13px] text-destructive">{error}</p>{/if}
		</div>
		<div class="flex items-baseline gap-4">
			{#if near(q) !== null}
				<span class="font-mono text-[10px] tracking-wider {near(q)! > q.max_length ? 'text-destructive' : 'text-muted-foreground'}">
					{near(q)!.toLocaleString('en')} / {q.max_length.toLocaleString('en')}
				</span>
			{/if}
			{@render meta(q)}
		</div>
	</div>
{/snippet}

<div class="space-y-10" data-testid="questionnaire-form">
	{#if locked}
		<Alert tone="info">
			{form.locked
				? 'Your application is being reviewed, so your answers are read-only for now. If we ask for changes you’ll be able to edit them again.'
				: 'These answers are read-only.'}
		</Alert>
	{/if}

	{#each questions as q (q.key)}
		{@const error = errorFor(q)}
		<div id="q-{q.key}" class="scroll-mt-24 space-y-3" data-question={q.key}>
			{#if q.type === 'short_text' || q.type === 'link'}
				{@render heading(q, `f-${q.key}`)}
				<Input
					id="f-{q.key}"
					type={q.type === 'link' ? 'url' : 'text'}
					placeholder={q.type === 'link' ? 'https://' : undefined}
					autocomplete="off"
					invalid={!!error}
					aria-describedby={error ? `q-${q.key}-error` : undefined}
					disabled={locked}
					lang={detectLang(values[q.key] ?? '') === 'am' ? 'am' : undefined}
					value={values[q.key] ?? ''}
					oninput={(e) => change(q, e.currentTarget.value)}
					onblur={() => (touched[q.key] = true)}
				/>
			{:else if q.type === 'long_text'}
				{@render heading(q, null)}
				<RichTextEditor
					value={values[q.key] ?? ''}
					label={q.prompt_en}
					invalid={!!error}
					describedby={error ? `q-${q.key}-error` : undefined}
					disabled={locked}
					onchange={(html) => change(q, html)}
				/>
			{:else if q.type === 'choice'}
				<fieldset class="space-y-3" aria-describedby={error ? `q-${q.key}-error` : undefined}>
					<legend class="contents">{@render heading(q, null)}</legend>
					<div class="space-y-2">
						{#each q.options ?? [] as option (option.value)}
							<label class="flex cursor-pointer items-start gap-3 text-sm {locked ? 'cursor-not-allowed opacity-60' : ''}">
								<input
									type="radio"
									name="q-{q.key}"
									class="mt-1 size-4 accent-[var(--accent)]"
									value={option.value}
									checked={values[q.key] === option.value}
									disabled={locked}
									onchange={() => change(q, option.value)}
								/>
								<span>
									{option.label_en}
									{#if option.label_am}<span class="ml-2 text-muted-foreground" lang="am">{option.label_am}</span>{/if}
								</span>
							</label>
						{/each}
					</div>
				</fieldset>
			{/if}
			{@render footer(q, error)}
		</div>
	{/each}

	<div class="flex min-h-6 items-center gap-3 text-sm" aria-live="polite">
		{#if status === 'error'}
			<p class="text-destructive" role="alert">
				Couldn’t save your latest changes{statusError ? `: ${statusError}` : ''}.
				<button type="button" class="ml-1 underline underline-offset-4" onclick={() => void autosave.flush()}>Try again</button>
			</p>
		{:else if statusText}
			<p class="text-muted-foreground">{statusText}</p>
		{:else if !locked}
			<p class="text-muted-foreground">Your answers save automatically as you type.</p>
		{/if}
	</div>
</div>
