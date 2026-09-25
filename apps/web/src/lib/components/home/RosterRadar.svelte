<script lang="ts">
	import type { ArtistProfile } from '$lib/core/domain/profile';
	import {
		artistDiscipline,
		artistName,
		artistSlug
	} from '$lib/utils/fields';

	type Props = {
		artists: ArtistProfile[];
	};

	let { artists }: Props = $props();

	const size = 500;
	const center = size / 2;

	const maxRadius = 200;
	const innerRadius = 64;
	const ringCount = 5;

	const rings = Array.from(
		{ length: ringCount },
		(_, index) => ((index + 1) / ringCount) * maxRadius
	);

	function seededUnit(seed: string): number {
		let hash = 0x811c9dc5;

		for (let i = 0; i < seed.length; i++) {
			hash ^= seed.charCodeAt(i);
			hash = Math.imul(hash, 0x01000193);
		}

		return ((hash >>> 0) % 10000) / 10000;
	}

	function primaryDiscipline(artist: ArtistProfile): string {
		const raw = artistDiscipline(artist) ?? '';
		const first = raw.split(/[/,]/)[0]?.trim();

		return first || 'Mixed Media';
	}

	const categories = $derived.by(() => {
		const seen: string[] = [];

		for (const artist of artists) {
			const discipline = primaryDiscipline(artist);

			if (!seen.includes(discipline)) {
				seen.push(discipline);
			}
		}

		return seen;
	});

	function axisAngle(index: number, count: number): number {
		return -90 + (index * 360) / count;
	}

	function polarPoint(angleDeg: number, radius: number) {
		const angle = (angleDeg * Math.PI) / 180;

		return {
			x: center + Math.cos(angle) * radius,
			y: center + Math.sin(angle) * radius
		};
	}

	type Node = {
		x: number;
		y: number;
		r: number;
		name: string;
		discipline: string;
		href: string;
	};

	type CategoryLabel = {
		name: string;
		angle: number;
		x: number;
		y: number;
		anchor: 'start' | 'middle' | 'end';
	};

	const categoryLabels = $derived.by((): CategoryLabel[] => {
		if (!categories.length) return [];

		return categories.map((category, index) => {
			const angle = axisAngle(index, categories.length);

			const point = polarPoint(angle, maxRadius + 42);

			const radians = (angle * Math.PI) / 180;
			const cos = Math.cos(radians);

			let anchor: 'start' | 'middle' | 'end' = 'middle';

			if (cos > 0.35) anchor = 'start';
			if (cos < -0.35) anchor = 'end';

			return {
				name: category.toUpperCase(),
				angle,
				x: point.x,
				y: point.y,
				anchor
			};
		});
	});

	const nodes = $derived.by((): Node[] => {
		if (!categories.length) return [];

		return artists.map((artist) => {
			const slug = artistSlug(artist);
			const discipline = primaryDiscipline(artist);

			const categoryIndex = categories.indexOf(discipline);
			const categoryAngle = axisAngle(categoryIndex, categories.length);

			const sector = 360 / categories.length;
			const spread = sector / 2.6;

			const jitter =
				(seededUnit(`${slug}:angle`) - 0.5) * 2 * spread;

			const angle = categoryAngle + jitter;

			const radius =
				innerRadius +
				seededUnit(`${slug}:radius`) *
					(maxRadius - innerRadius - 35);

			const point = polarPoint(angle, radius);

			return {
				x: point.x,
				y: point.y,
				r: 4 + seededUnit(`${slug}:size`) * 3,
				name: artistName(artist),
				discipline,
				href: `/artists/${slug}`
			};
		});
	});

	const ambientDots = $derived.by(() => {
		const count = Math.min(28, artists.length * 4);

		return Array.from({ length: count }, (_, index) => {
			const seed = `ambient:${index}`;

			const angle =
				seededUnit(`${seed}:angle`) * Math.PI * 2;

			const radius =
				30 +
				seededUnit(`${seed}:radius`) *
					(maxRadius - 30);

			return {
				x: center + Math.cos(angle) * radius,
				y: center + Math.sin(angle) * radius,
				r: 1 + seededUnit(`${seed}:size`) * 1.5
			};
		});
	});

	const sweepWidth = 38;

	const sweepPath = (() => {
		const start = -90;
		const end = start + sweepWidth;

		const startPoint = polarPoint(start, maxRadius);
		const endPoint = polarPoint(end, maxRadius);

		return `
			M ${center} ${center}
			L ${startPoint.x} ${startPoint.y}
			A ${maxRadius} ${maxRadius} 0 0 1
			${endPoint.x} ${endPoint.y}
			Z
		`;
	})();

	const compassGlyph =
		'M0,-14 L3,-3 L14,0 L3,3 L0,14 L-3,3 L-14,0 L-3,-3 Z';
</script>

