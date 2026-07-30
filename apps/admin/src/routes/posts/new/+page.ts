import { artistsService } from '$lib/application/artists';
import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
  const artists = await artistsService.list();
  return { artists };
};
