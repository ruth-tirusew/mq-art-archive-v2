<script lang="ts">
  import { onboardingService } from '$lib/application/onboarding';
  import { goto } from '$app/navigation';

  let { data } = $props();

  let notes = $state('');
  let submitting = $state(false);
  let errorMessage = $state('');

  async function review(status: 'approved' | 'rejected') {
    submitting = true;
    errorMessage = '';
    try {
      await onboardingService.review(data.application.id, status, notes);
      await goto('/applications');
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Review failed';
    } finally {
      submitting = false;
    }
  }
</script>

<svelte:head>
  <title>{data.application.display_name} — mq admin</title>
</svelte:head>

<h1>{data.application.display_name}</h1>
<p>Type: {data.application.applicant_type}</p>
<p>Status: {data.application.status}</p>

{#if data.application.status === 'pending'}
  <label>
    Notes
    <textarea bind:value={notes} rows="4"></textarea>
  </label>

  <div>
    <button disabled={submitting} onclick={() => review('approved')}>Approve</button>
    <button disabled={submitting} onclick={() => review('rejected')}>Reject</button>
  </div>
{/if}

{#if errorMessage}
  <p role="alert">{errorMessage}</p>
{/if}

<p><a href="/applications">Back to list</a></p>

<style>
  textarea {
    display: block;
    width: 100%;
    margin: 0.5rem 0 1rem;
  }

  button {
    margin-right: 0.5rem;
  }
</style>
