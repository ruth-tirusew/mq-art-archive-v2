<script lang="ts">
  import { Alert as FormAlert, ConfirmDialog, createConfirm, toast } from '$ui';
  import { beforeNavigate, invalidateAll } from '$app/navigation';
  import Plus from 'lucide-svelte/icons/plus';
  import AutoTextarea from '$lib/components/AutoTextarea.svelte';
  import EmptyState from '$lib/components/EmptyState.svelte';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import QuestionEditor from '$lib/components/questionnaire/QuestionEditor.svelte';
  import { Alert } from '$lib/components/ui/alert';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Label } from '$lib/components/ui/label';
  import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '$lib/components/ui/table';
  import { questionnaireService } from '$lib/composition';
  import type { ChangeKind } from '$lib/core/domain/questionnaire';
  import { plural } from '$lib/utils/plural';
  import {
    blankQuestion,
    fromQuestions,
    hasProblems,
    moveItem,
    summariseChanges,
    toSaveDraftInput,
    validateDraft,
    type EditableQuestion
  } from '$lib/utils/questionnaireDraft';

  let { data } = $props();

  const confirm = createConfirm();

  const publishedKeys = $derived(new Set((data.latest?.questions ?? []).map((q) => q.key)));

  let questions = $state<EditableQuestion[]>([]);
  let changeKind = $state<ChangeKind>('substantive');
  let note = $state('');
  let savedSnapshot = $state('');
  let loadedSignature = $state<string | null>(null);
  let showProblems = $state(false);
  let busy = $state(false);
  let errorMessage = $state('');

  // (Re)load the editor whenever the server hands us a different draft, e.g. after start or save.
  $effect(() => {
    const d = data.draft;
    const signature = d ? `${d.id}:${d.updated_at}` : '';
    if (signature === loadedSignature) return;
    loadedSignature = signature;
    questions = d ? fromQuestions(d.questions, publishedKeys) : [];
    changeKind = d?.change_kind ?? 'substantive';
    note = d?.note ?? '';
    savedSnapshot = JSON.stringify(toSaveDraftInput(changeKind, note, questions));
    showProblems = false;
  });

  const problems = $derived(validateDraft(questions));
  const summary = $derived(summariseChanges(data.latest?.questions ?? [], questions));
  const snapshot = $derived(JSON.stringify(toSaveDraftInput(changeKind, note, questions)));
  const dirty = $derived(!!data.draft && snapshot !== savedSnapshot);

  // Don't silently lose edits.
  beforeNavigate(({ cancel }) => {
    if (dirty && !window.confirm('You have unsaved changes to the questionnaire. Leave without saving?')) cancel();
  });
  $effect(() => {
    if (!dirty) return;
    const warn = (e: BeforeUnloadEvent) => e.preventDefault();
    window.addEventListener('beforeunload', warn);
    return () => window.removeEventListener('beforeunload', warn);
  });

  function formatDate(value?: string) {
    if (!value) return '—';
    return new Date(value).toLocaleDateString(undefined, { year: 'numeric', month: 'short', day: 'numeric' });
  }

  function messageOf(err: unknown, fallback: string) {
    return err instanceof Error ? err.message : fallback;
  }

  async function run(action: () => Promise<void>, fallback: string): Promise<boolean> {
    errorMessage = '';
    busy = true;
    try {
      await action();
      return true;
    } catch (err) {
      errorMessage = messageOf(err, fallback);
      return false;
    } finally {
      busy = false;
    }
  }

  async function startDraft() {
    if (await run(() => questionnaireService.startDraft().then(() => invalidateAll()), 'Could not start a draft')) {
      toast.success('Draft started.');
    }
  }

  function addQuestion() {
    questions = [...questions, blankQuestion(questions.map((q) => q.key))];
  }

  function removeQuestion(uid: string) {
    const target = questions.find((q) => q.uid === uid);
    if (!target) return;
    const remove = () => {
      questions = questions.filter((q) => q.uid !== uid);
    };
    if (!target.keyLocked) return remove();
    confirm.ask({
      title: 'Remove this question?',
      description:
        'It disappears from the form and from public profiles. Artists’ existing answers are kept, and admins can still see them on reviews.',
      confirmLabel: 'Remove',
      destructive: true,
      run: remove
    });
  }

  async function save(): Promise<boolean> {
    showProblems = true;
    if (hasProblems(problems)) {
      errorMessage = 'Fix the highlighted questions before saving.';
      return false;
    }
    return run(async () => {
      await questionnaireService.saveDraft(toSaveDraftInput(changeKind, note, questions));
      await invalidateAll();
      toast.success('Draft saved.');
    }, 'Could not save the draft');
  }

  function askPublish() {
    showProblems = true;
    if (hasProblems(problems)) {
      errorMessage = 'Fix the highlighted questions before publishing.';
      return;
    }
    if (summary.unchanged && data.latest) {
      errorMessage = 'Nothing has changed since the published version.';
      return;
    }
    const parts = [
      summary.added.length && `${plural(summary.added.length, 'new question')}`,
      summary.removed.length && `${plural(summary.removed.length, 'removed question')}`,
      summary.rules.length && `${plural(summary.rules.length, 'question')} with changed rules`,
      summary.reworded.length && `${plural(summary.reworded.length, 'reworded question')}`
    ].filter(Boolean);
    confirm.ask({
      title: `Publish as revision ${(data.latest?.number ?? 0) + 1}?`,
      description:
        `${parts.join(', ') || 'First version of the questionnaire'}. ` +
        (changeKind === 'substantive'
          ? 'Artists can be nudged to answer what is new (you choose whether to send it).'
          : 'This is a minor change, so artists will not be nudged.') +
        ' Published revisions can’t be edited, only replaced by a newer one.',
      confirmLabel: 'Publish',
      run: async () => {
        // Publishing uses what is saved on the server, so save first.
        if (dirty && !(await save())) return;
        await run(async () => {
          const revision = await questionnaireService.publish();
          await invalidateAll();
          toast.success(`Published revision ${revision.number}.`);
        }, 'Could not publish');
      }
    });
  }

  function askDiscard() {
    confirm.ask({
      title: 'Discard this draft?',
      description: 'Your unpublished edits are deleted. The published questionnaire is not affected.',
      confirmLabel: 'Discard draft',
      destructive: true,
      run: async () => {
        await run(async () => {
          await questionnaireService.discardDraft();
          await invalidateAll();
          toast.success('Draft discarded.');
        }, 'Could not discard the draft');
      }
    });
  }

  function askNudge() {
    const nudge = data.nudge;
    if (!nudge) return;
    confirm.ask({
      title: `Nudge ${plural(nudge.eligible, 'artist')}?`,
      description:
        'They get one email inviting them to answer the new questions. It is optional: their profiles stay exactly as they are. Nobody is emailed twice for the same revision.',
      confirmLabel: 'Send nudge',
      run: async () => {
        await run(async () => {
          const sent = await questionnaireService.sendNudge();
          await invalidateAll();
          toast.success(sent === 0 ? 'Everyone eligible was already nudged.' : `Nudge sent to ${plural(sent, 'artist')}.`);
        }, 'Could not send the nudge');
      }
    });
  }

  const nudgeReason = $derived(
    !data.nudge
      ? ''
      : !data.nudge.can_send
        ? 'Revision ' + data.nudge.revision_number + ' was a minor change, so artists are not nudged.'
        : data.nudge.eligible === 0
          ? 'Every approved artist has answered all questions.'
          : ''
  );
