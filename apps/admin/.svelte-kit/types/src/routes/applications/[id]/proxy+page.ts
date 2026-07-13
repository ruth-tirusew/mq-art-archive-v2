// @ts-nocheck
import { onboardingService } from '$lib/application/onboarding';
import { error } from '@sveltejs/kit';
import type { PageLoad } from './$types';

export const load = async ({ params }: Parameters<PageLoad>[0]) => {
  try {
    const application = await onboardingService.getById(params.id);
    return { application };
  } catch {
    throw error(404, 'Application not found');
  }
};
