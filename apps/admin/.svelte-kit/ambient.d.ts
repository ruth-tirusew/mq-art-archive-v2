
// this file is generated — do not edit it


/// <reference types="@sveltejs/kit" />

/**
 * This module provides access to environment variables that are injected _statically_ into your bundle at build time and are limited to _private_ access.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Static environment variables are [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env` at build time and then statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * **_Private_ access:**
 * 
 * - This module cannot be imported into client-side code
 * - This module only includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured)
 * 
 * For example, given the following build time environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { ENVIRONMENT, PUBLIC_BASE_URL } from '$env/static/private';
 * 
 * console.log(ENVIRONMENT); // => "production"
 * console.log(PUBLIC_BASE_URL); // => throws error during build
 * ```
 * 
 * The above values will be the same _even if_ different values for `ENVIRONMENT` or `PUBLIC_BASE_URL` are set at runtime, as they are statically replaced in your code with their build time values.
 */
declare module '$env/static/private' {
	export const SVELTEKIT_FORK: string;
	export const NODE_ENV: string;
	export const _ZO_DOCTOR: string;
	export const CURSOR_CONVERSATION_ID: string;
	export const XDG_DATA_DIRS: string;
	export const NVM_CD_FLAGS: string;
	export const npm_execpath: string;
	export const PWD: string;
	export const npm_config_globalconfig: string;
	export const __MISE_ZSH_PRECMD_RUN: string;
	export const XDG_VTNR: string;
	export const VSCODE_CODE_CACHE_PATH: string;
	export const QT_IM_MODULE: string;
	export const DRI_PRIME: string;
	export const npm_config_init_module: string;
	export const QT_ACCESSIBILITY: string;
	export const npm_lifecycle_event: string;
	export const npm_lifecycle_script: string;
	export const XDG_SESSION_DESKTOP: string;
	export const LS_COLORS: string;
	export const COSMIC_PANEL_NAME: string;
	export const LANG: string;
	export const _JAVA_AWT_WM_NONREPARENTING: string;
	export const AGENT_TRANSCRIPTS: string;
	export const CURSOR_AGENT: string;
	export const COSMIC_PANEL_ANCHOR: string;
	export const XDG_RUNTIME_DIR: string;
	export const VSCODE_PROCESS_TITLE: string;
	export const npm_package_name: string;
	export const NODE: string;
	export const CURSOR_WORKSPACE_LABEL: string;
	export const XMODIFIERS: string;
	export const NVM_BIN: string;
	export const QT_AUTO_SCREEN_SCALE_FACTOR: string;
	export const GTK_MODULES: string;
	export const GTK_IM_MODULE: string;
	export const NVM_INC: string;
	export const QT_QPA_PLATFORM: string;
	export const MISE_SHELL: string;
	export const GDK_BACKEND: string;
	export const NVM_DIR: string;
	export const __MISE_DIFF: string;
	export const HOME: string;
	export const CHROME_DESKTOP: string;
	export const npm_config_noproxy: string;
	export const npm_config_npm_version: string;
	export const npm_node_execpath: string;
	export const SHLVL: string;
	export const DISPLAY: string;
	export const VSCODE_IPC_HOOK: string;
	export const COSMIC_PANEL_OUTPUT: string;
	export const MOZ_ENABLE_WAYLAND: string;
	export const WAYLAND_DISPLAY: string;
	export const VSCODE_CWD: string;
	export const FORCE_COLOR: string;
	export const FC_FONTATIONS: string;
	export const USER: string;
	export const COSMIC_PANEL_BACKGROUND: string;
	export const QT_QPA_PLATFORMTHEME: string;
	export const PATH: string;
	export const VSCODE_NLS_CONFIG: string;
	export const OLDPWD: string;
	export const QT_ENABLE_HIGHDPI_SCALING: string;
	export const CURSOR_RIPGREP_PATH: string;
	export const CLUTTER_IM_MODULE: string;
	export const SSH_AUTH_SOCK: string;
	export const X_MINIMIZE_APPLET: string;
	export const CURSOR_LAYOUT: string;
	export const npm_config_global_prefix: string;
	export const VSCODE_ESM_ENTRYPOINT: string;
	export const npm_package_version: string;
	export const XDG_SEAT: string;
	export const npm_command: string;
	export const LOGNAME: string;
	export const INIT_CWD: string;
	export const __MISE_ORIG_PATH: string;
	export const XDG_SESSION_TYPE: string;
	export const COSMIC_PANEL_SIZE: string;
	export const npm_config_userconfig: string;
	export const npm_config_local_prefix: string;
	export const __MISE_ZSH_CHPWD_RAN: string;
	export const CURSOR_EXTENSION_HOST_ROLE: string;
	export const X_PRIVILEGED_WAYLAND_SOCKET: string;
	export const GSM_SKIP_SSH_AGENT_WORKAROUND: string;
	export const VSCODE_PID: string;
	export const __MISE_SESSION: string;
	export const NO_COLOR: string;
	export const npm_config_prefix: string;
	export const COLOR: string;
	export const npm_config_user_agent: string;
	export const COSMIC_PANEL_PADDING_OVERLAP: string;
	export const PANEL_NOTIFICATIONS_FD: string;
	export const npm_package_json: string;
	export const DBUS_SESSION_BUS_ADDRESS: string;
	export const VSCODE_CRASH_REPORTER_PROCESS_TYPE: string;
	export const EDITOR: string;
	export const IM_CONFIG_PHASE: string;
	export const XDG_CURRENT_DESKTOP: string;
	export const _: string;
	export const TERM: string;
	export const DCONF_PROFILE: string;
	export const VSCODE_HANDLES_UNCAUGHT_ERRORS: string;
	export const XDG_SESSION_ID: string;
	export const npm_config_cache: string;
	export const COSMIC_PANEL_SPACING: string;
	export const SHELL: string;
	export const npm_config_node_gyp: string;
}

