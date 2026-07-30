<script lang="ts">
  import { cn } from '$lib/utils';

  const variants = {
    lily: '/adornments/lily.png',
    slender: '/adornments/slender.png',
    sprig: '/adornments/sprig.png',
    grass: '/adornments/grass.png',
    buds: '/adornments/buds.png',
    peony: '/adornments/peony.png',
    bouquet: '/adornments/bouquet.png',
    daisies: '/adornments/daisies.png'
  } as const;

  type Variant = keyof typeof variants;

  let {
    variant = 'lily',
    class: className = '',
    alt = ''
  }: {
    variant?: Variant;
    class?: string;
    alt?: string;
  } = $props();
</script>

<img
  src={variants[variant]}
  {alt}
  aria-hidden={alt ? undefined : true}
  class={cn('pointer-events-none select-none', className)}
  draggable="false"
/>
