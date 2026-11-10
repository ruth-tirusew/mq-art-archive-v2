<script lang="ts">
  import { Badge } from '$lib/components/ui/badge';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import SafeHtml from '$lib/components/SafeHtml.svelte';
  import type { Answer, Question, ReviewView } from '$lib/core/domain/questionnaire';

  let { view }: { view: ReviewView } = $props();

  /** A choice answer is stored as the option's value; show the label the artist actually saw. */
  function choiceLabel(question: Question, value: string): string {
    return question.options?.find((o) => o.value === value)?.label_en ?? value;
  }

  const langName = { en: 'English', am: 'Amharic', mixed: 'Mixed' } as const;
</script>

{#snippet answerBody(question: Question | null, answer: Answer)}
  {#if answer.format === 'rich'}
    <SafeHtml html={answer.body} lang={answer.lang === 'am' ? 'am' : undefined} />
  {:else}
    <p class="whitespace-pre-wrap text-sm text-foreground" lang={answer.lang === 'am' ? 'am' : undefined}>
      {question?.type === 'choice' ? choiceLabel(question, answer.body) : answer.body}
    </p>
  {/if}
{/snippet}

<div class="mt-6" data-testid="application-answers">
<Card>
  <CardHeader>
    <CardTitle>Questionnaire answers</CardTitle>
    <p class="text-sm text-muted-foreground">
      Answered on revision {view.revision.number ?? '—'}. Answers marked “Admins only” never appear on the public profile.
    </p>
  </CardHeader>
  <CardContent class="space-y-6">
    {#each view.items as item (item.question.key)}
      <div class="space-y-2 border-b border-border pb-5 last:border-b-0 last:pb-0">
        <div class="flex flex-wrap items-center gap-2">
          <h3 class="text-sm font-medium text-foreground">{item.question.prompt_en}</h3>
          {#if item.question.required}<Badge variant="secondary">Required</Badge>{/if}
          {#if !item.question.is_public}<Badge variant="warning">Admins only</Badge>{/if}
          {#if item.answer && item.answer.lang !== 'en'}<Badge variant="muted">{langName[item.answer.lang]}</Badge>{/if}
        </div>
        {#if item.question.prompt_am}
          <p class="text-sm text-muted-foreground" lang="am">{item.question.prompt_am}</p>
        {/if}
        {#if item.answer && item.answer.body.trim()}
          {@render answerBody(item.question, item.answer)}
        {:else}
          <p class="text-sm italic text-muted-foreground">Not answered</p>
        {/if}
      </div>
    {/each}

    {#if view.other.length > 0}
      <div class="space-y-3 rounded-md border border-dashed border-border p-4">
        <p class="font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground">
          Answers to questions outside this revision
        </p>
        {#each view.other as answer (answer.key)}
          <div class="space-y-1">
            <p class="font-mono text-xs text-muted-foreground">{answer.key}</p>
            {@render answerBody(null, answer)}
          </div>
        {/each}
      </div>
    {/if}
  </CardContent>
</Card>
</div>
