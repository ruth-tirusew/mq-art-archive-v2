import { test, expect, type APIRequestContext } from '@playwright/test';

// These tests exercise the wiki-collaboration flow (submit → admin approve → propose an
// edit → owning-artist review with a diff) end to end against a real backend. They rely
// on dev-mode auth (AUTH_DEV_MODE=true, the local default): the backend trusts an
// X-User-ID header for a seeded user when no session cookie is present, so tests can act
// as a specific role without real OAuth or password login. See:
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

test.describe('@wiki artist submission', () => {
	test('artist can submit a new article via the UI and it shows as pending', async ({
		page,
		context
	}) => {
		await context.setExtraHTTPHeaders({ 'X-User-ID': ARTIST_ID });
		const title = unique('E2E new article');

		await page.goto('/studio/wiki/new');
		await page.getByLabel('Title').fill(title);
		await page.getByLabel('Article body').fill('First paragraph.\n\nSecond paragraph.');
		await page.getByRole('button', { name: /submit for review/i }).click();

		await expect(page).toHaveURL(/\/studio\/wiki$/);
		const item = page.getByTestId('web-wiki-submission-item').filter({ hasText: title });
		await expect(item).toBeVisible();
		await expect(item.getByText('pending')).toBeVisible();
	});
});

test.describe('@wiki artist self-review', () => {
	test('artist can review and approve an edit to their own article, seeing a diff', async ({
		page,
		context,
		request
	}) => {
		const originalTitle = unique('E2E owned article');
		const originalBody = 'Line one stays.\nLine two will be removed.\nLine three stays.';
		const editedBody = 'Line one stays.\nLine three stays.\nA brand new line is added.';

		// Fixture setup via API: submit + admin-approve a new article (this is what makes
		// the submitter its author, per the Phase 0 ownership fix), then submit an edit to
		// it as the same artist — there is no UI yet to propose an edit to an existing
		// article (that's Phase 4), so the edit submission is created directly via the API.
		const created = await apiPost<{ id: string }>(
			request,
			'/api/v1/me/wiki/submissions',
			ARTIST_ID,
			{ title: originalTitle, body: originalBody }
		);
		const approved = await apiPost<{ article_id: string }>(
			request,
			`/admin/v1/wiki/submissions/${created.id}/approve`,
			ADMIN_ID,
			{}
		);
		const articleId = approved.article_id;
		expect(articleId).toBeTruthy();

		const editSubmission = await apiPost<{ id: string }>(
			request,
			'/api/v1/me/wiki/submissions',
			ARTIST_ID,
			{ article_id: articleId, title: originalTitle, body: editedBody }
		);

		await context.setExtraHTTPHeaders({ 'X-User-ID': ARTIST_ID });

		await page.goto('/studio/wiki');
		await expect(page.getByRole('heading', { name: 'Awaiting your review' })).toBeVisible();
		const reviewItem = page
			.getByTestId('web-wiki-review-item')
			.filter({ hasText: originalTitle });
		await expect(reviewItem).toBeVisible();
		await reviewItem.click();

		await expect(page).toHaveURL(new RegExp(`/studio/wiki/${editSubmission.id}$`));
		const diff = page.getByTestId('web-wiki-diff');
		await expect(diff).toBeVisible();
		await expect(diff.getByText('Line two will be removed.')).toBeVisible();
		await expect(diff.getByText('A brand new line is added.')).toBeVisible();
		// Unchanged lines render too, not just the delta.
		await expect(diff.getByText('Line one stays.')).toBeVisible();

		await page.getByRole('button', { name: 'Approve' }).click();
		await expect(page).toHaveURL(/\/studio\/wiki$/);

		// Verify the approval actually persisted the edit (not just navigated away).
		const articleRes = await request.get(`${API_URL}/admin/v1/articles/${articleId}`, {
			headers: { 'X-User-ID': ADMIN_ID }
		});
		expect(articleRes.ok()).toBeTruthy();
		const article = await articleRes.json();
		expect(article.body).toBe(editedBody);
		expect(article.version).toBe(2);
	});

	test('artist reviewing a stale edit sees a warning before approving', async ({
		page,
		context,
		request
	}) => {
		const title = unique('E2E stale article');
		const created = await apiPost<{ id: string }>(
			request,
			'/api/v1/me/wiki/submissions',
			ARTIST_ID,
			{ title, body: 'original body' }
		);
		const approved = await apiPost<{ article_id: string }>(
			request,
			`/admin/v1/wiki/submissions/${created.id}/approve`,
			ADMIN_ID,
			{}
		);
		const articleId = approved.article_id;

		// Submission based on version 1...
		const editSubmission = await apiPost<{ id: string }>(
			request,
			'/api/v1/me/wiki/submissions',
			ARTIST_ID,
			{ article_id: articleId, title, body: 'edited from v1' }
		);

		// ...but the article moves to version 2 via a direct admin edit before review happens.
		const patchRes = await request.put(`${API_URL}/admin/v1/articles/${articleId}`, {
			headers: { 'X-User-ID': ADMIN_ID, 'Content-Type': 'application/json' },
			data: { title, body: 'admin changed this first' }
		});
		expect(patchRes.ok()).toBeTruthy();

		await context.setExtraHTTPHeaders({ 'X-User-ID': ARTIST_ID });
		await page.goto(`/studio/wiki/${editSubmission.id}`);

		await expect(page.getByText(/changed since the submission was written/i)).toBeVisible();

		// Approving should fail (optimistic-concurrency conflict), not silently overwrite.
		await page.getByRole('button', { name: 'Approve' }).click();
		await expect(page.getByRole('alert')).toBeVisible();
		await expect(page).toHaveURL(new RegExp(`/studio/wiki/${editSubmission.id}$`));

		const articleRes = await request.get(`${API_URL}/admin/v1/articles/${articleId}`, {
			headers: { 'X-User-ID': ADMIN_ID }
		});
		const article = await articleRes.json();
		expect(article.body).toBe('admin changed this first');
	});
});
