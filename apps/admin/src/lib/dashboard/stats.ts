/** Shared date helpers for editorial dashboard charts. */

export function startOfDay(d: Date): Date {
  const x = new Date(d);
  x.setHours(0, 0, 0, 0);
  return x;
}

export function parseDate(iso?: string | null): Date | null {
  if (!iso) return null;
  const d = new Date(iso);
  return Number.isNaN(d.getTime()) ? null : d;
}

export function daysAgo(n: number, from = new Date()): Date {
  const d = startOfDay(from);
  d.setDate(d.getDate() - n);
  return d;
}

export function dayKey(d: Date): string {
  return d.toISOString().slice(0, 10);
}

export function weekdayShort(d: Date): string {
  return d.toLocaleDateString('en-US', { weekday: 'short' });
}

/** Last `n` calendar days (oldest → newest), inclusive of today. */
export function lastNDays(n: number, from = new Date()): Date[] {
  return Array.from({ length: n }, (_, i) => daysAgo(n - 1 - i, from));
}

export function countByDay(
  dates: (string | undefined | null)[],
  days: Date[]
): number[] {
  const keys = new Set(days.map(dayKey));
  const map = new Map<string, number>();
  for (const k of keys) map.set(k, 0);
  for (const iso of dates) {
    const d = parseDate(iso);
    if (!d) continue;
    const k = dayKey(startOfDay(d));
    if (map.has(k)) map.set(k, (map.get(k) ?? 0) + 1);
  }
  return days.map((d) => map.get(dayKey(d)) ?? 0);
}

export function countInRange(
  dates: (string | undefined | null)[],
  from: Date,
  to: Date
): number {
  let n = 0;
  for (const iso of dates) {
    const d = parseDate(iso);
    if (!d) continue;
    if (d >= from && d < to) n += 1;
  }
  return n;
}

export function pctChange(current: number, previous: number): number | null {
  if (previous === 0) return current > 0 ? 100 : null;
  return Math.round(((current - previous) / previous) * 100);
}

export function relativeTime(iso?: string | null): string {
  const d = parseDate(iso);
  if (!d) return '';
  const diff = Date.now() - d.getTime();
  const mins = Math.floor(diff / 60000);
  if (mins < 1) return 'just now';
  if (mins < 60) return `${mins} min ago`;
  const hrs = Math.floor(mins / 60);
  if (hrs < 24) return `${hrs}h ago`;
  const days = Math.floor(hrs / 24);
  return `${days}d ago`;
}

export function sparklinePath(values: number[], w = 120, h = 36, pad = 2): string {
  if (values.length === 0) return '';
  const max = Math.max(...values, 1);
  const min = Math.min(...values, 0);
  const range = Math.max(max - min, 1);
  const step = values.length > 1 ? (w - pad * 2) / (values.length - 1) : 0;
  return values
    .map((v, i) => {
      const x = pad + i * step;
      const y = pad + (h - pad * 2) * (1 - (v - min) / range);
      return `${i === 0 ? 'M' : 'L'}${x.toFixed(1)},${y.toFixed(1)}`;
    })
    .join(' ');
}

export function normalizeMedium(raw?: string): string {
  if (!raw?.trim()) return 'Unspecified';
  const t = raw.trim();
  return t.charAt(0).toUpperCase() + t.slice(1);
}
