<script lang="ts">
	import { toast } from './toast.svelte';
	import Alert from './Alert.svelte';
</script>

<div class="pointer-events-none fixed inset-x-0 bottom-4 z-50 flex flex-col items-center gap-2 px-4 sm:items-end sm:px-6">
	{#each toast.items as t (t.id)}
		<div class="pointer-events-auto w-full max-w-sm shadow-lg">
			<Alert tone={t.tone === 'error' ? 'error' : t.tone === 'success' ? 'success' : 'info'} ondismiss={() => toast.dismiss(t.id)} class="bg-background">
				{t.message}
			</Alert>
		</div>
	{/each}
</div>
