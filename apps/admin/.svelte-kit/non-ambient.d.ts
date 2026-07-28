
// this file is generated — do not edit it


declare module "svelte/elements" {
	export interface HTMLAttributes<T> {
		'data-sveltekit-keepfocus'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-noscroll'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-preload-code'?:
			| true
			| ''
			| 'eager'
			| 'viewport'
			| 'hover'
			| 'tap'
			| 'off'
			| undefined
			| null;
		'data-sveltekit-preload-data'?: true | '' | 'hover' | 'tap' | 'off' | undefined | null;
		'data-sveltekit-reload'?: true | '' | 'off' | undefined | null;
		'data-sveltekit-replacestate'?: true | '' | 'off' | undefined | null;
	}
}

export {};


declare module "$app/types" {
	type MatcherParam<M> = M extends (param : string) => param is (infer U extends string) ? U : string;

	export interface AppTypes {
		RouteId(): "/" | "/applications" | "/applications/[id]" | "/artists" | "/artists/new" | "/artists/[id]" | "/auth" | "/auth/callback" | "/events" | "/events/new" | "/events/[id]" | "/login" | "/posts" | "/posts/new" | "/posts/[id]" | "/settings" | "/users" | "/wiki" | "/wiki/new" | "/wiki/submissions" | "/wiki/[id]";
		RouteParams(): {
			"/applications/[id]": { id: string };
			"/artists/[id]": { id: string };
			"/events/[id]": { id: string };
			"/posts/[id]": { id: string };
			"/wiki/[id]": { id: string }
		};
		LayoutParams(): {
			"/": { id?: string | undefined };
			"/applications": { id?: string | undefined };
			"/applications/[id]": { id: string };
			"/artists": { id?: string | undefined };
			"/artists/new": Record<string, never>;
			"/artists/[id]": { id: string };
			"/auth": Record<string, never>;
			"/auth/callback": Record<string, never>;
			"/events": { id?: string | undefined };
			"/events/new": Record<string, never>;
			"/events/[id]": { id: string };
			"/login": Record<string, never>;
			"/posts": { id?: string | undefined };
			"/posts/new": Record<string, never>;
			"/posts/[id]": { id: string };
			"/settings": Record<string, never>;
			"/users": Record<string, never>;
			"/wiki": { id?: string | undefined };
			"/wiki/new": Record<string, never>;
			"/wiki/submissions": Record<string, never>;
			"/wiki/[id]": { id: string }
		};
		Pathname(): "/" | "/applications" | `/applications/${string}` & {} | "/artists" | "/artists/new" | `/artists/${string}` & {} | "/auth/callback" | "/events" | "/events/new" | `/events/${string}` & {} | "/login" | "/posts" | "/posts/new" | `/posts/${string}` & {} | "/settings" | "/users" | "/wiki" | "/wiki/new" | "/wiki/submissions" | `/wiki/${string}` & {};
		ResolvedPathname(): `${"" | `/${string}`}${ReturnType<AppTypes['Pathname']>}`;
		Asset(): "/adornments/botanical-sheet.png" | "/adornments/bouquet.png" | "/adornments/buds.png" | "/adornments/daisies.png" | "/adornments/grass.png" | "/adornments/lily-test.png" | "/adornments/lily.png" | "/adornments/peony.png" | "/adornments/slender.png" | "/adornments/sprig.png" | string & {};
	}
}