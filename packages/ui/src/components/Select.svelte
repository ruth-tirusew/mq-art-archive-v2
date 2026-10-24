<script lang="ts">
	import type { HTMLSelectAttributes } from 'svelte/elements';
	import { cn } from '../cn';
	import { fieldClass } from './fieldClass';

	type Props = HTMLSelectAttributes & { invalid?: boolean; class?: string };

	let { value = $bindable(), invalid = false, class: className, children, ...rest }: Props = $props();
</script>

<select bind:value aria-invalid={invalid || undefined} class={cn(fieldClass, 'appearance-none pr-9', className)} {...rest}>
	{@render children?.()}
</select>