<figure class="mx-auto w-full">
	<div class="relative">
		<svg
			viewBox={`0 0 ${size} ${size}`}
			class="h-auto w-full overflow-visible"
			role="img"
			aria-label="Radar visualization of the Artiv artist roster"
		>
			<defs>
				<radialGradient id="radarSweep">
					<stop
						offset="0%"
						style="stop-color: var(--color-accent); stop-opacity: 0.16"
					/>
					<stop
						offset="100%"
						style="stop-color: var(--color-accent); stop-opacity: 0"
					/>
				</radialGradient>

				<radialGradient id="radarHalo">
					<stop
						offset="0%"
						style="stop-color: var(--color-accent); stop-opacity: 0.24"
					/>
					<stop
						offset="100%"
						style="stop-color: var(--color-accent); stop-opacity: 0"
					/>
				</radialGradient>
			</defs>

			<!-- Radar rings -->
			{#each rings as radius}
				<circle
					cx={center}
					cy={center}
					r={radius}
					class="fill-none stroke-border/70"
					stroke-width="1"
				/>
			{/each}

			<!-- Subtle crosshair -->
			<line
				x1={center}
				y1={center - maxRadius}
				x2={center}
				y2={center + maxRadius}
				class="stroke-border/50"
				stroke-width="1"
			/>

			<line
				x1={center - maxRadius}
				y1={center}
				x2={center + maxRadius}
				y2={center}
				class="stroke-border/50"
				stroke-width="1"
			/>

			<!-- Animated sweep -->
			<g
				class="radar-sweep"
				style={`transform-origin: ${center}px ${center}px;`}
			>
				<path
					d={sweepPath}
					fill="url(#radarSweep)"
				/>
			</g>

			<!-- Discipline labels -->
			{#each categoryLabels as category}
				{@const outer = polarPoint(category.angle, maxRadius)}
				{@const label = polarPoint(category.angle, maxRadius + 48)}

				<line
					x1={outer.x}
					y1={outer.y}
					x2={label.x}
					y2={label.y}
					class="stroke-accent/50"
					stroke-width="1"
				/>

				<circle
					cx={outer.x}
					cy={outer.y}
					r="2.5"
					class="fill-accent"
				/>

				<text
					x={category.x}
					y={category.y}
					text-anchor={category.anchor}
					dominant-baseline="middle"
					class="fill-muted-foreground font-mono text-[10px] tracking-[0.18em]"
				>
					{category.name}
				</text>
			{/each}

			<!-- Decorative dots -->
			{#each ambientDots as dot}
				<circle
					cx={dot.x}
					cy={dot.y}
					r={dot.r}
					class="fill-muted-foreground/25"
				/>
			{/each}

			<!-- Real artists -->
			{#each nodes as node (node.href)}
				<a
					href={node.href}
					aria-label={`${node.name} — ${node.discipline}`}
					class="group"
				>
					<circle
						cx={node.x}
						cy={node.y}
						r={node.r + 12}
						fill="url(#radarHalo)"
						class="opacity-70 transition-opacity duration-300 group-hover:opacity-100"
					/>

					<circle
						cx={node.x}
						cy={node.y}
						r={node.r}
						class="fill-accent transition-all duration-300 group-hover:r-[8]"
					/>

					<title>
						{node.name} — {node.discipline}
					</title>
				</a>
			{/each}

			<!-- Center hub -->
			<a
				href="/artists"
				aria-label="Explore all artists"
				class="group"
			>
				<circle
					cx={center}
					cy={center}
					r="68"
					fill="url(#radarHalo)"
				/>

				<circle
					cx={center}
					cy={center}
					r="49"
					class="fill-none stroke-accent/40 transition-all duration-300 group-hover:stroke-accent"
					stroke-width="1"
				/>

				<circle
					cx={center}
					cy={center}
					r="40"
					class="fill-accent"
				/>

				<path
					d={compassGlyph}
					class="fill-accent-foreground"
					transform={`translate(${center} ${center})`}
				/>

				<text
					x={center}
					y={center + 62}
					text-anchor="middle"
					class="fill-muted-foreground font-mono text-[9px] tracking-[0.22em]"
				>
					DISCOVER
				</text>

				<text
					x={center}
					y={center + 76}
					text-anchor="middle"
					class="fill-muted-foreground font-mono text-[9px] tracking-[0.22em]"
				>
					ARTISTS
				</text>
			</a>
		</svg>
	</div>

	<figcaption class="sr-only">
		A radar visualization of the Artiv artist roster. Each red point
		represents a real artist and links to their profile. The center
		links to the full artist directory.
	</figcaption>
</figure>

<style>
	.radar-sweep {
		transform-origin: center;
		animation: radar-spin 32s linear infinite;
	}

	@keyframes radar-spin {
		to {
			transform: rotate(360deg);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.radar-sweep {
			animation: none;
		}
	}
</style>