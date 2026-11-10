<script lang="ts">
  import { fieldClass } from '$ui';
  import { cn } from '$lib/utils';

  type Props = {
    value: string;
    id?: string;
    lang?: string;
    maxlength?: number;
    /** Smallest height, in text rows. The box grows from here to fit its text. */
    rows?: number;
    placeholder?: string;
    invalid?: boolean;
    /**
     * For text that is one line of prose (a question title): Enter does not add a line break, and line
     * breaks that are pasted in become spaces. The box still wraps and grows like any textarea.
     */
    singleLine?: boolean;
    /** Called after each edit, once `value` already holds the new text. */
    oninput?: () => void;
    class?: string;
    'aria-describedby'?: string;
  };

  let {
    value = $bindable(),
    id,
    lang,
    maxlength,
    rows = 2,
    placeholder,
    invalid = false,
    singleLine = false,
    oninput,
    class: className,
    'aria-describedby': describedby
  }: Props = $props();

  let node = $state<HTMLTextAreaElement>();

  // Size the box to its content instead of giving it a fixed height with an inner scrollbar. Run when
  // the text changes (typing, or the editor loading a draft) and when the width changes (text rewraps).
  function fit() {
    if (!node) return;
    node.style.height = 'auto';
    // scrollHeight covers content + padding but not the border, and the box is border-box, so add
    // the border back or the box ends up a couple of pixels short and its text overflows.
    const border = node.offsetHeight - node.clientHeight;
    node.style.height = `${node.scrollHeight + border}px`;
  }

  $effect(() => {
    void value;
    fit();
  });

  function handleInput() {
    if (!node) return;
    if (singleLine && /[\r\n]/.test(node.value)) {
      const caret = node.selectionStart;
      node.value = node.value.replace(/\r?\n/g, ' ');
      node.setSelectionRange(caret, caret);
    }
    value = node.value;
    fit();
    oninput?.();
  }

  function handleKeydown(event: KeyboardEvent) {
    if (singleLine && event.key === 'Enter') event.preventDefault();
  }

  $effect(() => {
    if (!node) return;
    let width = node.clientWidth;
    const observer = new ResizeObserver(() => {
      if (node && node.clientWidth !== width) {
        width = node.clientWidth;
        fit();
      }
    });
    observer.observe(node);
    return () => observer.disconnect();
  });
</script>

<textarea
  bind:this={node}
  bind:value
  {id}
  {lang}
  {maxlength}
  {rows}
  {placeholder}
  aria-describedby={describedby}
  aria-invalid={invalid || undefined}
  oninput={handleInput}
  onkeydown={handleKeydown}
  class={cn(fieldClass, 'resize-none overflow-hidden leading-relaxed', className)}
></textarea>
