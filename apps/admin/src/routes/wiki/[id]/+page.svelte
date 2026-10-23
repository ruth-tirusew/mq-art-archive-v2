<script lang="ts">
  import { ConfirmDialog, createConfirm, toast } from '$ui';
  import StatusBadge from '$lib/components/StatusBadge.svelte';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import { goto } from '$app/navigation';
  import PageHeader from '$lib/components/PageHeader.svelte';
  import { articlesService } from '$lib/application/articles';
  import type { ArticleRevision, ArticleStatus } from '$lib/core/domain/article';
  import { Alert } from '$lib/components/ui/alert';
  import { Badge } from '$lib/components/ui/badge';
  import { Button } from '$lib/components/ui/button';
  import { Card, CardContent } from '$lib/components/ui/card';
  import { Label } from '$lib/components/ui/label';
  import { Textarea } from '$lib/components/ui/textarea';
  import { onMount } from 'svelte';

  const inputClass =
    'flex h-9 w-full rounded-md border border-input bg-transparent px-3 py-1 text-sm shadow-xs outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50';

  let { data } = $props();

  let article = $state(data.article);
  let title = $state(data.article.title);
  let body = $state(data.article.body ?? '');
  let category = $state(data.article.category ?? 'General');
  let excerpt = $state(data.article.excerpt ?? '');
  let difficulty = $state(data.article.difficulty ?? 'Beginner');
  let verified = $state(!!data.article.verified);
  const confirm = createConfirm();
  let saving = $state(false);
  let errorMessage = $state('');

  let revisions = $state<ArticleRevision[]>([]);
  let revisionsLoading = $state(false);
  let selectedVersion = $state<number | null>(null);
  let preview = $state<ArticleRevision | null>(null);
  let previewLoading = $state(false);
  let restoring = $state(false);

  function applyArticle(next: typeof article) {
    article = next;
    title = next.title;
    body = next.body ?? '';
    category = next.category ?? 'General';
    excerpt = next.excerpt ?? '';
    difficulty = next.difficulty ?? 'Beginner';
    verified = !!next.verified;
  }

  async function loadRevisions() {
    revisionsLoading = true;
    try {
      revisions = await articlesService.listRevisions(article.id);
    } catch {
      revisions = [];
    } finally {
      revisionsLoading = false;
    }
  }

  onMount(() => {
    void loadRevisions();
  });

  async function selectRevision(version: number) {
    selectedVersion = version;
    previewLoading = true;
    preview = null;
    try {
      preview = await articlesService.getRevision(article.id, version);
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Failed to load revision';
    } finally {
      previewLoading = false;
    }
  }

  function askRestore() {
    if (selectedVersion == null) return;
    confirm.ask({
      title: `Restore version ${selectedVersion}?`,
      description: 'Current content is saved as a new revision first, so nothing is lost.',
      confirmLabel: 'Restore',
      run: restoreSelected
    });
  }

  async function restoreSelected() {
    if (selectedVersion == null) return;
    restoring = true;
    errorMessage = '';
    try {
      const restoredFrom = selectedVersion;
      applyArticle(await articlesService.restoreRevision(article.id, selectedVersion));
      selectedVersion = null;
      preview = null;
      await loadRevisions();
      toast.success(`Restored from v${restoredFrom}. Now at version ${article.version}.`);
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Restore failed';
    } finally {
      restoring = false;
    }
  }

  function truncateId(id: string) {
    return id.length > 8 ? `${id.slice(0, 8)}…` : id;
  }

  function formatTime(iso: string) {
    try {
      return new Date(iso).toLocaleString();
    } catch {
      return iso;
    }
  }

  async function saveContent(event: Event) {
    event.preventDefault();
    saving = true;
    errorMessage = '';
    try {
      applyArticle(
        await articlesService.update(article.id, {
          title,
          body,
          category,
          excerpt,
          difficulty,
          verified
        })
      );
      await loadRevisions();
      toast.success('Article saved.');
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Save failed';
    } finally {
      saving = false;
    }
  }

  function askStatus(status: ArticleStatus) {
    const copy = {
      published: { title: 'Publish this article?', description: 'It goes live on the public wiki immediately.', confirmLabel: 'Publish', destructive: false },
      draft: { title: 'Unpublish this article?', description: 'It disappears from the public wiki until you publish it again.', confirmLabel: 'Unpublish', destructive: true },
      archived: { title: 'Archive this article?', description: 'It is removed from the public wiki and hidden from lists.', confirmLabel: 'Archive', destructive: true }
    } as Record<string, { title: string; description: string; confirmLabel: string; destructive: boolean }>;
    confirm.ask({ ...copy[status], run: () => setStatus(status) });
  }

  async function setStatus(status: ArticleStatus) {
    saving = true;
    errorMessage = '';
    try {
      applyArticle(await articlesService.patch(article.id, { status }));
      toast.success(status === 'published' ? 'Published to the public wiki.' : `Status set to ${status}.`);
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Update failed';
    } finally {
      saving = false;
    }
  }

  async function toggleVerified() {
    saving = true;
    errorMessage = '';
    try {
      const next = !article.verified;
      applyArticle(await articlesService.patch(article.id, { verified: next }));
      verified = article.verified ?? next;
      toast.success(verified ? 'Marked verified.' : 'Unverified.');
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Update failed';
    } finally {
      saving = false;
    }
  }
