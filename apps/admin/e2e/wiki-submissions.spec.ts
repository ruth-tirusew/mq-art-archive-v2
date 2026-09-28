import { test, expect, type APIRequestContext } from '@playwright/test';

// Exercises the admin wiki-submissions review queue against a real backend, using
// dev-mode auth (AUTH_DEV_MODE=true, the local default): the backend trusts an
// X-User-ID header for a seeded user when no session cookie is present. See:
// backend/api/internal/adapter/driving/http/middleware/auth.go and
// backend/api/migrations/00009_seed_dev_admin.sql / 00010_seed_featured_artist.sql.
const API_URL = process.env.PUBLIC_API_URL ?? 'http://localhost:8080';
const ARTIST_ID = '33333333-3333-3333-3333-333333333333';
const ADMIN_ID = '00000000-0000-4000-8000-000000000001';

function unique(label: string) {
	return `${label} ${Date.now()}-${Math.floor(Math.random() * 1e6)}`;
}

async function apiPost<T>(
	request: APIRequestContext,
	path: string,
	userID: string,
	data: Record<string, unknown>
): Promise<T> {
	const res = await request.post(`${API_URL}${path}`, {
		headers: { 'X-User-ID': userID, 'Content-Type': 'application/json' },
		data
	});
	expect(res.ok(), `POST ${path} failed: ${res.status()} ${await res.text()}`).toBeTruthy();
	return res.json();
}

test.describe('@wiki admin queue', () => {
	test('admin can view a diff for a pending edit submission and approve it', async ({
		page,
		context,
		request
	}) => {
		const title = unique('E2E admin-reviewed article');
		const originalBody = 'First line unchanged.\nSecond line to delete.';
		const editedBody = 'First line unchanged.\nThird line inserted.';

		// Fixture: submit + approve a new article as the seeded artist (making them its
		// author), then submit an edit to it — this is the pending edit the admin will
		// review through the UI below.
		const created = await apiPost<{ id: string }>(
			request,
			'/api/v1/me/wiki/submissions',
			ARTIST_ID,
			{ title, body: originalBody }
		);
		const approved = await apiPost<{ article_id: string }>(
			request,
			`/admin/v1/wiki/submissions/${created.id}/approve`,
			ADMIN_ID,
			{}
		);
		const articleId = approved.article_id;
		expect(articleId).toBeTruthy();

		await apiPost(request, '/api/v1/me/wiki/submissions', ARTIST_ID, {
			article_id: articleId,
			title,
			body: editedBody
		});

		await context.setExtraHTTPHeaders({ 'X-User-ID': ADMIN_ID });
		await page.goto('/wiki/submissions');

		const submissionCard = page
			.getByTestId('admin-wiki-submission-item')
			.filter({ hasText: title });
		await expect(submissionCard).toBeVisible();

		await submissionCard.getByRole('button', { name: /show diff/i }).click();
		const diff = submissionCard.getByTestId('admin-wiki-diff');
		await expect(diff).toBeVisible();
		await expect(diff.getByText('Second line to delete.')).toBeVisible();
		await expect(diff.getByText('Third line inserted.')).toBeVisible();

		await submissionCard.getByRole('button', { name: 'Approve' }).click();
		await expect(submissionCard).not.toBeVisible();

		const articleRes = await request.get(`${API_URL}/admin/v1/articles/${articleId}`, {
			headers: { 'X-User-ID': ADMIN_ID }
		});
		expect(articleRes.ok()).toBeTruthy();
		const article = await articleRes.json();
		expect(article.body).toBe(editedBody);
	});

	test('admin can reject a submission with review notes', async ({ page, context, request }) => {
		const title = unique('E2E rejected article');
		await apiPost(request, '/api/v1/me/wiki/submissions', ARTIST_ID, {
			title,
			body: 'not going to be published'
		});

		await context.setExtraHTTPHeaders({ 'X-User-ID': ADMIN_ID });
		await page.goto('/wiki/submissions');

		const submissionCard = page
			.getByTestId('admin-wiki-submission-item')
			.filter({ hasText: title });
		await expect(submissionCard).toBeVisible();

		await submissionCard.getByPlaceholder('Optional review notes').fill('Not a good fit yet.');
		await submissionCard.getByRole('button', { name: 'Reject' }).click();
		await expect(submissionCard).not.toBeVisible();
	});
});
