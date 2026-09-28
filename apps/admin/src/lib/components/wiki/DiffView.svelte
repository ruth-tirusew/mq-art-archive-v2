<script lang="ts">
  import { diffLines } from '$lib/utils/diffLines';

  let { oldText, newText }: { oldText: string; newText: string } = $props();

  let lines = $derived(diffLines(oldText, newText));
</script>

<div
  data-testid="admin-wiki-diff"
  class="max-h-96 overflow-y-auto rounded-md border border-border/70 bg-muted/20 font-mono text-xs leading-relaxed"
>
  {#each lines as line, i (i)}
    <div
      data-testid="admin-wiki-diff-line"
      data-diff-type={line.type}
      class="whitespace-pre-wrap px-3 py-0.5 {line.type === 'insert'
        ? 'bg-emerald-500/10 text-emerald-700 dark:text-emerald-400'
        : line.type === 'delete'
          ? 'bg-red-500/10 text-red-700 line-through decoration-red-700/40 dark:text-red-400'
          : 'text-muted-foreground'}"
    >
      <span class="mr-2 select-none opacity-50">{line.type === 'insert' ? '+' : line.type === 'delete' ? '-' : ' '}</span>{line.value ||
        ' '}
    </div>
  {/each}
</div>
