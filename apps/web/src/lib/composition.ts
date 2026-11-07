// Composition root for the applicant onboarding flow: the one place that picks concrete adapters.
// Application services depend only on ports, so tests can build them over fakes instead.
import { InviteApi } from '$lib/adapters/api/inviteApi';
import { OnboardingApi } from '$lib/adapters/api/onboardingApi';
import { QuestionnaireApi } from '$lib/adapters/api/questionnaireApi';
import { createInviteService } from '$lib/application/invites';
import { createOnboardingService } from '$lib/application/onboarding';
import { createQuestionnaireService } from '$lib/application/questionnaire';

export const onboardingService = createOnboardingService(new OnboardingApi());
export const inviteService = createInviteService(new InviteApi());
export const questionnaireService = createQuestionnaireService(new QuestionnaireApi());
