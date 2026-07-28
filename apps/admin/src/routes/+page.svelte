<script lang="ts">
  import { onMount } from 'svelte';
  import { artistsService } from '$lib/application/artists';
  import { postsService } from '$lib/application/posts';
  import { onboardingService } from '$lib/application/onboarding';
  import { eventsService } from '$lib/application/events';
  import { getAnalyticsSummary, type AnalyticsSummary } from '$lib/application/analytics';
  import BotanicalMark from '$lib/components/BotanicalMark.svelte';
  import BotanicalAdornment from '$lib/components/BotanicalAdornment.svelte';
  import type { ArtistProfile } from '$lib/core/domain/artist';
  import type { ArtPost } from '$lib/core/domain/art';
  import type { OnboardingApplication } from '$lib/core/domain/onboarding';
  import type { Event } from '$lib/core/domain/event';
  import {
    countByDay,
    countInRange,
    daysAgo,
    lastNDays,
    normalizeMedium,
    pctChange,
    relativeTime,
    sparklinePath,
    startOfDay,
    weekdayShort
  } from '$lib/dashboard/stats';

  let loading = $state(true);
  let errorMessage = $state('');

  let artists = $state<ArtistProfile[]>([]);
  let posts = $state<ArtPost[]>([]);
  let applications = $state<OnboardingApplication[]>([]);
  let events = $state<Event[]>([]);
  let analytics = $state<AnalyticsSummary | null>(null);

  onMount(async () => {
    try {
      const [artistList, postList, apps, evts, viewSummary] = await Promise.all([
        artistsService.list(),
        postsService.list(),
        onboardingService.listPending(),
        eventsService.listPending(),
        getAnalyticsSummary().catch(() => null)
      ]);
      artists = artistList;
      posts = postList;
      applications = apps;
      events = evts;
      analytics = viewSummary;
    } catch (err) {
      errorMessage = err instanceof Error ? err.message : 'Failed to load dashboard';
    } finally {
      loading = false;
    }
  });

  const today = new Date();
  const todayLabel = today.toLocaleDateString('en-US', {
    weekday: 'long',
    month: 'long',
    day: 'numeric',
    year: 'numeric'
  });

  const weekDays = $derived(lastNDays(7, today));

  const artistCounts = $derived.by(() => {
    const draft = artists.filter((a) => a.status === 'draft').length;
    const pending = artists.filter((a) => a.status === 'pending').length;
    const approved = artists.filter((a) => a.status === 'approved').length;
    return { draft, pending, approved, total: artists.length };
  });

  const postCounts = $derived.by(() => {
    const draft = posts.filter((p) => (p.status ?? 'draft') === 'draft').length;
    const published = posts.filter((p) => p.status === 'published').length;
    const archived = posts.filter((p) => p.status === 'archived').length;
    return { draft, published, archived, total: posts.length };
  });

  const pendingReview = $derived(
    applications.length + artistCounts.pending + postCounts.draft + events.length
  );

  const queueCreatedDates = $derived([
    ...applications.map((a) => a.created_at),
    ...artists.filter((a) => a.status === 'pending').map((a) => a.created_at),
    ...posts.filter((p) => (p.status ?? 'draft') === 'draft').map((p) => p.created_at),
    ...events.map((e) => e.created_at)
  ]);

  const pendingDelta = $derived.by(() => {
    const todayStart = startOfDay(today);
    const yesterdayStart = daysAgo(1, today);
    const todayN = countInRange(queueCreatedDates, todayStart, new Date(todayStart.getTime() + 86400000));
    const yestN = countInRange(queueCreatedDates, yesterdayStart, todayStart);
    return todayN - yestN;
  });

  const appsOverTime = $derived(countByDay(applications.map((a) => a.created_at), weekDays));
  /** Prefer applications; fall back to all inbound timestamps so the chart isn't empty. */
  const seriesOverTime = $derived.by(() => {
    const fromApps = appsOverTime;
    if (fromApps.some((n) => n > 0)) return fromApps;
    return countByDay(
      [
        ...applications.map((a) => a.created_at),
        ...artists.map((a) => a.created_at),
        ...posts.map((p) => p.created_at)
      ],
      weekDays
    );
  });

  const lineChart = $derived.by(() => {
    const values = seriesOverTime;
    const w = 560;
    const h = 200;
    const padX = 28;
    const padY = 24;
    const max = Math.max(...values, 1);
    const step = values.length > 1 ? (w - padX * 2) / (values.length - 1) : 0;
    const points = values.map((v, i) => ({
      x: padX + i * step,
      y: padY + (h - padY * 2) * (1 - v / max),
      v,
      day: weekDays[i]
    }));
    const path = points
      .map((p, i) => `${i === 0 ? 'M' : 'L'}${p.x.toFixed(1)},${p.y.toFixed(1)}`)
      .join(' ');
    const area =
      path +
      ` L${points[points.length - 1]?.x ?? padX},${h - padY} L${padX},${h - padY} Z`;
    const peak = points.reduce((best, p) => (p.v >= best.v ? p : best), points[0] ?? { x: 0, y: 0, v: 0, day: today });
    return { w, h, padX, padY, points, path, area, peak, max };
  });

  const funnel = $derived.by(() => {
    const submitted = applications.length + artists.length + posts.length + events.length;
    const underReview = applications.length + artistCounts.pending + postCounts.draft + events.length;
    const approved = artistCounts.approved;
    const published = postCounts.published;
    const rows = [
      { label: 'Submitted', count: submitted, color: 'bg-foreground' },
      { label: 'Under review', count: underReview, color: 'bg-viz-tan' },
      { label: 'Approved', count: approved, color: 'bg-viz-olive' },
      { label: 'Published', count: published, color: 'bg-viz-navy' }
    ];
    const max = Math.max(...rows.map((r) => r.count), 1);
    return rows.map((r) => ({ ...r, width: `${(r.count / max) * 100}%` }));
  });

  const archiveHealth = $derived.by(() => {
    const approved = artistCounts.approved + postCounts.published;
    const pending = artistCounts.pending + postCounts.draft + applications.length + events.length;
    const other = artistCounts.draft + postCounts.archived;
    const total = Math.max(approved + pending + other, 1);
    const segs = [
      { label: 'Approved', count: approved, pct: Math.round((approved / total) * 100), tone: 'viz-olive' },
      { label: 'Pending', count: pending, pct: Math.round((pending / total) * 100), tone: 'viz-tan' },
      { label: 'Other', count: other, pct: Math.round((other / total) * 100), tone: 'viz-terracotta' }
    ];
    const r = 38;
    const c = 2 * Math.PI * r;
    let offset = 0;
    const rings = segs.map((s) => {
      const len = (s.count / total) * c;
      const ring = { ...s, dash: `${len} ${c - len}`, offset: -offset };
      offset += len;
      return ring;
    });
    return { segs: rings, approvedPct: segs[0].pct, total };
  });

  type ActivityTone = 'ok' | 'warn' | 'mute';
  const recentActivity = $derived.by(() => {
    const items: { text: string; when: string; tone: ActivityTone; ts: number }[] = [];
    for (const a of applications) {
      items.push({
        text: `Application from ${a.display_name}`,
        when: relativeTime(a.created_at),
        tone: 'mute',
        ts: a.created_at ? new Date(a.created_at).getTime() : 0
      });
    }
    for (const a of artists.filter((x) => x.status === 'approved').slice(0, 8)) {
      items.push({
        text: `Artist profile approved — ${a.display_name}`,
        when: relativeTime(a.updated_at ?? a.created_at),
        tone: 'ok',
        ts: new Date(a.updated_at ?? a.created_at ?? 0).getTime()
      });
    }
    for (const p of posts.filter((x) => x.status === 'published').slice(0, 8)) {
      items.push({
        text: `Post published — ${p.title}`,
        when: relativeTime(p.published_at ?? p.updated_at),
        tone: 'ok',
        ts: new Date(p.published_at ?? p.updated_at ?? 0).getTime()
      });
    }
    for (const p of posts.filter((x) => (x.status ?? 'draft') === 'draft').slice(0, 4)) {
      items.push({
        text: `Draft awaiting review — ${p.title}`,
        when: relativeTime(p.created_at),
        tone: 'mute',
        ts: new Date(p.created_at ?? 0).getTime()
      });
    }
    for (const e of events.slice(0, 6)) {
      items.push({
        text: `Event pending — ${e.title}`,
        when: relativeTime(e.created_at),
        tone: 'warn',
        ts: new Date(e.created_at ?? 0).getTime()
      });
    }
    return items.sort((a, b) => b.ts - a.ts).slice(0, 6);
  });

  function entitySpark(
    dates: (string | undefined)[],
    total: number
  ): { total: number; change: number | null; path: string; values: number[] } {
    const week = lastNDays(7, today);
    const prevFrom = daysAgo(14, today);
    const prevTo = daysAgo(7, today);
    const thisWeek = countInRange(dates, daysAgo(7, today), new Date(today.getTime() + 86400000));
    const lastWeek = countInRange(dates, prevFrom, prevTo);
    const values = countByDay(dates, week);
    return {
      total,
      change: pctChange(thisWeek, lastWeek),
      path: sparklinePath(values),
      values
    };
  }

  const sparkArtists = $derived(entitySpark(artists.map((a) => a.created_at), artists.length));
  const sparkPosts = $derived(entitySpark(posts.map((p) => p.created_at), posts.length));
  const sparkEvents = $derived(entitySpark(events.map((e) => e.created_at), events.length));

  const contentByType = $derived.by(() => {
    const map = new Map<string, number>();
    for (const p of posts) {
      const key = normalizeMedium(p.medium);
      map.set(key, (map.get(key) ?? 0) + 1);
    }
    const rows = [...map.entries()]
      .map(([label, count]) => ({ label, count }))
      .sort((a, b) => b.count - a.count)
      .slice(0, 5);
    const total = Math.max(
      rows.reduce((s, r) => s + r.count, 0),
      1
    );
    return rows.map((r) => ({
      ...r,
      pct: Math.round((r.count / total) * 100),
      width: `${(r.count / total) * 100}%`
    }));
  });

  const heatmap = $derived.by(() => {
    const hours = [8, 10, 12, 14, 16, 18];
    const dayLabels = ['M', 'T', 'W', 'T', 'F', 'S', 'S'];
    // Align Mon=0 … Sun=6
    const grid: number[][] = Array.from({ length: 7 }, () => hours.map(() => 0));
    const allDates = [
      ...applications.map((a) => a.created_at),
      ...artists.map((a) => a.created_at),
      ...posts.map((p) => p.created_at),
      ...events.map((e) => e.created_at)
    ];
    for (const iso of allDates) {
      if (!iso) continue;
      const d = new Date(iso);
      if (Number.isNaN(d.getTime())) continue;
      const dow = (d.getDay() + 6) % 7; // Mon=0
      const h = d.getHours();
      let hi = hours.findIndex((hour, i) => h >= hour && (i === hours.length - 1 || h < hours[i + 1]));
      if (hi < 0) continue;
      grid[dow][hi] += 1;
    }
    const max = Math.max(...grid.flat(), 1);
    return { hours, dayLabels, grid, max };
  });

  function heatOpacity(n: number, max: number) {
    if (n <= 0) return 0.12;
    return 0.22 + (n / max) * 0.78;
  }

  const recentApprovals = $derived.by(() => {
    const items: {
      title: string;
      meta: string;
      when: string;
      href: string;
      image?: string;
      ts: number;
    }[] = [];

    for (const p of posts.filter((x) => x.status === 'published' && x.published_at)) {
      const approvedAt = p.published_at!;
      items.push({
        title: p.title,
        meta: p.artist_name ? `Artwork by ${p.artist_name}` : 'Artwork',
        when: `Approved ${relativeTime(approvedAt)}`,
        href: `/posts/${p.id}`,
        image: p.media?.[0]?.url,
        ts: new Date(approvedAt).getTime()
      });
    }

    for (const a of artists.filter((x) => x.status === 'approved' && x.approved_at)) {
      items.push({
        title: a.display_name,
        meta: a.discipline ? `Artist · ${a.discipline}` : 'Artist profile',
        when: `Approved ${relativeTime(a.approved_at)}`,
        href: `/artists/${a.id}`,
        image: a.portrait_url || undefined,
        ts: new Date(a.approved_at!).getTime()
      });
    }

    return items.sort((a, b) => b.ts - a.ts).slice(0, 3);
  });

  function toneDot(tone: ActivityTone) {
    if (tone === 'ok') return 'bg-viz-olive';
    if (tone === 'warn') return 'bg-viz-terracotta';
    return 'bg-muted-foreground/40';
  }
