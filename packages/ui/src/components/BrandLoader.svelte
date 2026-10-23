<script lang="ts">
	import { onMount } from 'svelte';

	type Props = {
		/** Shortest time the splash stays up, so it never flashes by. */
		minMs?: number;
		/** Give up waiting after this long and reveal the page anyway. */
		maxMs?: number;
	};

	let { minMs = 1300, maxMs = 6000 }: Props = $props();

	// Rendered on the server, so it is on screen before any JS runs. It hides itself once the
	// page has hydrated and the web fonts are ready (so the wordmark never swaps mid-animation).
	let leaving = $state(false);
	let gone = $state(false);

	onMount(() => {
		const started = performance.now();
		let finished = false;

		const finish = () => {
			if (finished) return;
			finished = true;
			const wait = Math.max(0, minMs - (performance.now() - started));
			setTimeout(() => {
				leaving = true;
				setTimeout(() => (gone = true), 600);
			}, wait);
		};

		const fonts = document.fonts?.ready ?? Promise.resolve();
		void fonts.then(finish);
		const cap = setTimeout(finish, maxMs);
		return () => clearTimeout(cap);
	});
</script>

{#if !gone}
	<div class="loader" class:leaving role="status" aria-live="polite" aria-label="Loading Artiv">
		<p class="word" aria-hidden="true">
			<span class="ch" style="--i:0">a</span><span class="ch" style="--i:1">r</span><span class="ch" style="--i:2">t</span><span
				class="ch i"
				style="--i:3"
				>ı<span class="dot" class:settled={leaving}></span></span
			><span class="ch" style="--i:4">v</span><span class="ch period" style="--i:5">.</span>
		</p>
	</div>
{/if}

<style>
	.loader {
		position: fixed;
		inset: 0;
		z-index: 100;
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--background);
		color: var(--foreground);
		transition: opacity 0.55s ease;
		/* If scripts never load, still get out of the way. */
		animation: loader-failsafe 0.5s ease 8s forwards;
	}
	.loader.leaving {
		opacity: 0;
		pointer-events: none;
	}

	.word {
		margin: 0;
		font-family: var(--font-display);
		font-weight: 500;
		font-size: clamp(3.5rem, 12vw, 7rem);
		line-height: 1;
		letter-spacing: -0.03em;
		user-select: none;
	}

	.ch {
		display: inline-block;
		opacity: 0;
		transform: translateY(0.18em);
		animation: letter-in 0.7s cubic-bezier(0.22, 1, 0.36, 1) forwards;
		animation-delay: calc(var(--i) * 90ms + 120ms);
	}
	.period {
		color: var(--accent);
	}

	/* The dotless "ı" is the stem; this circle is its dot. Sized and placed in em so it tracks the
	   glyph at any font size, and positioned inside the letter's own box so a late font swap
	   moves them together. */
	.i {
		position: relative;
	}
	.dot {
		position: absolute;
		left: 50%;
		top: 0.1em;
		width: 0.15em;
		height: 0.15em;
		margin-left: -0.075em;
		border-radius: 50%;
		background: currentColor;
		transform-origin: 50% 100%;
		animation: dot-bounce 1s infinite;
		animation-delay: 0.7s;
	}
	.dot.settled {
		animation: none;
		transform: none;
	}

	@keyframes letter-in {
		to {
			opacity: 1;
			transform: translateY(0);
		}
	}

	/* Fall, squash on impact, spring back up. Gravity easing on the way down, the reverse up. */
	@keyframes dot-bounce {
		0% {
			transform: translateY(-0.62em) scale(1, 1);
			animation-timing-function: cubic-bezier(0.55, 0, 1, 0.45);
		}
		46% {
			transform: translateY(0) scale(1, 1);
		}
		50% {
			transform: translateY(0) scale(1.35, 0.62);
			animation-timing-function: cubic-bezier(0, 0.55, 0.45, 1);
		}
		56% {
			transform: translateY(0) scale(0.92, 1.1);
		}
		100% {
			transform: translateY(-0.62em) scale(1, 1);
		}
	}

	@keyframes loader-failsafe {
		to {
			opacity: 0;
			visibility: hidden;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.ch {
			animation: letter-fade 0.4s ease forwards;
			transform: none;
		}
		.dot {
			animation: none;
		}
		@keyframes letter-fade {
			to {
				opacity: 1;
			}
		}
	}
</style>
