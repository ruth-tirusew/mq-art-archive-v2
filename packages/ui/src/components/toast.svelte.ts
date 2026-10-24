export type ToastTone = 'success' | 'error' | 'info';

export type ToastItem = { id: number; tone: ToastTone; message: string };

let items = $state<ToastItem[]>([]);
let next = 0;

function push(tone: ToastTone, message: string, ms = tone === 'error' ? 8000 : 4000) {
	const id = ++next;
	items = [...items, { id, tone, message }];
	setTimeout(() => dismiss(id), ms);
	return id;
}

function dismiss(id: number) {
	items = items.filter((t) => t.id !== id);
}

/** Fire-and-forget feedback for mutations: toast.success('Published.'). Mount <Toaster /> once in the root layout. */
export const toast = {
	success: (message: string) => push('success', message),
	error: (message: string) => push('error', message),
	info: (message: string) => push('info', message),
	dismiss,
	get items() {
		return items;
	}
};
