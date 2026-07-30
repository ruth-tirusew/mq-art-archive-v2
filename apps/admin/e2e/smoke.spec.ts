import { test, expect } from '@playwright/test';

test.describe('@smoke login', () => {
	test('renders the login form', async ({ page }) => {
		await page.goto('/login');
		await expect(page.getByTestId('admin-login-form')).toBeVisible();
		await expect(page.getByLabel('Email')).toBeVisible();
		await expect(page.getByLabel('Password')).toBeVisible();
		await expect(page.getByRole('button', { name: 'Sign in' })).toBeVisible();
	});
});

test.describe('@smoke auth guard', () => {
	test('redirects unauthenticated users to login', async ({ page }) => {
		await page.goto('/artists');
		await expect(page).toHaveURL(/\/login/, { timeout: 15_000 });
		await expect(page.getByTestId('admin-login-form')).toBeVisible();
	});
});
