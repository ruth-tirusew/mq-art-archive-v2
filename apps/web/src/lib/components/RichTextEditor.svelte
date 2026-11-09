<script lang="ts">
	import { onMount } from 'svelte';
	import { cn } from '$lib/utils';
	import { plainToHtml, safeHttpUrl } from '$lib/utils/answers';

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
		/** A short answer: a one- or two-line box instead of a tall one. Same toolbar. */
		compact?: boolean;
		class?: string;
	};

	let { value, label, onchange, disabled = false, invalid = false, describedby, lang, compact = false, class: className }: Props = $props();

	type EditorLike = import('@tiptap/core').Editor;
	type BlockStyle = 'paragraph' | 'h2' | 'h3';

	let host = $state<HTMLDivElement>();
	let editor = $state.raw<EditorLike>();
	let failed = $state(false);
	let active = $state({
		bold: false,
		italic: false,
		underline: false,
		strike: false,
		link: false,
		bullet: false,
		ordered: false,
		quote: false,
		canUndo: false,
		canRedo: false,
		block: 'paragraph' as BlockStyle
	});
	let linkOpen = $state(false);
	let linkValue = $state('');
	let linkError = $state('');
	let fallbackText = $state('');

	function sync() {
		if (!editor) return;
		active = {
			bold: editor.isActive('bold'),
			italic: editor.isActive('italic'),
			underline: editor.isActive('underline'),
			strike: editor.isActive('strike'),
			link: editor.isActive('link'),
			bullet: editor.isActive('bulletList'),
			ordered: editor.isActive('orderedList'),
			quote: editor.isActive('blockquote'),
			canUndo: editor.can().undo(),
			canRedo: editor.can().redo(),
			block: editor.isActive('heading', { level: 2 })
				? 'h2'
				: editor.isActive('heading', { level: 3 })
					? 'h3'
					: 'paragraph'
		};
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
					// Exactly what the server keeps (backend richtext sanitizer): paragraphs, line breaks,
					// h2/h3, bold, italic, underline, strikethrough, lists, quotes, links.
					extensions: [
						StarterKit.configure({
							heading: { levels: [2, 3] },
							code: false,
							codeBlock: false,
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
							class: `rte-content ${compact ? 'min-h-[2.75rem]' : 'min-h-[9rem]'} px-3.5 py-3 text-sm leading-relaxed text-foreground focus:outline-none`
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

	function setBlock(style: BlockStyle) {
		const chain = editor?.chain().focus();
		if (!chain) return;
		if (style === 'paragraph') chain.setParagraph().run();
		else chain.setHeading({ level: style === 'h2' ? 2 : 3 }).run();
	}

	type Tool = { id: string; label: string; icon: string; on?: boolean; off?: boolean; run: () => void };

	// Grouped like Tiptap's editor toolbar: history | inline marks | blocks | link.
	const groups = $derived<Tool[][]>([
		[
			{ id: 'undo', label: 'Undo', icon: 'undo', off: !active.canUndo, run: () => editor?.chain().focus().undo().run() },
			{ id: 'redo', label: 'Redo', icon: 'redo', off: !active.canRedo, run: () => editor?.chain().focus().redo().run() }
		],
		[
			{ id: 'bold', label: 'Bold', icon: 'bold', on: active.bold, run: () => editor?.chain().focus().toggleBold().run() },
			{ id: 'italic', label: 'Italic', icon: 'italic', on: active.italic, run: () => editor?.chain().focus().toggleItalic().run() },
			{ id: 'underline', label: 'Underline', icon: 'underline', on: active.underline, run: () => editor?.chain().focus().toggleUnderline().run() },
			{ id: 'strike', label: 'Strikethrough', icon: 'strike', on: active.strike, run: () => editor?.chain().focus().toggleStrike().run() }
		],
		[
			{ id: 'bullet', label: 'Bulleted list', icon: 'list', on: active.bullet, run: () => editor?.chain().focus().toggleBulletList().run() },
			{ id: 'ordered', label: 'Numbered list', icon: 'ordered', on: active.ordered, run: () => editor?.chain().focus().toggleOrderedList().run() },
			{ id: 'quote', label: 'Quote', icon: 'quote', on: active.quote, run: () => editor?.chain().focus().toggleBlockquote().run() }
		],
		[
			{
				id: 'link',
				label: active.link ? 'Remove link' : 'Add link',
				icon: active.link ? 'unlink' : 'link',
				on: active.link,
				run: openLink
			}
		]
	]);

	// Lucide icon paths (ISC licence), inlined so the editor adds no icon dependency.
	const ICONS: Record<string, string[]> = {
		undo: ['M9 14 4 9l5-5', 'M4 9h10.5a5.5 5.5 0 0 1 5.5 5.5a5.5 5.5 0 0 1-5.5 5.5H11'],
		redo: ['m15 14 5-5-5-5', 'M20 9H9.5A5.5 5.5 0 0 0 4 14.5A5.5 5.5 0 0 0 9.5 20H13'],
		bold: ['M6 12h9a4 4 0 0 1 0 8H7a1 1 0 0 1-1-1V5a1 1 0 0 1 1-1h7a4 4 0 0 1 0 8'],
		italic: ['M19 4h-9', 'M14 20H5', 'M15 4 9 20'],
		underline: ['M6 4v6a6 6 0 0 0 12 0V4', 'M4 20h16'],
		strike: ['M16 4H9a3 3 0 0 0-2.83 4', 'M14 12a4 4 0 0 1 0 8H6', 'M4 12h16'],
		list: ['M3 6h.01', 'M3 12h.01', 'M3 18h.01', 'M8 6h13', 'M8 12h13', 'M8 18h13'],
		ordered: ['M10 6h11', 'M10 12h11', 'M10 18h11', 'M4 6h1v4', 'M4 10h2', 'M6 18H4c0-1 2-2 2-3s-1-1.5-2-1'],
		quote: [
			'M16 3a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2 1 1 0 0 1 1 1v1a2 2 0 0 1-2 2 1 1 0 0 0-1 1v2a1 1 0 0 0 1 1 6 6 0 0 0 6-6V5a2 2 0 0 0-2-2z',
			'M5 3a2 2 0 0 0-2 2v6a2 2 0 0 0 2 2 1 1 0 0 1 1 1v1a2 2 0 0 1-2 2 1 1 0 0 0-1 1v2a1 1 0 0 0 1 1 6 6 0 0 0 6-6V5a2 2 0 0 0-2-2z'
		],
		link: [
			'M10 13a5 5 0 0 0 7.54.54l3-3a5 5 0 0 0-7.07-7.07l-1.72 1.71',
			'M14 11a5 5 0 0 0-7.54-.54l-3 3a5 5 0 0 0 7.07 7.07l1.71-1.71'
		],
		unlink: [
			'm18.84 12.25 1.72-1.71a5 5 0 0 0-7.07-7.07l-1.72 1.71',
			'm5.17 11.75-1.71 1.71a5 5 0 0 0 7.07 7.07l1.71-1.71',
			'M8 2v3',
			'M2 8h3',
			'M16 19v3',
			'M19 16h3'
		]
	};
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
			class="block {compact ? 'min-h-[2.75rem]' : 'min-h-[9rem]'} w-full resize-y bg-transparent px-3.5 py-3 text-sm leading-relaxed text-foreground focus:outline-none"
			aria-label={label}
			aria-describedby={describedby}
			{lang}
			{disabled}
			bind:value={fallbackText}
			oninput={() => onchange(plainToHtml(fallbackText))}
			placeholder="Formatting isn't available right now. You can still write plain text."
		></textarea>
	{:else}
		<div class="flex flex-wrap items-center gap-0.5 border-b border-border/70 px-1.5 py-1" role="toolbar" aria-label="Formatting">
			<select
				class="h-8 rounded-sm bg-transparent px-1.5 text-[13px] text-foreground transition hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40 disabled:opacity-40"
				aria-label="Text style"
				title="Text style"
				disabled={disabled || !editor}
				value={active.block}
				onchange={(e) => setBlock(e.currentTarget.value as BlockStyle)}
			>
				<option value="paragraph">Paragraph</option>
				<option value="h2">Heading</option>
				<option value="h3">Subheading</option>
			</select>
			{#each groups as group, gi (gi)}
				<span class="mx-1 h-5 w-px bg-border" aria-hidden="true"></span>
				{#each group as tool (tool.id)}
					<button
						type="button"
						class={cn(
							'inline-flex h-8 w-8 items-center justify-center rounded-sm text-muted-foreground transition hover:bg-muted hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40 disabled:pointer-events-none disabled:opacity-35',
							tool.on && 'bg-muted text-foreground'
						)}
						aria-pressed={tool.on}
						aria-label={tool.label}
						title={tool.label}
						disabled={disabled || !editor || tool.off}
						onmousedown={(e) => e.preventDefault()}
						onclick={tool.run}
					>
						<svg
							viewBox="0 0 24 24"
							class="h-4 w-4"
							fill="none"
							stroke="currentColor"
							stroke-width="2"
							stroke-linecap="round"
							stroke-linejoin="round"
							aria-hidden="true"
						>
							{#each ICONS[tool.icon] as d (d)}<path {d} />{/each}
						</svg>
					</button>
				{/each}
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
	.rte :global(.rte-content h2) {
		margin: 1rem 0 0.4rem;
		font-family: var(--font-display, inherit);
		font-size: 1.3em;
		line-height: 1.3;
	}
	.rte :global(.rte-content h3) {
		margin: 0.85rem 0 0.3rem;
		font-size: 1.08em;
		font-weight: 600;
	}
	.rte :global(.rte-content > :is(h2, h3):first-child) {
		margin-top: 0;
	}
	.rte :global(.rte-content s) {
		text-decoration: line-through;
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
