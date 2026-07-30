import { defineConfig, devices } from '@playwright/test';

const port = 4173;
const baseURL = `http://127.0.0.1:${port}`;
const apiURL = process.env.PUBLIC_API_URL ?? 'http://localhost:8080';

export default defineConfig({
	testDir: './e2e',
	fullyParallel: true,
	forbidOnly: !!process.env.CI,
	retries: process.env.CI ? 2 : 0,
	workers: process.env.CI ? 1 : undefined,
	reporter: 'list',
	use: {
		baseURL,
		trace: 'on-first-retry',
		...devices['Desktop Chrome']
	},
	projects: [{ name: 'chromium', use: { ...devices['Desktop Chrome'] } }],
	webServer: {
		command: `npm run build && npm run preview -- --host 127.0.0.1 --port ${port}`,
		url: baseURL,
		reuseExistingServer: !process.env.CI,
		timeout: 180_000,
		env: {
			...process.env,
			PUBLIC_API_URL: apiURL
		}
	}
});