</script>

<svelte:head>
  <title>Editorial dashboard — Artiv</title>
</svelte:head>

{#if errorMessage}
  <p class="mb-6 text-sm text-destructive" role="alert">{errorMessage}</p>
{/if}

<!-- Page header -->
<header class="mb-8 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
  <div>
    <p class="text-[11px] font-medium uppercase tracking-[0.22em] text-muted-foreground">
      01 — Operations
    </p>
    <h1 class="mt-2 font-display text-4xl tracking-tight text-foreground md:text-5xl">
      Editorial dashboard
    </h1>
    <p class="mt-2 text-sm text-muted-foreground md:text-base">
      Everything happening inside the archive today.
    </p>
  </div>
  <div class="text-left lg:text-right">
    <p class="text-sm text-foreground">{todayLabel}</p>
    <p class="mt-1 text-sm text-muted-foreground">
      {#if loading}
        Loading queue…
      {:else}
        <span class="font-medium text-foreground">{pendingReview}</span> actions require attention today.
      {/if}
    </p>
  </div>
</header>

<div class="mb-4 grid grid-cols-2 gap-3 sm:grid-cols-4">
  {#each [
    { label: 'Total visits', value: analytics?.total },
    { label: 'Artist visits', value: analytics?.artists },
    { label: 'Post visits', value: analytics?.posts },
    { label: 'Wiki visits', value: analytics?.articles }
  ] as metric}
    <div class="rounded-lg border border-border/70 bg-card px-4 py-3">
      <p class="text-xs text-muted-foreground">{metric.label}</p>
      <p class="mt-1 font-display text-2xl text-foreground">{metric.value?.toLocaleString() ?? '—'}</p>
    </div>
  {/each}
</div>

<!-- Row 1 -->
<div class="grid gap-4 lg:grid-cols-12">
  <!-- Pending review -->
  <article
    class="relative overflow-hidden rounded-xl border border-border/70 bg-card p-6 lg:col-span-4"
  >
    <BotanicalAdornment
      variant="lily"
      class="absolute -left-4 -top-6 w-28 opacity-[0.22] sm:w-32"
    />
    <BotanicalAdornment
      variant="daisies"
      class="absolute -bottom-4 -right-3 w-32 opacity-[0.2] sm:w-36"
    />
    <BotanicalAdornment
      variant="slender"
      class="absolute right-8 top-2 w-14 opacity-[0.12] rotate-12"
    />
    <p class="relative text-sm text-muted-foreground">Pending review</p>
    <p class="relative mt-3 font-display text-6xl leading-none tracking-tight text-foreground">
      {loading ? '—' : pendingReview}
    </p>
    {#if !loading}
      <p
        class="relative mt-3 text-sm {pendingDelta > 0
          ? 'text-viz-sage'
          : pendingDelta < 0
            ? 'text-viz-terracotta'
            : 'text-muted-foreground'}"
      >
        {#if pendingDelta > 0}▲ +{pendingDelta} from yesterday
        {:else if pendingDelta < 0}▼ {pendingDelta} from yesterday
        {:else}— No change vs yesterday{/if}
      </p>
    {/if}
    <a
      href="/applications"
      class="relative mt-8 inline-flex items-center rounded-md bg-foreground px-4 py-2.5 text-sm font-medium text-cream transition hover:bg-foreground/90"
    >
      View review queue →
    </a>
  </article>

  <!-- Applications over time -->
  <article class="rounded-xl border border-border/70 bg-card p-6 lg:col-span-8">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h2 class="text-base font-medium text-foreground">Applications over time</h2>
        <p class="mt-0.5 text-xs text-muted-foreground">Submitted applications</p>
      </div>
      <span
        class="rounded-md border border-border/80 bg-background px-2.5 py-1 text-xs text-muted-foreground"
      >
        This week
      </span>
    </div>

    <div class="relative mt-4">
      <svg viewBox="0 0 {lineChart.w} {lineChart.h}" class="h-48 w-full" role="img" aria-label="Applications over time">
        <!-- grid -->
        {#each [0.25, 0.5, 0.75, 1] as t}
          <line
            x1={lineChart.padX}
            x2={lineChart.w - lineChart.padX}
            y1={lineChart.padY + (lineChart.h - lineChart.padY * 2) * (1 - t)}
            y2={lineChart.padY + (lineChart.h - lineChart.padY * 2) * (1 - t)}
            class="stroke-border/60"
            stroke-width="1"
          />
        {/each}
        <path d={lineChart.area} class="fill-viz-tan/25" />
        <path
          d={lineChart.path}
          fill="none"
          class="stroke-foreground"
          stroke-width="2.25"
          stroke-linecap="round"
          stroke-linejoin="round"
        />
        {#each lineChart.points as p}
          <circle cx={p.x} cy={p.y} r="3.5" class="fill-foreground" />
        {/each}
        {#if lineChart.peak && lineChart.peak.v > 0}
          <g>
            <rect
              x={Math.min(Math.max(lineChart.peak.x - 54, 8), lineChart.w - 116)}
              y={lineChart.peak.y - 44}
              width="108"
              height="34"
              rx="6"
              class="fill-foreground"
            />
            <text
              x={Math.min(Math.max(lineChart.peak.x - 54, 8), lineChart.w - 116) + 54}
              y={lineChart.peak.y - 30}
              text-anchor="middle"
              class="fill-cream"
              style="font-size: 9px; font-family: Inter, sans-serif"
            >
              {weekdayShort(lineChart.peak.day)}, {lineChart.peak.day.toLocaleDateString('en-US', {
                month: 'short',
                day: 'numeric'
              })}
            </text>
            <text
              x={Math.min(Math.max(lineChart.peak.x - 54, 8), lineChart.w - 116) + 54}
              y={lineChart.peak.y - 16}
              text-anchor="middle"
              class="fill-cream"
              style="font-size: 10px; font-weight: 600; font-family: Inter, sans-serif"
            >
              {lineChart.peak.v} submissions
            </text>
          </g>
        {/if}
      </svg>
      <div class="mt-1 flex justify-between px-2 text-[11px] text-muted-foreground">
        {#each weekDays as d}
          <span>{weekdayShort(d)}</span>
        {/each}
      </div>
    </div>
  </article>
</div>

<!-- Row 2 -->
<div class="mt-4 grid gap-4 lg:grid-cols-12">
  <article class="rounded-xl border border-border/70 bg-card p-6 lg:col-span-4">
    <h2 class="text-base font-medium text-foreground">Submission funnel</h2>
    <div class="mt-6 space-y-4">
      {#each funnel as row}
        <div>
          <div class="mb-1.5 flex justify-between text-sm">
            <span class="text-muted-foreground">{row.label}</span>
            <span class="font-medium text-foreground">{loading ? '—' : row.count}</span>
          </div>
          <div class="h-3 overflow-hidden rounded-sm bg-muted/60">
            <div
              class="h-full rounded-sm transition-all duration-700 {row.color}"
              style:width={loading ? '0%' : row.width}
            ></div>
          </div>
        </div>
      {/each}
    </div>
    <a
      href="/applications"
      class="mt-6 inline-block text-sm text-viz-navy underline-offset-4 hover:underline"
    >
      View full pipeline →
    </a>
  </article>

  <article class="rounded-xl border border-border/70 bg-card p-6 lg:col-span-4">
    <h2 class="text-base font-medium text-foreground">Recent activity</h2>
    {#if loading}
      <p class="mt-6 text-sm text-muted-foreground">Loading…</p>
    {:else if recentActivity.length === 0}
      <p class="mt-6 text-sm text-muted-foreground">No recent activity.</p>
    {:else}
      <ol class="relative mt-5 space-y-0">
        {#each recentActivity as item, i}
          <li class="relative flex gap-3 pb-5 last:pb-0">
            {#if i < recentActivity.length - 1}
              <span
                class="absolute top-3 left-[5px] h-[calc(100%-4px)] w-px bg-border"
                aria-hidden="true"
              ></span>
            {/if}
            <span
              class="relative z-10 mt-1.5 h-2.5 w-2.5 shrink-0 rounded-full ring-4 ring-card {toneDot(item.tone)}"
              aria-hidden="true"
            ></span>
            <div class="min-w-0 flex-1">
              <p class="truncate text-sm leading-snug text-foreground">{item.text}</p>
              <time class="mt-0.5 block text-xs text-muted-foreground">{item.when}</time>
            </div>
          </li>
        {/each}
      </ol>
    {/if}
  </article>

  <article class="rounded-xl border border-border/70 bg-card p-6 lg:col-span-4">
    <h2 class="text-base font-medium text-foreground">Archive health</h2>
    <p class="text-xs text-muted-foreground">Status overview</p>
    <div class="mt-4 flex items-center gap-6">
      <div class="relative h-36 w-36 shrink-0">
        <svg viewBox="0 0 100 100" class="h-full w-full -rotate-90">
          <circle cx="50" cy="50" r="38" fill="none" stroke-width="10" class="stroke-muted" />
          {#if !loading}
            {#each archiveHealth.segs as seg}
              <circle
                cx="50"
                cy="50"
                r="38"
                fill="none"
                stroke-width="10"
                stroke-dasharray={seg.dash}
                stroke-dashoffset={seg.offset}
                class={seg.tone === 'viz-olive'
                  ? 'stroke-viz-olive'
                  : seg.tone === 'viz-tan'
                    ? 'stroke-viz-tan'
                    : 'stroke-viz-terracotta'}
              />
            {/each}
          {/if}
        </svg>
        <div class="absolute inset-0 flex flex-col items-center justify-center text-center">
          <span class="font-display text-3xl leading-none text-foreground">
            {loading ? '—' : `${archiveHealth.approvedPct}%`}
          </span>
          <span class="mt-1 text-[10px] uppercase tracking-wider text-muted-foreground">Approved</span>
        </div>
      </div>
      <ul class="space-y-2.5 text-sm">
        {#each archiveHealth.segs as seg}
          <li class="flex items-center gap-2">
            <span
              class="h-2.5 w-2.5 rounded-full {seg.tone === 'viz-olive'
                ? 'bg-viz-olive'
                : seg.tone === 'viz-tan'
                  ? 'bg-viz-tan'
                  : 'bg-viz-terracotta'}"
            ></span>
            <span class="text-muted-foreground">{seg.label}</span>
            <span class="ml-auto font-medium text-foreground">{seg.pct}%</span>
          </li>
        {/each}
      </ul>
    </div>
  </article>
</div>

<!-- Row 3: sparklines -->
<div class="mt-4 grid gap-4 sm:grid-cols-3">
  {#each [
    { label: 'Artists', spark: sparkArtists, href: '/artists' },
    { label: 'Posts', spark: sparkPosts, href: '/posts' },
    { label: 'Events', spark: sparkEvents, href: '/events' }
  ] as card}
    <a
      href={card.href}
      class="rounded-xl border border-border/70 bg-card p-5 transition hover:border-foreground/30"
    >
      <div class="flex items-start justify-between gap-3">
        <div>
          <p class="text-sm text-muted-foreground">{card.label}</p>
          <p class="mt-1 font-display text-3xl text-foreground">
            {loading ? '—' : card.spark.total.toLocaleString()}
          </p>
          {#if !loading && card.spark.change !== null}
            <p
              class="mt-1 text-xs {card.spark.change >= 0
                ? 'text-viz-sage'
                : 'text-viz-terracotta'}"
            >
              {card.spark.change >= 0 ? '▲' : '▼'}
              {Math.abs(card.spark.change)}% vs last 7 days
            </p>
          {:else if !loading}
            <p class="mt-1 text-xs text-muted-foreground">No prior week data</p>
          {/if}
        </div>
        <svg viewBox="0 0 120 36" class="mt-1 h-10 w-28 text-foreground/70" aria-hidden="true">
          <path
            d={card.spark.path}
            fill="none"
            stroke="currentColor"
            stroke-width="1.75"
            stroke-linecap="round"
            stroke-linejoin="round"
          />
        </svg>
      </div>
    </a>
  {/each}
</div>

<!-- Row 4 -->
<div class="mt-4 grid gap-4 lg:grid-cols-12">
  <article class="rounded-xl border border-border/70 bg-card p-6 lg:col-span-4">
    <h2 class="text-base font-medium text-foreground">Content by type</h2>
    {#if loading}
      <p class="mt-6 text-sm text-muted-foreground">Loading…</p>
    {:else if contentByType.length === 0}
      <p class="mt-6 text-sm text-muted-foreground">No posts yet.</p>
    {:else}
      <div class="mt-6 space-y-4">
        {#each contentByType as row, i}
          <div>
            <div class="mb-1.5 flex justify-between text-sm">
              <span class="text-muted-foreground">{row.label}</span>
              <span class="font-medium text-foreground">{row.pct}%</span>
            </div>
            <div class="h-2.5 overflow-hidden rounded-sm bg-muted/60">
              <div
                class="h-full rounded-sm {i === 0
                  ? 'bg-foreground'
                  : i === 1
                    ? 'bg-viz-tan'
                    : i === 2
                      ? 'bg-viz-olive'
                      : i === 3
                        ? 'bg-viz-navy'
                        : 'bg-viz-terracotta'}"
                style:width={row.width}
              ></div>
            </div>
          </div>
        {/each}
      </div>
    {/if}
  </article>

  <article class="rounded-xl border border-border/70 bg-card p-6 lg:col-span-4">
    <h2 class="text-base font-medium text-foreground">Submissions heatmap</h2>
    <div class="mt-5 overflow-x-auto">
      <div
        class="grid gap-1"
        style="grid-template-columns: 1.5rem repeat({heatmap.hours.length}, minmax(1.6rem, 1fr))"
      >
        <div></div>
        {#each heatmap.hours as h}
          <div class="text-center text-[10px] text-muted-foreground">
            {h > 12 ? `${h - 12} PM` : h === 12 ? '12 PM' : `${h} AM`}
          </div>
        {/each}
        {#each heatmap.dayLabels as label, di}
          <div class="flex items-center text-[11px] text-muted-foreground">{label}</div>
          {#each heatmap.grid[di] as cell, hi}
            <div
              class="aspect-square rounded-sm bg-foreground"
              style:opacity={heatOpacity(cell, heatmap.max)}
              title="{label} {heatmap.hours[hi]}:00 — {cell}"
            ></div>
          {/each}
        {/each}
      </div>
    </div>
  </article>

  <article class="rounded-xl border border-border/70 bg-card p-6 lg:col-span-4">
    <h2 class="text-base font-medium text-foreground">Recent approvals</h2>
    {#if loading}
      <p class="mt-6 text-sm text-muted-foreground">Loading…</p>
    {:else if recentApprovals.length === 0}
      <p class="mt-6 text-sm text-muted-foreground">No approvals yet.</p>
    {:else}
      <ul class="mt-5 space-y-4">
        {#each recentApprovals as item}
          <li>
            <a href={item.href} class="group flex gap-3">
              <div
                class="h-12 w-12 shrink-0 overflow-hidden rounded-md bg-muted"
              >
                {#if item.image}
                  <img src={item.image} alt="" class="h-full w-full object-cover" />
                {:else}
                  <div class="flex h-full w-full items-center justify-center">
                    <BotanicalMark class="h-7 w-6 text-foreground/25" />
                  </div>
                {/if}
              </div>
              <div class="min-w-0">
                <p class="truncate text-sm font-medium text-foreground group-hover:text-accent">
                  {item.title}
                </p>
                <p class="truncate text-xs text-muted-foreground">{item.meta}</p>
                <p class="text-xs text-muted-foreground/80">{item.when}</p>
              </div>
            </a>
          </li>
        {/each}
      </ul>
    {/if}
  </article>
</div>
