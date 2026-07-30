import { postsService } from '$lib/application/posts';
import { artistsService } from '$lib/application/artists';
import type { PageLoad } from './$types';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params }) => {
  try {
    const post = await postsService.getById(params.id);
    let artist = null;
    try {
      artist = await artistsService.getById(post.artist_id);
    } catch {
      // Artist may be missing; page still shows the post.
    }
    return { post, artist };
  } catch {
    error(404, 'Post not found');
  }
};