/**
 * This module provides access to environment variables that are injected _statically_ into your bundle at build time and are _publicly_ accessible.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Static environment variables are [loaded by Vite](https://vitejs.dev/guide/env-and-mode.html#env-files) from `.env` files and `process.env` at build time and then statically injected into your bundle at build time, enabling optimisations like dead code elimination.
 * 
 * **_Public_ access:**
 * 
 * - This module _can_ be imported into client-side code
 * - **Only** variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`) are included
 * 
 * For example, given the following build time environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { ENVIRONMENT, PUBLIC_BASE_URL } from '$env/static/public';
 * 
 * console.log(ENVIRONMENT); // => throws error during build
 * console.log(PUBLIC_BASE_URL); // => "http://site.com"
 * ```
 * 
 * The above values will be the same _even if_ different values for `ENVIRONMENT` or `PUBLIC_BASE_URL` are set at runtime, as they are statically replaced in your code with their build time values.
 */
declare module '$env/static/public' {
	export const PUBLIC_API_URL: string;
}

/**
 * This module provides access to environment variables set _dynamically_ at runtime and that are limited to _private_ access.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Dynamic environment variables are defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`.
 * 
 * **_Private_ access:**
 * 
 * - This module cannot be imported into client-side code
 * - This module includes variables that _do not_ begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) _and do_ start with [`config.kit.env.privatePrefix`](https://svelte.dev/docs/kit/configuration#env) (if configured)
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 * 
 * > [!NOTE] To get correct types, environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * >
 * > ```env
 * > MY_FEATURE_FLAG=
 * > ```
 * >
 * > You can override `.env` values from the command line like so:
 * >
 * > ```sh
 * > MY_FEATURE_FLAG="enabled" npm run dev
 * > ```
 * 
 * For example, given the following runtime environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://site.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { env } from '$env/dynamic/private';
 * 
 * console.log(env.ENVIRONMENT); // => "production"
 * console.log(env.PUBLIC_BASE_URL); // => undefined
 * ```
 */