</script>

<svelte:head>
  <title>Questionnaire — Artiv admin</title>
</svelte:head>

<PageHeader
  eyebrow="Onboarding"
  title="Artist questionnaire"
  description="The questions artists answer when they apply. Publishing creates a new numbered revision; new applicants always answer the latest one."
/>

{#if errorMessage}
  <Alert variant="destructive" class="mb-6">{errorMessage}</Alert>
{/if}

<Card class="mb-6">
  <CardHeader>
    <CardTitle>Published form</CardTitle>
  </CardHeader>
  <CardContent>
    {#if data.latest}
      <div class="flex flex-wrap items-center gap-2 text-sm">
        <Badge variant="success">Revision {data.latest.number}</Badge>
        <Badge variant="secondary">{data.latest.change_kind}</Badge>
        <span class="text-muted-foreground">
          {plural(data.latest.questions.length, 'question')} · published {formatDate(data.latest.published_at)}
        </span>
      </div>
      {#if data.latest.note}<p class="mt-2 text-sm text-muted-foreground">{data.latest.note}</p>{/if}
    {:else}
      <p class="text-sm text-muted-foreground">
        Nothing is published yet, so artists apply with the plain form. Publish a revision to start asking questions.
      </p>
    {/if}
  </CardContent>
</Card>

{#if !data.draft}
  {#if !data.latest}
    <EmptyState
      title="No questionnaire yet"
      description="Start a draft, add your questions, then publish when you're happy with them."
    />
  {/if}
  <div class="mt-4">
    <Button disabled={busy} onclick={startDraft}>{data.latest ? 'Edit questions' : 'Start a draft'}</Button>
  </div>
{:else}
  <div class="space-y-4" data-testid="questionnaire-editor">
    {#if questions.length === 0}
      <EmptyState title="No questions yet" description="Add the first question below." />
    {/if}

    {#each questions as question, i (question.uid)}
      <QuestionEditor
        bind:question={questions[i]}
        index={i}
        total={questions.length}
        problems={showProblems ? (problems[question.uid] ?? []) : []}
        otherKeys={() => questions.filter((q) => q.uid !== question.uid).map((q) => q.key)}
        onmoveup={() => (questions = moveItem(questions, i, i - 1))}
        onmovedown={() => (questions = moveItem(questions, i, i + 1))}
        onremove={() => removeQuestion(question.uid)}
      />
    {/each}

    <Button variant="outline" onclick={addQuestion}><Plus class="h-4 w-4" aria-hidden="true" /> Add a question</Button>

    <Card>
      <CardHeader>
        <CardTitle>About this change</CardTitle>
      </CardHeader>
      <CardContent class="space-y-5">
        <fieldset class="space-y-3">
          <legend class="mb-2 font-mono text-[11px] uppercase tracking-[0.2em] text-muted-foreground">
            What kind of change is it?
          </legend>
          <label class="flex cursor-pointer items-start gap-3 text-sm">
            <input type="radio" class="mt-1 accent-[var(--accent)]" name="change_kind" value="substantive" bind:group={changeKind} />
            <span>
              <span class="font-medium text-foreground">Substantive</span>
              <span class="block text-muted-foreground">A question was added, removed, or its rules changed. You can then nudge artists to answer.</span>
            </span>
          </label>
          <label class="flex cursor-pointer items-start gap-3 text-sm">
            <input type="radio" class="mt-1 accent-[var(--accent)]" name="change_kind" value="minor" bind:group={changeKind} />
            <span>
              <span class="font-medium text-foreground">Minor</span>
              <span class="block text-muted-foreground">Typos or wording only. Nobody is nudged.</span>
            </span>
          </label>
          {#if data.latest && !summary.unchanged && summary.suggested !== changeKind}
            <FormAlert tone="info">
              Based on your edits, <strong>{summary.suggested}</strong> looks right
              {#if summary.suggested === 'substantive'}
                ({[
                  summary.added.length && `${plural(summary.added.length, 'question')} added`,
                  summary.removed.length && `${plural(summary.removed.length, 'question')} removed`,
                  summary.rules.length && `${plural(summary.rules.length, 'question')} with changed rules`
                ]
                  .filter(Boolean)
                  .join(', ')}).
              {:else}
                (wording changes only).
              {/if}
            </FormAlert>
          {/if}
        </fieldset>

        <div class="space-y-2">
          <Label for="note">Note for the team (optional)</Label>
          <AutoTextarea id="note" rows={2} maxlength={1000} bind:value={note} placeholder="e.g. Added a question about influences." />
        </div>

        <div class="flex flex-wrap items-center gap-3">
          <Button variant="secondary" disabled={busy || !dirty} onclick={save}>Save draft</Button>
          <Button disabled={busy} onclick={askPublish}>Publish</Button>
          <Button variant="ghost" disabled={busy} onclick={askDiscard}>Discard draft</Button>
          {#if dirty}<span class="text-sm text-muted-foreground">Unsaved changes</span>{/if}
        </div>
      </CardContent>
    </Card>
  </div>
{/if}

{#if data.nudge}
  <div class="mt-10" data-testid="nudge-card">
  <Card>
    <CardHeader>
      <CardTitle>Nudge artists</CardTitle>
    </CardHeader>
    <CardContent class="space-y-3">
      <p class="text-sm text-foreground">
        {plural(data.nudge.eligible, 'approved artist')} {data.nudge.eligible === 1 ? 'has' : 'have'} unanswered questions
        in revision {data.nudge.revision_number}.
      </p>
      <p class="text-sm text-muted-foreground">
        A nudge is one optional email. It never changes anyone's profile, and nobody is emailed twice for the same revision.
      </p>
      {#if nudgeReason}<p class="text-sm text-muted-foreground">{nudgeReason}</p>{/if}
      <Button variant="secondary" disabled={busy || !data.nudge.can_send || data.nudge.eligible === 0} onclick={askNudge}>
        Send nudge
      </Button>
    </CardContent>
  </Card>
  </div>
{/if}

{#if data.revisions.length > 0}
  <h2 class="mb-3 mt-10 font-display text-xl text-foreground">History</h2>
  <Table>
    <TableHeader>
      <TableRow>
        <TableHead>Revision</TableHead>
        <TableHead>Kind</TableHead>
        <TableHead>Published</TableHead>
        <TableHead>Note</TableHead>
      </TableRow>
    </TableHeader>
    <TableBody>
      {#each data.revisions as revision (revision.id)}
        <TableRow>
          <TableCell class="font-medium">#{revision.number}</TableCell>
          <TableCell><Badge variant="secondary">{revision.change_kind}</Badge></TableCell>
          <TableCell class="text-muted-foreground">{formatDate(revision.published_at)}</TableCell>
          <TableCell class="text-muted-foreground">{revision.note || '—'}</TableCell>
        </TableRow>
      {/each}
    </TableBody>
  </Table>
{/if}

<ConfirmDialog bind:open={confirm.open} {...confirm.dialog} />
