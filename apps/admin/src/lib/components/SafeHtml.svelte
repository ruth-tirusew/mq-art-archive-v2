<script lang="ts">
  import { cn } from '$lib/utils';

  /**
   * Renders HTML that the API has ALREADY sanitised: rich text such as questionnaire answers (format "rich"),
   * which the server cleans to a small allowlist on write and again on read.
   *
   * This is the only place in the app that uses {@html}. Never pass it anything else: not plain
   * answers, not names, not notes. Plain text goes through normal Svelte escaping.
   */
  let { html, lang, class: className }: { html: string; lang?: string; class?: string } = $props();
</script>

<div
  {lang}
  class={cn(
    'text-sm leading-relaxed text-foreground [&_a]:underline [&_a]:underline-offset-4 [&_blockquote]:border-l-2 [&_blockquote]:border-border [&_blockquote]:pl-3 [&_blockquote]:text-muted-foreground [&_ol]:list-decimal [&_ol]:pl-5 [&_p]:my-2 [&_p:first-child]:mt-0 [&_p:last-child]:mb-0 [&_ul]:list-disc [&_ul]:pl-5 [&_h2]:mt-6 [&_h2]:mb-2 [&_h2]:font-display [&_h2]:text-[1.35em] [&_h2]:leading-snug [&_h3]:mt-5 [&_h3]:mb-1.5 [&_h3]:font-semibold [&_h3]:text-[1.1em] [&_h2:first-child]:mt-0 [&_h3:first-child]:mt-0 [&_u]:underline [&_u]:underline-offset-2 [&_s]:line-through',
    className
  )}
>
  {@html html}
</div>
