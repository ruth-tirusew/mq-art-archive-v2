import { artistsService } from '$lib/application/artists';
import type { PageLoad } from './$types';
import { error } from '@sveltejs/kit';

export const load: PageLoad = async ({ params }) => {
  try {
    const artist = await artistsService.getById(params.id);
    return { artist };
  } catch {
    error(404, 'Artist not found');
  }
};
