import path from 'node:path';
import { fileURLToPath } from 'node:url';
import { defineConfig } from 'vitest/config';

const root = path.dirname(fileURLToPath(import.meta.url));

const shared = {
	resolve: {
		alias: {
			$lib: path.resolve(root, 'src/lib'),
			'$env/static/public': path.resolve(root, 'src/test/mocks/public-env.ts'),
			'$app/environment': path.resolve(root, 'src/test/mocks/app-environment.ts'),
			'$app/server': path.resolve(root, 'src/test/mocks/app-server.ts')
		}
	}
};

export default defineConfig({
	test: {
		projects: [
			{
				...shared,
				test: {
					name: 'unit',
					include: ['src/**/*.test.ts'],
					exclude: ['src/**/*.integration.test.ts', 'src/**/*.live.integration.test.ts'],
					environment: 'node',
					setupFiles: ['./src/test/setup.ts']
				}
			},
			{
				...shared,
				test: {
					name: 'integration',
					include: ['src/**/*.integration.test.ts'],
					exclude: ['src/**/*.live.integration.test.ts'],
					environment: 'node',
					setupFiles: ['./src/test/setup.integration.ts']
				}
			},
			{
				...shared,
				test: {
					name: 'live',
					include: ['src/**/*.live.integration.test.ts'],
					environment: 'node',
					setupFiles: ['./src/test/setup.live.ts'],
					passWithNoTests: true
				}
			}
		]
	}
});