</script>

<svelte:head>
  <title>{article.title} — Artiv admin</title>
</svelte:head>

<PageHeader
  eyebrow="Wiki"
  title={article.title}
  description="Edit content, then publish when ready for the public wiki."
/>

<div class="mb-4 flex flex-wrap items-center gap-2">
  <StatusBadge status={article.status} />
  {#if article.verified}<Badge variant="secondary">Verified</Badge>{/if}
  <Badge variant="outline">v{article.version ?? 1}</Badge>
  <span class="font-mono text-[10px] uppercase tracking-[0.2em] text-muted-foreground">
    /{article.slug}
  </span>
</div>

<div class="grid gap-6 lg:grid-cols-[minmax(0,1fr)_minmax(16rem,20rem)]">
  <Card>
    <CardContent class="space-y-5 pt-6">
      <form class="space-y-5" onsubmit={saveContent}>
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
          <input id="excerpt" bind:value={excerpt} class={inputClass} />
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

        <div class="flex flex-wrap gap-2">
          <Button type="submit" disabled={saving}>{saving ? 'Saving…' : 'Save changes'}</Button>
          <Button type="button" disabled={saving} onclick={() => askStatus('published')}>
            Publish
          </Button>
          <Button type="button" variant="secondary" disabled={saving} onclick={() => askStatus('draft')}>
            Draft
          </Button>
          <Button
            type="button"
            variant="secondary"
            disabled={saving}
            onclick={() => askStatus('archived')}
          >
            Archive
          </Button>
          <Button type="button" variant="outline" disabled={saving} onclick={toggleVerified}>
            {article.verified ? 'Unverify' : 'Verify'}
          </Button>
        </div>
      </form>

      {#if errorMessage}<Alert variant="destructive">{errorMessage}</Alert>{/if}

      <Button variant="ghost" onclick={() => goto('/wiki')}>Back to list</Button>
    </CardContent>
  </Card>

  <Card>
    <CardContent class="space-y-4 pt-6">
      <div>
        <h2 class="text-sm font-medium tracking-wide">History</h2>
        <p class="mt-1 text-xs text-muted-foreground">
          Snapshots from each content save. Status-only changes are not versioned.
        </p>
      </div>

      {#if revisionsLoading}
        <div class="space-y-2" role="status"><span class="sr-only">Loading revisions</span><Skeleton class="h-8 w-full" /><Skeleton class="h-8 w-full" /></div>
      {:else if revisions.length === 0}
        <p class="text-xs text-muted-foreground">No prior revisions yet.</p>
      {:else}
        <ul class="max-h-64 space-y-1 overflow-y-auto border-y border-border py-2">
          {#each revisions as rev (rev.version)}
            <li>
              <button
                type="button"
                class="flex w-full flex-col gap-0.5 rounded-md px-2 py-1.5 text-left text-xs transition-colors hover:bg-muted {selectedVersion ===
                rev.version
                  ? 'bg-muted'
                  : ''}"
                onclick={() => selectRevision(rev.version)}
              >
                <span class="font-medium">v{rev.version}</span>
                <span class="text-muted-foreground">{formatTime(rev.created_at)}</span>
                <span class="font-mono text-[10px] text-muted-foreground">
                  {truncateId(rev.editor_id)}
                </span>
              </button>
            </li>
          {/each}
        </ul>
      {/if}

      {#if previewLoading}
        <div class="space-y-2" role="status"><span class="sr-only">Loading preview</span><Skeleton class="h-24 w-full" /></div>
      {:else if preview}
        <div class="space-y-2 border-t border-border pt-4">
          <p class="text-xs font-medium">Preview v{preview.version}</p>
          <p class="text-sm font-medium">{preview.title}</p>
          <pre
            class="max-h-48 overflow-y-auto whitespace-pre-wrap rounded-md bg-muted/50 p-2 font-mono text-[11px] leading-relaxed text-muted-foreground"
          >{preview.body || '(empty)'}</pre>
          <Button
            type="button"
            variant="secondary"
            size="sm"
            disabled={restoring || saving}
            onclick={askRestore}
          >
            {restoring ? 'Restoring…' : `Restore v${preview.version}`}
          </Button>
        </div>
      {/if}
    </CardContent>
  </Card>
</div>

<ConfirmDialog bind:open={confirm.open} {...confirm.dialog} />
