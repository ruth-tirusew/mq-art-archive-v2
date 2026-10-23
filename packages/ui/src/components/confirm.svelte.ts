type Ask = {
	title: string;
	description?: string;
	confirmLabel?: string;
	/** Irreversible or public-facing (delete, unpublish, reject, archive). */
	destructive?: boolean;
	/** Runs on confirm. The dialog shows a busy state until it settles, then closes. */
	run: () => void | Promise<void>;
};

/**
 * One confirm dialog per page, many actions:
 *   const confirm = createConfirm();
 *   confirm.ask({ title: 'Unpublish?', destructive: true, run: () => setStatus('draft') });
 *   <ConfirmDialog bind:open={confirm.open} {...confirm.dialog} />
 * `run` should handle its own errors (show them inline or via toast); it is not rethrown.
 */
export function createConfirm() {
	let open = $state(false);
	let busy = $state(false);
	let config = $state<Omit<Ask, 'run'>>({ title: '' });
	let run: Ask['run'] = () => {};

	async function onconfirm() {
		busy = true;
		try {
			await run();
		} finally {
			busy = false;
			open = false;
		}
	}

	return {
		get open() {
			return open;
		},
		set open(value: boolean) {
			open = value;
		},
		get dialog() {
			return {
				title: config.title,
				description: config.description,
				confirmLabel: config.confirmLabel,
				destructive: config.destructive ?? false,
				busy,
				onconfirm
			};
		},
		ask({ run: next, ...rest }: Ask) {
			config = rest;
			run = next;
			open = true;
		}
	};
}
