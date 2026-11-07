import { describe, expect, it } from 'vitest';
import { http, HttpResponse } from 'msw';
import { PUBLIC_API_URL } from '$env/static/public';
import { server } from '../../../test/msw/server';
import { UsersApi } from '$lib/adapters/api/usersApi';

const base = PUBLIC_API_URL.replace(/\/$/, '');

describe('UsersApi (MSW)', () => {
  const api = new UsersApi();

  it('pages with limit and offset', async () => {
    let query = '';
    server.use(
      http.get(`${base}/admin/v1/users`, ({ request }) => {
        query = new URL(request.url).search;
        return HttpResponse.json([{ id: 'u-1', email: 'a@b.co', role: 'artist' }]);
      })
    );
    const users = await api.list({ limit: 200, offset: 400 });
    expect(query).toBe('?limit=200&offset=400');
    expect(users).toEqual([{ id: 'u-1', email: 'a@b.co', role: 'artist' }]);
  });

  it('patches the role and encodes the id', async () => {
    let path = '';
    let body: unknown;
    server.use(
      http.patch(`${base}/admin/v1/users/:id`, async ({ request }) => {
        path = new URL(request.url).pathname;
        body = await request.json();
        return HttpResponse.json({ id: 'a/b', email: 'a@b.co', role: 'contributor' });
      })
    );
    const updated = await api.changeRole('a/b', 'contributor');
    expect(path).toBe('/admin/v1/users/a%2Fb');
    expect(body).toEqual({ role: 'contributor' });
    expect(updated.role).toBe('contributor');
  });
});
