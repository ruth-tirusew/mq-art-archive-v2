<script lang="ts">
	import { getSelectionAnchor, type TextAnchor } from '$lib/utils/textAnchor';

	let {
		container,
		body,
		onHighlight,
		onComment
	}: {
		container: HTMLElement | undefined;
		body: string;
		onHighlight: (anchor: TextAnchor) => void;
		onComment: (anchor: TextAnchor) => void;
	} = $props();

	let visible = $state(false);
	let top = $state(0);
	let left = $state(0);
	let pendingAnchor = $state<TextAnchor | null>(null);

	function handleSelectionChange() {
		if (!container) return;
		const anchor = getSelectionAnchor(container, body);
		if (!anchor) {
			visible = false;
			pendingAnchor = null;
			return;
		}
		const selection = window.getSelection();
		const range = selection?.rangeCount ? selection.getRangeAt(0) : null;
		if (!range) return;
		const rect = range.getBoundingClientRect();
		if (rect.width === 0 && rect.height === 0) return;

		// position: fixed is already viewport-relative — do NOT add window.scrollX/scrollY
		// here (that double-counts scroll and pushes the toolbar off-screen for any
		// selection below the first screenful).
		pendingAnchor = anchor;
		top = rect.top - 44;
		left = rect.left + rect.width / 2;
		visible = true;
	}

	$effect(() => {
		document.addEventListener('selectionchange', handleSelectionChange);
		return () => document.removeEventListener('selectionchange', handleSelectionChange);
	});

	function act(fn: (anchor: TextAnchor) => void) {
		if (!pendingAnchor) return;
		fn(pendingAnchor);
		visible = false;
		window.getSelection()?.removeAllRanges();
	}
</script>

{#if visible && pendingAnchor}
	<div
		data-testid="web-wiki-selection-toolbar"
		class="fixed z-50 flex -translate-x-1/2 items-center gap-1 rounded-full border border-border bg-foreground px-1.5 py-1 text-background shadow-lg"
		style="top: {top}px; left: {left}px;"
	>
		<button
			type="button"
			data-testid="web-wiki-selection-highlight"
			onclick={() => act(onHighlight)}
			class="flex items-center gap-1 rounded-full px-2.5 py-1 font-mono text-[10px] uppercase tracking-[0.15em] transition hover:bg-background/20"
		>
			<span aria-hidden="true">★</span> Highlight
		</button>
		<span class="h-4 w-px bg-background/30"></span>
		<button
			type="button"
			data-testid="web-wiki-selection-comment"
			onclick={() => act(onComment)}
			class="flex items-center gap-1 rounded-full px-2.5 py-1 font-mono text-[10px] uppercase tracking-[0.15em] transition hover:bg-background/20"
		>
			<span aria-hidden="true">💬</span> Comment
		</button>
	</div>
{/if}
