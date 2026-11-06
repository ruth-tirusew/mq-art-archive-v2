<script lang="ts">
	import { Button, PageHeader } from '$ui';

	export type InviteProblemKind = 'invalid' | 'claimed' | 'expired' | 'revoked' | 'unreachable';

	type Props = {
		kind: InviteProblemKind;
		/** Where "sign in" leads, keeping the invite in the return path. */
		signInHref: string;
		/** Where a regular application (no invitation) starts. */
		applyHref: string;
		onretry: () => void;
		retrying?: boolean;
	};

	let { kind, signInHref, applyHref, onretry, retrying = false }: Props = $props();

	const copy: Record<InviteProblemKind, { title: string; body: string }> = {
		invalid: {
			title: 'We couldn’t find that invitation',
			body: 'The link may have been copied incompletely. Open it again from your email, or ask the Artiv team to send a new one.'
		},
		claimed: {
			title: 'This invitation has already been used',
			body: 'If that was you, sign in to pick up where you left off. If not, ask the Artiv team for a new invitation.'
		},
		expired: {
			title: 'This invitation has expired',
			body: 'Invitations only last a couple of weeks. Ask the Artiv team for a new one, or apply directly.'
		},
		revoked: {
			title: 'This invitation was cancelled',
			body: 'Ask the Artiv team for a new invitation, or apply directly.'
		},
		unreachable: {
			title: 'We couldn’t check your invitation',
			body: 'This is usually a connection problem on our side or yours, and nothing is wrong with your link. Try again in a moment.'
		}
	};
</script>

<section class="mx-auto max-w-xl px-6 py-16 md:px-10 md:py-24" data-testid="invite-problem" data-kind={kind}>
	<PageHeader eyebrow="Invitation">{copy[kind].title}</PageHeader>
	<p class="mt-6 text-sm leading-relaxed text-muted-foreground">{copy[kind].body}</p>
	<div class="mt-8 flex flex-wrap items-center gap-4">
		{#if kind === 'unreachable'}
			<Button loading={retrying} onclick={onretry}>Try again</Button>
		{/if}
		<Button variant={kind === 'unreachable' ? 'secondary' : 'primary'} href={signInHref}>Sign in</Button>
		{#if kind !== 'unreachable'}
			<a href={applyHref} class="text-sm text-accent underline underline-offset-4">Apply without an invitation</a>
		{/if}
	</div>
</section>
