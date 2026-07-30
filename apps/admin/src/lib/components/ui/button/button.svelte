<script lang="ts">
  import { cn } from '$lib/utils';

  type Variant = 'default' | 'destructive' | 'outline' | 'secondary' | 'ghost' | 'link';
  type Size = 'default' | 'sm' | 'lg' | 'icon';

  type Props = {
    variant?: Variant;
    size?: Size;
    class?: string;
    href?: string;
    type?: 'button' | 'submit' | 'reset';
    disabled?: boolean;
    onclick?: (e: MouseEvent) => void;
    children: import('svelte').Snippet;
  };

  let {
    variant = 'default',
    size = 'default',
    class: className,
    href,
    type = 'button',
    disabled = false,
    onclick,
    children
  }: Props = $props();

  const variants: Record<Variant, string> = {
    default: 'bg-primary text-primary-foreground hover:bg-primary/90',
    destructive: 'bg-destructive text-destructive-foreground hover:bg-destructive/90',
    outline: 'border border-border bg-background hover:bg-muted',
    secondary: 'bg-secondary text-secondary-foreground hover:bg-secondary/80',
    ghost: 'hover:bg-muted',
    link: 'text-primary underline-offset-4 hover:underline'
  };

  const sizes: Record<Size, string> = {
    default: 'h-9 px-4 py-2',
    sm: 'h-8 rounded-md px-3 text-xs',
    lg: 'h-10 rounded-md px-8',
    icon: 'h-9 w-9'
  };

  const classes = $derived(
    cn(
      'inline-flex items-center justify-center gap-2 whitespace-nowrap rounded-md text-sm font-medium transition-colors focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring disabled:pointer-events-none disabled:opacity-50',
      variants[variant],
      sizes[size],
      className
    )
  );
</script>

{#if href}
  <a {href} class={classes}>
    {@render children()}
  </a>
{:else}
  <button {type} {disabled} class={classes} {onclick}>
    {@render children()}
  </button>
{/if}
