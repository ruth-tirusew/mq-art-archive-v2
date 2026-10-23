// packages/ui has no node_modules of its own; each app aliases these two at runtime (see vite.config.ts).
// Declared here so type-checking doesn't depend on which app is checking.
declare module 'clsx' {
	export type ClassValue = string | number | bigint | boolean | null | undefined | ClassValue[] | { [key: string]: unknown };
	export function clsx(...inputs: ClassValue[]): string;
}
declare module 'tailwind-merge' {
	export function twMerge(...classLists: Array<string | null | undefined | false>): string;
}
