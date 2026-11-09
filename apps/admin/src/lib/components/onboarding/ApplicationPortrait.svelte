<script lang="ts">
  import type { ApplicantPortrait } from '$lib/core/domain/onboarding';
  import { Card, CardContent, CardHeader, CardTitle } from '$lib/components/ui/card';
  import { deliveryUrl } from '$lib/utils/cloudinary';

  let { portrait }: { portrait: ApplicantPortrait | null | undefined } = $props();
</script>

<div class="mt-6" data-testid="application-portrait">
<Card>
  <CardHeader>
    <CardTitle>Portrait</CardTitle>
    <p class="text-sm text-muted-foreground">
      The photo the applicant sent. On approval it becomes their profile photo.
    </p>
  </CardHeader>
  <CardContent>
    {#if portrait}
      <a href={portrait.url} target="_blank" rel="noopener noreferrer" class="block h-32 w-32 overflow-hidden rounded-full bg-muted">
        <img src={deliveryUrl(portrait.url, 256)} alt="Applicant portrait" loading="lazy" class="h-full w-full object-cover" />
      </a>
    {:else}
      <p class="text-sm text-muted-foreground">No portrait was attached to this application.</p>
    {/if}
  </CardContent>
</Card>
</div>
