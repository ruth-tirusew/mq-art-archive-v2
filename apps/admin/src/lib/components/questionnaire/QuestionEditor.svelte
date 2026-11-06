<script lang="ts">
  import { Checkbox, Field, Input, Select } from '$ui';
  import ArrowDown from 'lucide-svelte/icons/arrow-down';
  import ArrowUp from 'lucide-svelte/icons/arrow-up';
  import Plus from 'lucide-svelte/icons/plus';
  import Trash2 from 'lucide-svelte/icons/trash-2';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import AutoTextarea from '$lib/components/AutoTextarea.svelte';
  import { Card, CardContent } from '$lib/components/ui/card';
  import { questionTypeLabels, type QuestionType } from '$lib/core/domain/questionnaire';
  import { suggestKey, type EditableQuestion } from '$lib/utils/questionnaireDraft';

  let {
    question = $bindable(),
    index,
    total,
    problems = [],
    otherKeys,
    onmoveup,
    onmovedown,
    onremove
  }: {
    question: EditableQuestion;
    index: number;
    total: number;
    problems?: string[];
    /** Keys used by the other questions, so a suggested key never collides. */
    otherKeys: () => string[];
    onmoveup: () => void;
    onmovedown: () => void;
    onremove: () => void;
  } = $props();

  // A key follows the English prompt until someone edits it by hand. Published keys never move.
  let keyTouched = $state(question.keyLocked || question.prompt_en.trim() !== '');

  function onPromptInput() {
    if (!keyTouched && !question.keyLocked) question.key = suggestKey(question.prompt_en, otherKeys());
  }

  function addOption() {
    question.options = [...question.options, { value: '', label_en: '', label_am: '' }];
  }

  function removeOption(i: number) {
    question.options = question.options.filter((_, n) => n !== i);
  }

  const types = Object.entries(questionTypeLabels) as [QuestionType, string][];
  const invalid = $derived(problems.length > 0);
  // Problems are only passed in once the admin has tried to save or publish. A blank title is the one
  // problem that belongs to a specific field, so mark that field (red border, aria-invalid) as well as
  // listing the message at the bottom of the card.
  const titleInvalid = $derived(invalid && !question.prompt_en.trim());
</script>

<Card class={invalid ? 'border-destructive/60' : undefined}>
  <CardContent class="space-y-5 pt-6">
    <div class="flex flex-wrap items-center gap-2">
      <span class="font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground">
        Question {index + 1}
      </span>
      {#if question.required}<Badge variant="secondary">Required</Badge>{/if}
      <Badge variant={question.is_public ? 'success' : 'warning'}>
        {question.is_public ? 'Shown publicly' : 'Admins only'}
      </Badge>
      <div class="ml-auto flex items-center gap-1">
        <Button variant="ghost" size="icon" disabled={index === 0} onclick={onmoveup}>
          <ArrowUp class="h-4 w-4" aria-hidden="true" /><span class="sr-only">Move question {index + 1} up</span>
        </Button>
        <Button variant="ghost" size="icon" disabled={index === total - 1} onclick={onmovedown}>
          <ArrowDown class="h-4 w-4" aria-hidden="true" /><span class="sr-only">Move question {index + 1} down</span>
        </Button>
        <Button variant="ghost" size="icon" onclick={onremove}>
          <Trash2 class="h-4 w-4" aria-hidden="true" /><span class="sr-only">Remove question {index + 1}</span>
        </Button>
      </div>
    </div>

    <div class="grid gap-4 md:grid-cols-2">
      <Field label="Question (English)" required>
        {#snippet children({ id, describedby, invalid: isInvalid })}
          <AutoTextarea {id} rows={1} singleLine aria-describedby={describedby} invalid={isInvalid || titleInvalid} bind:value={question.prompt_en} oninput={onPromptInput} maxlength={500} />
        {/snippet}
      </Field>
      <Field label="Question (Amharic)" hint="Optional. Shown to artists next to the English text.">
        {#snippet children({ id, describedby })}
          <AutoTextarea {id} rows={1} singleLine lang="am" aria-describedby={describedby} bind:value={question.prompt_am} maxlength={500} />
        {/snippet}
      </Field>
      <Field label="Help text (English)" hint="Optional. Guidance shown under the question.">
        {#snippet children({ id, describedby })}
          <AutoTextarea {id} rows={2} aria-describedby={describedby} bind:value={question.help_en} maxlength={1000} />
        {/snippet}
      </Field>
      <Field label="Help text (Amharic)">
        {#snippet children({ id, describedby })}
          <AutoTextarea {id} rows={2} lang="am" aria-describedby={describedby} bind:value={question.help_am} maxlength={1000} />
        {/snippet}
      </Field>
    </div>

    <div class="grid gap-4 md:grid-cols-3">
      <Field label="Answer type">
        {#snippet children({ id })}
          <Select {id} bind:value={question.type}>
            {#each types as [value, label]}
              <option {value}>{label}</option>
            {/each}
          </Select>
        {/snippet}
      </Field>
      <Field label="Length limit" hint="Characters. Blank uses the default for this type.">
        {#snippet children({ id, describedby })}
          <Input {id} type="number" min="1" max="20000" aria-describedby={describedby} bind:value={question.max_length} />
        {/snippet}
      </Field>
      <Field
        label="Key"
        hint={question.keyLocked
          ? 'Fixed: artists have already answered under this key.'
          : 'Identifies the question across revisions. Changing it later starts a fresh question.'}
      >
        {#snippet children({ id, describedby })}
          <Input
            {id}
            class="font-mono text-xs"
            aria-describedby={describedby}
            readonly={question.keyLocked}
            bind:value={question.key}
            oninput={() => (keyTouched = true)}
            maxlength={64}
          />
        {/snippet}
      </Field>
    </div>

    <div class="flex flex-wrap gap-x-8 gap-y-3">
      <Checkbox bind:checked={question.required}>Required to apply</Checkbox>
      <Checkbox bind:checked={question.is_public}>Show the answer on the artist's public profile</Checkbox>
    </div>

    {#if question.type === 'choice'}
      <div class="space-y-3 rounded-md border border-border p-4">
        <p class="font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground">Options</p>
        {#each question.options as option, i}
          <div class="grid gap-3 sm:grid-cols-[1fr_1fr_1fr_auto] sm:items-end">
            <Field label="Value" hint={i === 0 ? 'Stored with the answer.' : undefined}>
              {#snippet children({ id })}
                <Input {id} class="font-mono text-xs" bind:value={option.value} maxlength={100} />
              {/snippet}
            </Field>
            <Field label="Label (English)">
              {#snippet children({ id })}
                <Input {id} bind:value={option.label_en} maxlength={200} />
              {/snippet}
            </Field>
            <Field label="Label (Amharic)">
              {#snippet children({ id })}
                <Input {id} lang="am" bind:value={option.label_am} maxlength={200} />
              {/snippet}
            </Field>
            <Button variant="ghost" size="icon" onclick={() => removeOption(i)}>
              <Trash2 class="h-4 w-4" aria-hidden="true" /><span class="sr-only">Remove option {i + 1}</span>
            </Button>
          </div>
        {/each}
        <Button variant="outline" size="sm" onclick={addOption}>
          <Plus class="h-4 w-4" aria-hidden="true" /> Add option
        </Button>
      </div>
    {/if}

    {#if invalid}
      <ul class="space-y-1 text-[13px] text-destructive" role="alert">
        {#each problems as problem}<li>{problem}</li>{/each}
      </ul>
    {/if}
  </CardContent>
</Card>
