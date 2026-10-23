<script lang="ts">
	import {
		Alert,
		Badge,
		Button,
		Card,
		ConfirmDialog,
		EmptyState,
		ErrorState,
		Field,
		Input,
		PageHeader,
		Select,
		Skeleton,
		Textarea,
		toast
	} from '$ui';

	let dialogOpen = $state(false);
	let busy = $state(false);
	let name = $state('');

	const swatches = [
		['background', 'bg-background'],
		['card', 'bg-card'],
		['muted', 'bg-muted'],
		['foreground', 'bg-foreground'],
		['ink', 'bg-ink'],
		['accent', 'bg-accent'],
		['ochre', 'bg-ochre'],
		['success', 'bg-success'],
		['warning', 'bg-warning'],
		['destructive', 'bg-destructive']
	];

	async function confirm() {
		busy = true;
		await new Promise((r) => setTimeout(r, 900));
		busy = false;
		dialogOpen = false;
		toast.success('Article unpublished.');
	}
</script>

<svelte:head>
	<title>Design system — Artiv</title>
	<meta name="robots" content="noindex" />
</svelte:head>

<div class="container-page flex flex-col gap-20 py-16 md:py-24">
	<PageHeader eyebrow="Design system" lead="Every inner page is built from these. If a page needs something that isn’t here, add it here first.">
		One kit, <em class="italic">held in common.</em>
	</PageHeader>

	<section class="flex flex-col gap-6">
		<p class="t-eyebrow text-accent">Colour</p>
		<div class="grid grid-cols-2 gap-4 md:grid-cols-5">
			{#each swatches as [label, cls]}
				<div class="flex flex-col gap-2">
					<div class="h-16 rounded-sm border border-border {cls}"></div>
					<p class="t-meta text-muted-foreground">{label}</p>
				</div>
			{/each}
		</div>
	</section>

	<section class="flex flex-col gap-6">
		<p class="t-eyebrow text-accent">Type</p>
		<div class="flex flex-col gap-6 border-t border-border pt-6">
			<p class="t-display">Display <em class="italic">statement</em></p>
			<p class="t-h1">Page title <em class="italic">emphasis</em></p>
			<p class="t-h2">Section heading</p>
			<p class="t-h3">Item title inside a section</p>
			<p class="t-lead text-muted-foreground">Lead copy sits under a heading and is capped at a readable measure, so long sentences never stretch across the whole page.</p>
			<p class="t-body">Body text for reading. Inter at 15px, generous line height.</p>
			<p class="t-eyebrow text-accent">Eyebrow · kicker label</p>
			<p class="t-meta text-muted-foreground">Meta · 2024 · oil on canvas · 90 × 120 cm</p>
		</div>
	</section>

	<section class="flex flex-col gap-6">
		<p class="t-eyebrow text-accent">Buttons</p>
		<div class="flex flex-wrap items-center gap-3">
			<Button>Primary</Button>
			<Button variant="accent">Accent</Button>
			<Button variant="secondary">Secondary</Button>
			<Button variant="ghost">Ghost</Button>
			<Button variant="destructive">Destructive</Button>
			<Button variant="link">Link →</Button>
			<Button size="sm">Small</Button>
			<Button loading>Saving</Button>
			<Button disabled>Disabled</Button>
			<Button href="/apply">As link</Button>
		</div>
		<div class="flex flex-wrap items-center gap-3 rounded-sm bg-ink p-6">
			<Button tone="on-dark">Primary</Button>
			<Button tone="on-dark" variant="accent">Accent</Button>
			<Button tone="on-dark" variant="secondary">Secondary</Button>
			<Button tone="on-dark" variant="link">Link →</Button>
		</div>
	</section>

	<section class="grid gap-10 md:grid-cols-2">
		<div class="flex flex-col gap-6">
			<p class="t-eyebrow text-accent">Form fields</p>
			<Field label="Display name" required hint="Shown on your public profile.">
				{#snippet children({ id, describedby, invalid, required })}
					<Input {id} aria-describedby={describedby} {invalid} {required} bind:value={name} placeholder="Selam Tesfaye" />
				{/snippet}
			</Field>
			<Field label="Handle" error="That handle is taken.">
				{#snippet children({ id, describedby, invalid })}
					<Input {id} aria-describedby={describedby} {invalid} value="selam" />
				{/snippet}
			</Field>
			<Field label="Discipline">
				{#snippet children({ id })}
					<Select {id}>
						<option>Painting</option>
						<option>Sculpture</option>
						<option>Photography</option>
					</Select>
				{/snippet}
			</Field>
			<Field label="Statement">
				{#snippet children({ id })}
					<Textarea {id} placeholder="A few lines about your practice" />
				{/snippet}
			</Field>
		</div>

		<div class="flex flex-col gap-6">
			<p class="t-eyebrow text-accent">Status</p>
			<div class="flex flex-wrap gap-2">
				<Badge>Draft</Badge>
				<Badge tone="accent">Featured</Badge>
				<Badge tone="success">Published</Badge>
				<Badge tone="warning">Pending</Badge>
				<Badge tone="destructive">Rejected</Badge>
				<Badge tone="ink">Verified</Badge>
			</div>
			<Alert title="Heads up">Your application is in review.</Alert>
			<Alert tone="success">Profile saved.</Alert>
			<Alert tone="warning" title="Unsaved changes">Leaving now discards your edits.</Alert>
			<Alert tone="error" title="Couldn’t save">The handle is already in use.</Alert>
			<div class="flex flex-wrap gap-3">
				<Button variant="secondary" size="sm" onclick={() => toast.success('Saved.')}>Success toast</Button>
				<Button variant="secondary" size="sm" onclick={() => toast.error('Couldn’t reach the server.')}>Error toast</Button>
				<Button variant="secondary" size="sm" onclick={() => (dialogOpen = true)}>Confirm dialog</Button>
			</div>
		</div>
	</section>

	<section class="flex flex-col gap-6">
		<p class="t-eyebrow text-accent">Surfaces</p>
		<div class="grid gap-6 md:grid-cols-3">
			<Card>
				<p class="t-eyebrow text-accent">Card</p>
				<p class="t-h3 mt-3">Paper surface</p>
			</Card>
			<Card variant="outline">
				<p class="t-eyebrow text-accent">Outline</p>
				<p class="t-h3 mt-3">Hairline only</p>
			</Card>
			<Card variant="ink" class="flex flex-col gap-4">
				<p class="t-eyebrow text-ochre">Ink</p>
				<p class="t-h3">Feature card</p>
				<div><Button tone="on-dark" variant="accent" size="sm">Apply now →</Button></div>
			</Card>
		</div>
	</section>

	<section class="flex flex-col gap-6">
		<p class="t-eyebrow text-accent">States</p>
		<div class="grid gap-6 md:grid-cols-2">
			<EmptyState title="No works match these filters." body="Try removing a filter, or browse the whole archive.">
				{#snippet action()}<Button variant="secondary" size="sm">Clear filters</Button>{/snippet}
			</EmptyState>
			<ErrorState status={503} message="Your changes are safe. Try again in a moment." onretry={() => new Promise((r) => setTimeout(r, 800))} />
			<ErrorState status={404} />
			<div class="flex flex-col gap-3">
				<Skeleton class="h-6 w-1/3" />
				<Skeleton class="h-4 w-full" />
				<Skeleton class="h-4 w-5/6" />
				<Skeleton class="aspect-[4/5] w-full" />
			</div>
		</div>
	</section>
</div>

<ConfirmDialog
	bind:open={dialogOpen}
	{busy}
	destructive
	title="Unpublish this article?"
	description="It disappears from the public wiki immediately. You can publish it again later."
	confirmLabel="Unpublish"
	onconfirm={confirm}
/>
