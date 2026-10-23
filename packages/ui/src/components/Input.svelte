<script lang="ts">
	import type { HTMLInputAttributes } from 'svelte/elements';
	import { cn } from '../cn';
	import { fieldClass } from './fieldClass';

	type Props = HTMLInputAttributes & { invalid?: boolean; class?: string };

	let { value = $bindable(), invalid = false, class: className, ...rest }: Props = $props();
</script>

<input bind:value aria-invalid={invalid || undefined} class={cn(fieldClass, className)} {...rest} />
