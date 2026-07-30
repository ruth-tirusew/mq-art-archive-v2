import { test, expect } from '@playwright/test';

test.describe('@smoke home', () => {
	test('loads the homepage', async ({ page }) => {
		await page.goto('/');
		await expect(page.getByTestId('web-site-header')).toBeVisible();
		await expect(page.getByRole('link', { name: 'artiv.' })).toBeVisible();
	});
});

test.describe('@smoke artists', () => {
	test('renders the artists roster page', async ({ page }) => {
		await page.goto('/artists');
		await expect(page.getByTestId('web-site-header')).toBeVisible();
		await expect(page.getByRole('heading', { name: /roster/i })).toBeVisible();

		// Without a live API the roster is empty; with API (or seed) cards appear.
		const cards = page.getByTestId('web-artist-card');
		const unavailable = page.getByText(/Roster unavailable/i);
		await expect(cards.first().or(unavailable)).toBeVisible({ timeout: 15_000 });
	});
});

test.describe('@smoke wiki', () => {
	test('loads wiki index with empty or unavailable API states', async ({ page }) => {
		await page.goto('/wiki');
		await expect(page.getByTestId('web-site-header')).toBeVisible();
		await expect(
			page
				.getByRole('heading', { name: /handbook for working as an artist/i })
				.or(page.getByRole('heading', { name: /wiki temporarily unavailable/i }))
		).toBeVisible();
	});
});
