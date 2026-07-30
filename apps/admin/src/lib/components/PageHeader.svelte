<script lang="ts">
  let {
    eyebrow,
    title,
    description
  }: {
    eyebrow: string;
    title: string;
    description?: string;
  } = $props();
</script>

<header class="mb-8 space-y-2">
  <p class="font-mono text-[11px] uppercase tracking-[0.3em] text-muted-foreground">{eyebrow}</p>
  <h1 class="font-display text-3xl text-foreground md:text-4xl">{title}</h1>
  {#if description}
    <p class="max-w-2xl text-sm text-muted-foreground">{description}</p>
  {/if}
</header>
