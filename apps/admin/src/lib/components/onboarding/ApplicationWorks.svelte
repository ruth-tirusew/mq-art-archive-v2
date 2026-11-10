<script lang="ts">
  import { MIN_APPLICATION_WORKS, type ApplicantWork } from '$lib/core/domain/onboarding';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { deliveryUrl } from '$lib/utils/cloudinary';

  let { works }: { works: ApplicantWork[] } = $props();

  const enough = $derived(works.length >= MIN_APPLICATION_WORKS);
</script>

<div class="mt-6" data-testid="application-works">
<Card>
  <CardHeader>
    <CardTitle>Submitted work</CardTitle>
    <p class="text-sm text-muted-foreground">
      {works.length} {works.length === 1 ? 'piece' : 'pieces'}, as sent with the application.
      {#if !enough && works.length > 0}
        <span class="text-warning">Fewer than the {MIN_APPLICATION_WORKS} now required: this application was made before that rule.</span>
      {/if}
    </p>
  </CardHeader>
  <CardContent>
    {#if works.length === 0}
      <p class="text-sm text-muted-foreground">No work was attached to this application.</p>
    {:else}
      <ul class="grid grid-cols-2 gap-4 sm:grid-cols-3 lg:grid-cols-5">
        {#each works as work (work.id)}
          <li>
            <a href={work.url} target="_blank" rel="noopener noreferrer" class="block aspect-[4/5] overflow-hidden rounded-sm bg-muted">
              <img src={deliveryUrl(work.url, 480)} alt={work.title} loading="lazy" class="h-full w-full object-cover" />
            </a>
            <p class="mt-2 truncate text-sm text-foreground">{work.title}</p>
            <p class="truncate font-mono text-[10px] uppercase tracking-[0.15em] text-muted-foreground">
              {[work.medium, work.year].filter(Boolean).join(' · ')}
            </p>
          </li>
        {/each}
      </ul>
    {/if}
  </CardContent>
</Card>
</div>
