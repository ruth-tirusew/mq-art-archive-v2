/** Shared surface for Input, Textarea and Select so every form control matches. */
export const fieldClass =
	'w-full rounded-sm border border-input bg-card px-3.5 py-2.5 text-sm text-foreground placeholder:text-muted-foreground/70 transition-colors ' +
	'hover:border-foreground/30 focus-visible:border-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring/30 ' +
	'disabled:cursor-not-allowed disabled:opacity-50 aria-[invalid=true]:border-destructive aria-[invalid=true]:ring-destructive/20';
