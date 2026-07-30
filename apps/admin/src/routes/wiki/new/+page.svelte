<script lang="ts">
  import { goto } from '$app/navigation';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import { articlesService } from '$lib/application/articles';
  import type { ArticleStatus } from '$lib/core/domain/article';
  import { Alert } from '$lib/components/ui/alert';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent } from '$lib/components/ui/card';
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from '$lib/components/ui/textarea';

  const inputClass =
    'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50';

  let title = $state('');
  let body = $state('');
  let category = $state('General');
  let excerpt = $state('');
  let difficulty = $state('Beginner');
  let verified = $state(false);
  let status = $state<ArticleStatus>('draft');
  let saving = $state(false);
  let errorMessage = $state('');

  async function handleSubmit(event: Event) {
    event.preventDefault();
    saving = true;
    errorMessage = '';
    try {
      const article = await articlesService.create({
        title,
        body,
        category,
        excerpt,
        difficulty,
        verified,
        status
      });
      await goto(`/wiki/${article.id}`);
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Create failed';
    } finally {
      saving = false;
    }
  }
</script>

<svelte:head>
  <title>New article — mq admin</title>
</svelte:head>

<PageHeader
  eyebrow="06 — Wiki"
  title="New article"
  description="Draft a handbook entry. Publish when it is ready for the public wiki."
/>

<Card>
  <CardContent class="pt-6">
    <form class="space-y-5" onsubmit={handleSubmit}>
      <div class="space-y-2">
        <Label for="title">Title</Label>
        <input id="title" required bind:value={title} class={inputClass} />
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label for="category">Category</Label>
          <input id="category" bind:value={category} class={inputClass} />
        </div>
        <div class="space-y-2">
          <Label for="difficulty">Difficulty</Label>
          <select id="difficulty" bind:value={difficulty} class={inputClass}>
            <option>Beginner</option>
            <option>Intermediate</option>
            <option>Advanced</option>
          </select>
        </div>
      </div>

      <div class="space-y-2">
        <Label for="excerpt">Excerpt</Label>
        <input id="excerpt" bind:value={excerpt} class={inputClass} placeholder="Short summary" />
      </div>

      <div class="space-y-2">
        <Label for="body">Body</Label>
        <Textarea
          id="body"
          rows={14}
          bind:value={body}
          placeholder="Separate paragraphs with a blank line."
          class="font-mono text-sm"
        />
      </div>

      <div class="grid gap-4 sm:grid-cols-2">
        <div class="space-y-2">
          <Label for="status">Status</Label>
          <select id="status" bind:value={status} class={inputClass}>
            <option value="draft">draft</option>
            <option value="published">published</option>
            <option value="archived">archived</option>
          </select>
        </div>
        <div class="flex items-end gap-2 pb-1">
          <input id="verified" type="checkbox" bind:checked={verified} class="h-4 w-4" />
          <Label for="verified">Verified</Label>
        </div>
      </div>

      {#if errorMessage}
        <Alert variant="destructive">{errorMessage}</Alert>
      {/if}

      <div class="flex flex-wrap gap-2">
        <Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Create article'}</Button>
        <Button type="button" variant="ghost" onclick={() => goto('/wiki')}>Cancel</Button>
      </div>
    </form>
  </CardContent>
</Card>
