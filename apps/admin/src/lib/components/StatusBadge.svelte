<script lang="ts">
  import { Badge } from '$lib/components/ui/badge';

  let { status, class: className }: { status?: string; class?: string } = $props();

  const variants: Record<string, 'success' | 'warning' | 'muted' | 'destructive'> = {
    approved: 'success',
    published: 'success',
    pending: 'warning',
    draft: 'muted',
    archived: 'muted',
    rejected: 'destructive'
  };
</script>

<Badge variant={variants[status ?? ''] ?? 'muted'} class={className}>{status ?? 'draft'}</Badge>
