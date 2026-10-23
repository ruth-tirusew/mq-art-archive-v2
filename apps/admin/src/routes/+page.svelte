<script lang="ts">
  import { goto } from '$app/navigation';
  import { ApiError } from '$lib/adapters/api/client';
  import { ErrorState } from '$ui';
  import { Skeleton } from '$lib/components/ui/skeleton';
  import { onMount } from 'svelte';
  import { artistsService } from '$lib/application/artists';
  import { postsService } from '$lib/application/posts';
  import { onboardingService } from '$lib/application/onboarding';
  import { eventsService } from '$lib/application/events';
  import { getAnalyticsSummary, type AnalyticsSummary } from '$lib/application/analytics';
  import { plural } from '$lib/utils/plural';
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

  async function load() {
    errorMessage = '';
    loading = true;
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
      // This page loads on mount rather than in a guarded loader, so handle an expired session here.
      if (err instanceof ApiError && (err.status === 401 || err.status === 403)) {
        await goto('/login');
        return;
      }
      errorMessage = err instanceof Error ? err.message : 'Failed to load dashboard';
    } finally {
      loading = false;
    }
  }

  onMount(load);

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

  /** Only items an admin can actually act on: applications, profiles and events awaiting a decision. */
  const pendingReview = $derived(applications.length + artistCounts.pending + events.length);

  const queueCreatedDates = $derived([
    ...applications.map((a) => a.created_at),
    ...artists.filter((a) => a.status === 'pending').map((a) => a.created_at),
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
  const seriesOverTime = $derived(appsOverTime);

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
  <div class="mx-auto max-w-2xl py-10"><ErrorState message={errorMessage} onretry={load} /></div>
{:else}

<!-- Page header -->
<header class="mb-8 flex flex-col gap-4 lg:flex-row lg:items-end lg:justify-between">
  <div>
    <p class="text-[11px] font-medium uppercase tracking-[0.22em] text-muted-foreground">
      Operations
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
    <div class="mt-1 text-sm text-muted-foreground">
      {#if loading}
        <Skeleton class="h-4 w-48 lg:ml-auto" />
      {:else}
        <span class="font-medium text-foreground">{pendingReview}</span>
        {pendingReview === 1 ? 'item needs' : 'items need'} your attention today.
      {/if}
    </div>
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
      variant="daisies"
      class="pointer-events-none absolute -bottom-4 -right-3 w-32 opacity-[0.14] sm:w-36"
    />
    <p class="relative text-sm text-muted-foreground">Awaiting review</p>
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
      Review applications →
    </a>
  </article>

  <!-- Applications over time -->
  <article class="rounded-xl border border-border/70 bg-card p-6 lg:col-span-8">
    <div class="flex items-start justify-between gap-4">
      <div>
        <h2 class="text-base font-medium text-foreground">Applications over time</h2>
        <p class="mt-0.5 text-xs text-muted-foreground">Pending applications, by day received</p>
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
              {plural(lineChart.peak.v, 'submission')}
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
    <h2 class="text-base font-medium text-foreground">Recent activity</h2>
    {#if loading}
      <div class="mt-6 space-y-2" role="status" aria-live="polite">
        <span class="sr-only">Loading</span>
        <Skeleton class="h-10 w-full" />
        <Skeleton class="h-10 w-full" />
        <Skeleton class="h-10 w-full" />
      </div>
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
    <h2 class="text-base font-medium text-foreground">Content by type</h2>
    {#if loading}
      <div class="mt-6 space-y-2" role="status" aria-live="polite">
        <span class="sr-only">Loading</span>
        <Skeleton class="h-10 w-full" />
        <Skeleton class="h-10 w-full" />
        <Skeleton class="h-10 w-full" />
      </div>
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
    <h2 class="text-base font-medium text-foreground">Recent approvals</h2>
    {#if loading}
      <div class="mt-6 space-y-2" role="status" aria-live="polite">
        <span class="sr-only">Loading</span>
        <Skeleton class="h-10 w-full" />
        <Skeleton class="h-10 w-full" />
        <Skeleton class="h-10 w-full" />
      </div>
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

<!-- Row 3: sparklines -->
<div class="mt-4 grid gap-4 sm:grid-cols-2">
  {#each [
    { label: 'Artists', spark: sparkArtists, href: '/artists' },
    { label: 'Posts', spark: sparkPosts, href: '/posts' }
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
{/if}
