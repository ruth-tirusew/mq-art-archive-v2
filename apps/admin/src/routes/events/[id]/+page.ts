import { eventsService } from '$lib/application/events';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load: PageLoad = async ({ params }) => {
  try {
    const event = await eventsService.getById(params.id);
    return { event };
  } catch {
    throw error(404, 'Event not found');
  }
};
