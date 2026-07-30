<script lang="ts">
  import { cn } from '$lib/utils';

  let {
    class: className,
    value = $bindable(''),
    rows = 4,
    id,
    placeholder
  }: {
    class?: string;
    value?: string;
    rows?: number;
    id?: string;
    placeholder?: string;
  } = $props();
</script>

<textarea
  {id}
  {rows}
  {placeholder}
  bind:value
  class={cn(
    'flex min-h-[80px] w-full rounded-sm border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50',
    className
  )}
></textarea>
