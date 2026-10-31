<script lang="ts">
	import { onMount } from 'svelte';
	import { cn } from '$lib/utils';
	import { safeHttpUrl } from '$lib/utils/answers';

	type Props = {
		/** Initial HTML. The editor owns the content afterwards; it is not re-synced from here. */
		value: string;
		/** Accessible name, usually the question. */
		label: string;
		onchange: (html: string) => void;
		disabled?: boolean;
		invalid?: boolean;
		describedby?: string;
		lang?: string;
		class?: string;
	};

	let { value, label, onchange, disabled = false, invalid = false, describedby, lang, class: className }: Props = $props();

	type EditorLike = import('@tiptap/core').Editor;

	let host = $state<HTMLDivElement>();
	let editor = $state.raw<EditorLike>();
	let failed = $state(false);
	let active = $state({ bold: false, italic: false, link: false, bullet: false, ordered: false, quote: false });
	let linkOpen = $state(false);
	let linkValue = $state('');
	let linkError = $state('');
	let fallbackText = $state('');

	function sync() {
		if (!editor) return;
		active = {
			bold: editor.isActive('bold'),
			italic: editor.isActive('italic'),
			link: editor.isActive('link'),
			bullet: editor.isActive('bulletList'),
			ordered: editor.isActive('orderedList'),
			quote: editor.isActive('blockquote')
		};
	}

	const escapeHtml = (s: string) =>
		s.replace(/&/g, '&amp;').replace(/</g, '&lt;').replace(/>/g, '&gt;').replace(/"/g, '&quot;');

	/** Only used if the editor library can't load: plain paragraphs, escaped, so nothing is lost. */
	function fromPlain(text: string): string {
		return text
			.split(/\n{2,}/)
			.map((p) => p.trim())
			.filter(Boolean)
			.map((p) => `<p>${escapeHtml(p).replace(/\n/g, '<br>')}</p>`)
			.join('');
	}

	onMount(() => {
		let destroyed = false;
		(async () => {
			try {
				// Loaded on demand: pages without an editor never download it.
				const [{ Editor }, { default: StarterKit }] = await Promise.all([
					import('@tiptap/core'),
					import('@tiptap/starter-kit')
				]);
				if (destroyed || !host) return;
				editor = new Editor({
					element: host,
					content: value,
					editable: !disabled,
					// Exactly what the server keeps: paragraphs, line breaks, bold, italic, lists, quotes, links.
					extensions: [
						StarterKit.configure({
							heading: false,
							code: false,
							codeBlock: false,
							strike: false,
							underline: false,
							horizontalRule: false,
							dropcursor: false,
							gapcursor: false,
							trailingNode: false,
							link: {
								openOnClick: false,
								autolink: false,
								linkOnPaste: false,
								protocols: ['http', 'https'],
								isAllowedUri: (url: string) => safeHttpUrl(url) !== null,
								HTMLAttributes: { rel: 'noopener noreferrer nofollow', target: '_blank' }
							}
						})
					],
					editorProps: {
						attributes: {
							role: 'textbox',
							'aria-multiline': 'true',
							'aria-label': label,
							...(describedby ? { 'aria-describedby': describedby } : {}),
							...(lang ? { lang } : {}),
							class: 'rte-content min-h-[9rem] px-3.5 py-3 text-sm leading-relaxed text-foreground focus:outline-none'
						}
					},
					onUpdate: ({ editor: e }) => {
						onchange(e.isEmpty ? '' : e.getHTML());
						sync();
					},
					onSelectionUpdate: sync,
					onTransaction: sync
				});
			} catch {
				failed = true;
				fallbackText = '';
			}
		})();
		return () => {
			destroyed = true;
			editor?.destroy();
		};
	});

	$effect(() => {
		editor?.setEditable(!disabled);
	});

	function openLink() {
		if (!editor) return;
		if (active.link) {
			editor.chain().focus().extendMarkRange('link').unsetLink().run();
			return;
		}
		linkValue = '';
		linkError = '';
		linkOpen = true;
	}

	function applyLink() {
		const href = safeHttpUrl(linkValue);
		if (!href) {
			linkError = 'Enter a full link starting with http:// or https://.';
			return;
		}
		editor?.chain().focus().extendMarkRange('link').setLink({ href }).run();
		linkOpen = false;
	}

	const tools = $derived([
		{ id: 'bold', label: 'Bold', text: 'B', cls: 'font-bold', on: active.bold, run: () => editor?.chain().focus().toggleBold().run() },
		{ id: 'italic', label: 'Italic', text: 'I', cls: 'italic', on: active.italic, run: () => editor?.chain().focus().toggleItalic().run() },
		{ id: 'link', label: active.link ? 'Remove link' : 'Add link', text: 'Link', cls: 'underline', on: active.link, run: openLink },
		{ id: 'bullet', label: 'Bulleted list', text: '• List', cls: '', on: active.bullet, run: () => editor?.chain().focus().toggleBulletList().run() },
		{ id: 'ordered', label: 'Numbered list', text: '1. List', cls: '', on: active.ordered, run: () => editor?.chain().focus().toggleOrderedList().run() },
		{ id: 'quote', label: 'Quote', text: '“ Quote', cls: '', on: active.quote, run: () => editor?.chain().focus().toggleBlockquote().run() }
	]);
</script>

<div
	class={cn(
		'rounded-sm border border-input bg-card transition-colors focus-within:border-accent focus-within:ring-2 focus-within:ring-ring/30',
		invalid && 'border-destructive focus-within:border-destructive',
		disabled && 'opacity-60',
		className
	)}
>
	{#if failed}
		<textarea
			class="block min-h-[9rem] w-full resize-y bg-transparent px-3.5 py-3 text-sm leading-relaxed text-foreground focus:outline-none"
			aria-label={label}
			aria-describedby={describedby}
			{lang}
			{disabled}
			bind:value={fallbackText}
			oninput={() => onchange(fromPlain(fallbackText))}
			placeholder="Formatting isn't available right now. You can still write plain text."
		></textarea>
	{:else}
		<div class="flex flex-wrap items-center gap-1 border-b border-border/70 px-2 py-1.5" role="toolbar" aria-label="Formatting">
			{#each tools as tool (tool.id)}
				<button
					type="button"
					class={cn(
						'rounded-sm px-2 py-1 font-mono text-[11px] tracking-wide text-muted-foreground transition hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40 disabled:opacity-40',
						tool.cls,
						tool.on && 'bg-muted text-foreground'
					)}
					aria-pressed={tool.on}
					aria-label={tool.label}
					title={tool.label}
					disabled={disabled || !editor}
					onmousedown={(e) => e.preventDefault()}
					onclick={tool.run}
				>
					{tool.text}
				</button>
			{/each}
		</div>
		{#if linkOpen}
			<div class="flex flex-wrap items-center gap-2 border-b border-border/70 bg-background/60 px-2 py-2">
				<input
					type="url"
					class="min-w-0 flex-1 rounded-sm border border-input bg-card px-2.5 py-1.5 text-sm focus-visible:border-accent focus-visible:outline-none"
					placeholder="https://"
					aria-label="Link address"
					aria-invalid={linkError ? 'true' : undefined}
					bind:value={linkValue}
					onkeydown={(e) => {
						if (e.key === 'Enter') {
							e.preventDefault();
							applyLink();
						} else if (e.key === 'Escape') {
							linkOpen = false;
							editor?.commands.focus();
						}
					}}
				/>
				<button type="button" class="rounded-sm border border-border px-2.5 py-1.5 font-mono text-[11px] uppercase tracking-wider hover:bg-muted" onclick={applyLink}>Apply</button>
				<button type="button" class="px-1.5 py-1.5 font-mono text-[11px] uppercase tracking-wider text-muted-foreground hover:text-foreground" onclick={() => { linkOpen = false; editor?.commands.focus(); }}>Cancel</button>
				{#if linkError}<p class="w-full text-[13px] text-destructive" role="alert">{linkError}</p>{/if}
			</div>
		{/if}
		<div bind:this={host} class="rte">
			{#if !editor}<p class="px-3.5 py-3 text-sm text-muted-foreground">Loading editor…</p>{/if}
		</div>
	{/if}
</div>

<style>
	/* ProseMirror content lives outside Svelte's scoping, so target it globally. */
	.rte :global(.rte-content p) {
		margin: 0.6rem 0;
	}
	.rte :global(.rte-content > p:first-child) {
		margin-top: 0;
	}
	.rte :global(.rte-content > p:last-child) {
		margin-bottom: 0;
	}
	.rte :global(.rte-content ul) {
		list-style: disc;
		padding-left: 1.25rem;
	}
	.rte :global(.rte-content ol) {
		list-style: decimal;
		padding-left: 1.25rem;
	}
	.rte :global(.rte-content blockquote) {
		border-left: 2px solid var(--border);
		padding-left: 0.9rem;
		color: var(--muted-foreground);
	}
	.rte :global(.rte-content a) {
		text-decoration: underline;
		text-underline-offset: 3px;
	}
</style>
