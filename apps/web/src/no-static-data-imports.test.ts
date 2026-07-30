import { readdirSync, readFileSync } from 'node:fs';
import { extname, join } from 'node:path';
import { describe, expect, it } from 'vitest';

const sourceRoot = join(process.cwd(), 'src');

function productionFiles(directory: string): string[] {
	return readdirSync(directory, { withFileTypes: true }).flatMap((entry) => {
		const path = join(directory, entry.name);
		if (entry.isDirectory()) return productionFiles(path);
		if (!['.ts', '.svelte'].includes(extname(entry.name)) || entry.name.includes('.test.')) return [];
		return [path];
	});
}

describe('production data sources', () => {
	it('does not import hardcoded runtime catalogs', () => {
		const offenders = productionFiles(sourceRoot).filter((path) =>
			readFileSync(path, 'utf8').includes('$lib/data')
		);

		expect(offenders).toEqual([]);
	});
});
