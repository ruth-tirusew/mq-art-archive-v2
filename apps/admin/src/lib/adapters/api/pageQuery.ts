export type PageParams = { limit?: number; offset?: number };

export function pageQuery(status: string | undefined, page?: PageParams): string {
  const p = new URLSearchParams();
  if (status) p.set('status', status);
  if (page?.limit) p.set('limit', String(page.limit));
  if (page?.offset) p.set('offset', String(page.offset));
  const q = p.toString();
  return q ? `?${q}` : '';
}

/** Pulls every page of a list endpoint. Stops at a short page, with a hard ceiling as a safety net. */
export async function fetchAll<T>(fetchPage: (page: PageParams) => Promise<T[]>, pageSize = 200, maxRows = 5000): Promise<T[]> {
  const out: T[] = [];
  while (out.length < maxRows) {
    const rows = await fetchPage({ limit: pageSize, offset: out.length });
    out.push(...rows);
    if (rows.length < pageSize) break;
  }
  return out;
}
