import { postsService } from '$lib/application/posts';
import { artistsService } from '$lib/application/artists';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) =>
  requireAdmin(async () => {
    const post = await postsService.getById(params.id);
    // The artist may have been removed; the page still shows the post.
    const artist = await artistsService.getById(post.artist_id).catch(() => null);
    return { post, artist };
  });
