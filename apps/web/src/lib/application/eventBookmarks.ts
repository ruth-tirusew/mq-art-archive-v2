const STORAGE_KEY = 'mq:event-bookmarks';

function readIds(): string[] {
  if (typeof localStorage === 'undefined') return [];
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return [];
    const parsed = JSON.parse(raw) as unknown;
    return Array.isArray(parsed) ? parsed.filter((id): id is string => typeof id === 'string') : [];
  } catch {
    return [];
  }
}

function writeIds(ids: string[]) {
  if (typeof localStorage === 'undefined') return;
  localStorage.setItem(STORAGE_KEY, JSON.stringify(ids));
}

export function listBookmarkedEventIds(): string[] {
  return readIds();
}

export function isEventBookmarked(id: string): boolean {
  return readIds().includes(id);
}

export function toggleEventBookmark(id: string): boolean {
  const ids = readIds();
  const idx = ids.indexOf(id);
  if (idx >= 0) {
    ids.splice(idx, 1);
    writeIds(ids);
    return false;
  }
  ids.push(id);
  writeIds(ids);
  return true;
}
