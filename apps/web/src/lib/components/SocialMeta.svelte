<script lang="ts">
	import { shareImage } from '$lib/utils/share';

	type Props = {
		/** The page title as shown in the preview card. */
		title: string;
		description?: string;
		/** The full address of this page: previews link back to it. */
		url: string;
		/** A picture for the card. Ignored unless it is a real http(s) address. */
		image?: string | null;
		type?: 'profile' | 'website' | 'article';
	};

	let { title, description = '', url, image, type = 'profile' }: Props = $props();

	const picture = $derived(shareImage(image));
</script>

<!-- Link previews (Telegram, WhatsApp, Instagram, Facebook, X, Slack) read these from the page the server sends,
     which is why they live here and are rendered on the server. -->
<svelte:head>
	<link rel="canonical" href={url} />
	<meta property="og:site_name" content="Artiv" />
	<meta property="og:type" content={type} />
	<meta property="og:title" content={title} />
	<meta property="og:url" content={url} />
	{#if description}<meta property="og:description" content={description} />{/if}
	{#if picture}<meta property="og:image" content={picture} />{/if}
	<meta name="twitter:card" content={picture ? 'summary_large_image' : 'summary'} />
	<meta name="twitter:title" content={title} />
	{#if description}<meta name="twitter:description" content={description} />{/if}
	{#if picture}<meta name="twitter:image" content={picture} />{/if}
</svelte:head>
