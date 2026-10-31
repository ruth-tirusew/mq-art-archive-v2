import { onboardingService } from '$lib/application/onboarding';
import { questionnaireService } from '$lib/application/questionnaire';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = ({ params }) =>
  requireAdmin(async () => {
    const [application, answers] = await Promise.all([
      onboardingService.getById(params.id),
      // Null when there are no questionnaire answers (e.g. applications from before it existed).
      questionnaireService.getApplicationAnswers(params.id)
    ]);
    return { application, answers };
  });