declare module '$env/dynamic/private' {
	export const env: {
		SVELTEKIT_FORK: string;
		NODE_ENV: string;
		_ZO_DOCTOR: string;
		CURSOR_CONVERSATION_ID: string;
		XDG_DATA_DIRS: string;
		NVM_CD_FLAGS: string;
		npm_execpath: string;
		PWD: string;
		npm_config_globalconfig: string;
		__MISE_ZSH_PRECMD_RUN: string;
		XDG_VTNR: string;
		VSCODE_CODE_CACHE_PATH: string;
		QT_IM_MODULE: string;
		DRI_PRIME: string;
		npm_config_init_module: string;
		QT_ACCESSIBILITY: string;
		npm_lifecycle_event: string;
		npm_lifecycle_script: string;
		XDG_SESSION_DESKTOP: string;
		LS_COLORS: string;
		COSMIC_PANEL_NAME: string;
		LANG: string;
		_JAVA_AWT_WM_NONREPARENTING: string;
		AGENT_TRANSCRIPTS: string;
		CURSOR_AGENT: string;
		COSMIC_PANEL_ANCHOR: string;
		XDG_RUNTIME_DIR: string;
		VSCODE_PROCESS_TITLE: string;
		npm_package_name: string;
		NODE: string;
		CURSOR_WORKSPACE_LABEL: string;
		XMODIFIERS: string;
		NVM_BIN: string;
		QT_AUTO_SCREEN_SCALE_FACTOR: string;
		GTK_MODULES: string;
		GTK_IM_MODULE: string;
		NVM_INC: string;
		QT_QPA_PLATFORM: string;
		MISE_SHELL: string;
		GDK_BACKEND: string;
		NVM_DIR: string;
		__MISE_DIFF: string;
		HOME: string;
		CHROME_DESKTOP: string;
		npm_config_noproxy: string;
		npm_config_npm_version: string;
		npm_node_execpath: string;
		SHLVL: string;
		DISPLAY: string;
		VSCODE_IPC_HOOK: string;
		COSMIC_PANEL_OUTPUT: string;
		MOZ_ENABLE_WAYLAND: string;
		WAYLAND_DISPLAY: string;
		VSCODE_CWD: string;
		FORCE_COLOR: string;
		FC_FONTATIONS: string;
		USER: string;
		COSMIC_PANEL_BACKGROUND: string;
		QT_QPA_PLATFORMTHEME: string;
		PATH: string;
		VSCODE_NLS_CONFIG: string;
		OLDPWD: string;
		QT_ENABLE_HIGHDPI_SCALING: string;
		CURSOR_RIPGREP_PATH: string;
		CLUTTER_IM_MODULE: string;
		SSH_AUTH_SOCK: string;
		X_MINIMIZE_APPLET: string;
		CURSOR_LAYOUT: string;
		npm_config_global_prefix: string;
		VSCODE_ESM_ENTRYPOINT: string;
		npm_package_version: string;
		XDG_SEAT: string;
		npm_command: string;
		LOGNAME: string;
		INIT_CWD: string;
		__MISE_ORIG_PATH: string;
		XDG_SESSION_TYPE: string;
		COSMIC_PANEL_SIZE: string;
		npm_config_userconfig: string;
		npm_config_local_prefix: string;
		__MISE_ZSH_CHPWD_RAN: string;
		CURSOR_EXTENSION_HOST_ROLE: string;
		X_PRIVILEGED_WAYLAND_SOCKET: string;
		GSM_SKIP_SSH_AGENT_WORKAROUND: string;
		VSCODE_PID: string;
		__MISE_SESSION: string;
		NO_COLOR: string;
		npm_config_prefix: string;
		COLOR: string;
		npm_config_user_agent: string;
		COSMIC_PANEL_PADDING_OVERLAP: string;
		PANEL_NOTIFICATIONS_FD: string;
		npm_package_json: string;
		DBUS_SESSION_BUS_ADDRESS: string;
		VSCODE_CRASH_REPORTER_PROCESS_TYPE: string;
		EDITOR: string;
		IM_CONFIG_PHASE: string;
		XDG_CURRENT_DESKTOP: string;
		_: string;
		TERM: string;
		DCONF_PROFILE: string;
		VSCODE_HANDLES_UNCAUGHT_ERRORS: string;
		XDG_SESSION_ID: string;
		npm_config_cache: string;
		COSMIC_PANEL_SPACING: string;
		SHELL: string;
		npm_config_node_gyp: string;
		[key: `PUBLIC_${string}`]: undefined;
		[key: `${string}`]: string | undefined;
	}
}

/**
 * This module provides access to environment variables set _dynamically_ at runtime and that are _publicly_ accessible.
 * 
 * |         | Runtime                                                                    | Build time                                                               |
 * | ------- | -------------------------------------------------------------------------- | ------------------------------------------------------------------------ |
 * | Private | [`$env/dynamic/private`](https://svelte.dev/docs/kit/$env-dynamic-private) | [`$env/static/private`](https://svelte.dev/docs/kit/$env-static-private) |
 * | Public  | [`$env/dynamic/public`](https://svelte.dev/docs/kit/$env-dynamic-public)   | [`$env/static/public`](https://svelte.dev/docs/kit/$env-static-public)   |
 * 
 * Dynamic environment variables are defined by the platform you're running on. For example if you're using [`adapter-node`](https://github.com/sveltejs/kit/tree/main/packages/adapter-node) (or running [`vite preview`](https://svelte.dev/docs/kit/cli)), this is equivalent to `process.env`.
 * 
 * **_Public_ access:**
 * 
 * - This module _can_ be imported into client-side code
 * - **Only** variables that begin with [`config.kit.env.publicPrefix`](https://svelte.dev/docs/kit/configuration#env) (which defaults to `PUBLIC_`) are included
 * 
 * > [!NOTE] In `dev`, `$env/dynamic` includes environment variables from `.env`. In `prod`, this behavior will depend on your adapter.
 * 
 * > [!NOTE] To get correct types, environment variables referenced in your code should be declared (for example in an `.env` file), even if they don't have a value until the app is deployed:
 * >
 * > ```env
 * > MY_FEATURE_FLAG=
 * > ```
 * >
 * > You can override `.env` values from the command line like so:
 * >
 * > ```sh
 * > MY_FEATURE_FLAG="enabled" npm run dev
 * > ```
 * 
 * For example, given the following runtime environment:
 * 
 * ```env
 * ENVIRONMENT=production
 * PUBLIC_BASE_URL=http://example.com
 * ```
 * 
 * With the default `publicPrefix` and `privatePrefix`:
 * 
 * ```ts
 * import { env } from '$env/dynamic/public';
 * console.log(env.ENVIRONMENT); // => undefined, not public
 * console.log(env.PUBLIC_BASE_URL); // => "http://example.com"
 * ```
 * 
 * ```
 * 
 * ```
 */
declare module '$env/dynamic/public' {
	export const env: {
		PUBLIC_API_URL: string;
		[key: `PUBLIC_${string}`]: string | undefined;
	}
}
