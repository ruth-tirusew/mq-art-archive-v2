<script lang="ts">
	import type { HTMLTextareaAttributes } from 'svelte/elements';
	import { cn } from '../cn';
	import { fieldClass } from './fieldClass';

	type Props = HTMLTextareaAttributes & { invalid?: boolean; class?: string };

	let { value = $bindable(), invalid = false, rows = 5, class: className, ...rest }: Props = $props();
</script>

<textarea bind:value {rows} aria-invalid={invalid || undefined} class={cn(fieldClass, 'resize-y leading-relaxed', className)} {...rest}></textarea>
