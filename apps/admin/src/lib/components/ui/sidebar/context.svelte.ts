import { getContext, setContext } from 'svelte';

const KEY = Symbol('sidebar');
const STORAGE_KEY = 'artiv-admin:sidebar-open';
const MOBILE_QUERY = '(max-width: 1023px)';

export class SidebarState {
  /** Desktop: expanded (true) or docked to icons (false). */
  open = $state(true);
  /** Mobile drawer. */
  openMobile = $state(false);
  isMobile = $state(false);

  readonly state = $derived(this.open ? 'expanded' : 'collapsed');
  /** True only when the desktop sidebar is docked to its icon rail. */
  readonly collapsed = $derived(!this.isMobile && !this.open);

  /** Reads the saved preference and starts tracking the viewport. Returns a cleanup function. */
  init() {
    try {
      const saved = localStorage.getItem(STORAGE_KEY);
      if (saved !== null) this.open = saved === 'true';
    } catch {
      // Storage can be unavailable (private mode); the default is fine.
    }
    const query = window.matchMedia(MOBILE_QUERY);
    const sync = () => {
      this.isMobile = query.matches;
      if (!query.matches) this.openMobile = false;
    };
    sync();
    query.addEventListener('change', sync);
    return () => query.removeEventListener('change', sync);
  }

  toggle = () => {
    if (this.isMobile) {
      this.openMobile = !this.openMobile;
      return;
    }
    this.open = !this.open;
    try {
      localStorage.setItem(STORAGE_KEY, String(this.open));
    } catch {
      // Ignore: preference just won't persist.
    }
  };
}

export function setSidebar(): SidebarState {
  return setContext(KEY, new SidebarState());
}

export function useSidebar(): SidebarState {
  return getContext<SidebarState>(KEY);
}
