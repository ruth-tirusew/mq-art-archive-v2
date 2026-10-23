<script lang="ts">
  import { ConfirmDialog, createConfirm, toast } from '$ui';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { goto } from '$app/navigation';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import { onboardingService } from '$lib/application/onboarding';
  import { Alert } from '$lib/components/ui/alert';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from '$lib/components/ui/textarea';

  let { data } = $props();

  const confirm = createConfirm();
  let notes = $state('');
  let submitting = $state(false);
  let errorMessage = $state('');

  function formatDate(value?: string | null) {
    if (!value) return '—';
    return new Date(value).toLocaleString(undefined, {
      year: 'numeric',
      month: 'short',
      day: 'numeric',
      hour: '2-digit',
      minute: '2-digit'
    });
  }

  function askReview(status: 'approved' | 'rejected') {
    errorMessage = '';
    if (status === 'rejected' && !notes.trim()) {
      errorMessage = 'Add a reason before rejecting. The applicant will see it.';
      return;
    }
    confirm.ask(
      status === 'approved'
        ? {
            title: 'Approve this application?',
            description: 'The applicant is notified of the decision and gets access to the studio.',
            confirmLabel: 'Approve',
            run: () => review(status)
          }
        : {
            title: 'Reject this application?',
            description: 'The applicant is notified and sees your note.',
            confirmLabel: 'Reject',
            destructive: true,
            run: () => review(status)
          }
    );
  }

  async function review(status: 'approved' | 'rejected') {
    errorMessage = '';
    submitting = true;
    try {
      await onboardingService.review(data.application.id, status, notes);
      toast.success(status === 'approved' ? 'Application approved.' : 'Application rejected.');
      await goto('/applications');
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Review failed';
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>{data.application.display_name} — Artiv admin</title>
</svelte:head>

<PageHeader
  eyebrow="Applications"
  title={data.application.display_name}
  description="Approve or reject this application. Your note is shown to the applicant."
/>

<Card>
  <CardHeader>
    <div class="flex flex-wrap items-center gap-2">
      <Badge variant="secondary">{data.application.applicant_type}</Badge>
      <StatusBadge status={data.application.status} />
    </div>
    <CardTitle class="pt-2">{data.application.display_name}</CardTitle>
  </CardHeader>
  <CardContent class="space-y-6">
    <dl class="grid gap-3 text-sm sm:grid-cols-2">
      <div>
        <dt class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Application ID</dt>
        <dd class="mt-1 break-all font-mono text-xs text-foreground">{data.application.id}</dd>
      </div>
      <div>
        <dt class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Applicant ID</dt>
        <dd class="mt-1 break-all font-mono text-xs text-foreground">{data.application.applicant_id ?? '—'}</dd>
      </div>
      <div>
        <dt class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Requested handle</dt>
        <dd class="mt-1 font-mono text-xs text-foreground">
          {data.application.requested_handle ? `@${data.application.requested_handle}` : '—'}
        </dd>
      </div>
      <div>
        <dt class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Created</dt>
        <dd class="mt-1 text-foreground">{formatDate(data.application.created_at)}</dd>
      </div>
      <div>
        <dt class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Updated</dt>
        <dd class="mt-1 text-foreground">{formatDate(data.application.updated_at)}</dd>
      </div>
      <div>
        <dt class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Reviewed at</dt>
        <dd class="mt-1 text-foreground">{formatDate(data.application.reviewed_at)}</dd>
      </div>
      <div>
        <dt class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Reviewed by</dt>
        <dd class="mt-1 break-all font-mono text-xs text-foreground">{data.application.reviewed_by ?? '—'}</dd>
      </div>
    </dl>

    {#if data.application.notes}
      <div>
        <p class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Applicant notes</p>
        <p class="mt-2 whitespace-pre-wrap text-sm text-foreground">{data.application.notes}</p>
      </div>
    {/if}

    {#if data.application.review_notes}
      <div>
        <p class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">Note sent to applicant</p>
        <p class="mt-2 whitespace-pre-wrap text-sm text-foreground">{data.application.review_notes}</p>
      </div>
    {/if}

    {#if data.application.status === 'pending'}
      <div class="space-y-2">
        <Label for="notes">Note to applicant</Label>
        <Textarea id="notes" bind:value={notes} placeholder="Required when rejecting (explain what to fix). Optional when approving, e.g. a welcome message." />
      </div>

      <div class="flex flex-wrap gap-3">
        <Button disabled={submitting} onclick={() => askReview('approved')}>Approve</Button>
        <Button variant="destructive" disabled={submitting} onclick={() => askReview('rejected')}>Reject</Button>
      </div>
    {/if}

    {#if errorMessage}
      <Alert variant="destructive">{errorMessage}</Alert>
    {/if}
  </CardContent>
</Card>
<ConfirmDialog bind:open={confirm.open} {...confirm.dialog} />
