/** What went wrong, independent of how it was transported. Screens branch on this, not on HTTP codes. */
export type ErrorKind =
	| 'invalid'
	| 'unauthorized'
	| 'forbidden'
	| 'not_found'
	| 'conflict'
	| 'rate_limited'
	| 'unknown';

export class DomainError extends Error {
	constructor(
		message: string,
		readonly kind: ErrorKind
	) {
		super(message);
		this.name = 'DomainError';
	}
}

export function errorKind(error: unknown): ErrorKind {
	return error instanceof DomainError ? error.kind : 'unknown';
}

export function isErrorKind(error: unknown, kind: ErrorKind): boolean {
	return errorKind(error) === kind;
}

export function kindForStatus(status: number): ErrorKind {
	switch (status) {
		case 400:
		case 422:
			return 'invalid';
		case 401:
			return 'unauthorized';
		case 403:
			return 'forbidden';
		case 404:
			return 'not_found';
		case 409:
			return 'conflict';
		case 429:
			return 'rate_limited';
		default:
			return 'unknown';
	}
}
