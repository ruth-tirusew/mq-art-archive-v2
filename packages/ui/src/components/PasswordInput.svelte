<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { cn } from '../cn';
	import { fieldClass } from './fieldClass';

	type Props = Omit<HTMLInputAttributes, 'type'> & {
		invalid?: boolean;
		/** Classes for the input itself, as with Input. */
		class?: string;
		/** Classes for the box around the input and its eye button (spacing, width). */
		wrapperClass?: string;
	};

	let { value = $bindable(), invalid = false, class: className, wrapperClass, ...rest }: Props = $props();

	let visible = $state(false);
	const label = $derived(visible ? 'Hide password' : 'Show password');
</script>

<div class={cn('relative', wrapperClass)}>
	<input
		bind:value
		type={visible ? 'text' : 'password'}
		aria-invalid={invalid || undefined}
		class={cn(fieldClass, className, 'pr-11')}
		{...rest}
	/>
	<button
		type="button"
		class="absolute right-1 top-1/2 inline-flex h-8 w-8 -translate-y-1/2 items-center justify-center rounded-sm text-muted-foreground transition hover:text-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/40 disabled:pointer-events-none disabled:opacity-40"
		aria-label={label}
		aria-pressed={visible}
		title={label}
		disabled={rest.disabled}
		onmousedown={(e) => e.preventDefault()}
		onclick={() => (visible = !visible)}
	>
		<svg viewBox="0 0 24 24" class="h-[18px] w-[18px]" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">
			{#if visible}
				<!-- Eye with a slash: the password is showing, press to hide it. -->
				<path d="M10.733 5.076a10.744 10.744 0 0 1 11.205 6.575 1 1 0 0 1 0 .696 10.747 10.747 0 0 1-1.444 2.49" />
				<path d="M14.084 14.158a3 3 0 0 1-4.242-4.242" />
				<path d="M17.479 17.499a10.75 10.75 0 0 1-15.417-5.151 1 1 0 0 1 0-.696 10.75 10.75 0 0 1 4.446-5.143" />
				<path d="m2 2 20 20" />
			{:else}
				<path d="M2.062 12.348a1 1 0 0 1 0-.696 10.75 10.75 0 0 1 19.876 0 1 1 0 0 1 0 .696 10.75 10.75 0 0 1-19.876 0" />
				<circle cx="12" cy="12" r="3" />
			{/if}
		</svg>
	</button>
</div>
