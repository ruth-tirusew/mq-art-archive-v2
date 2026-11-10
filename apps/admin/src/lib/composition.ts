// Composition root for the admin onboarding screens: the one place that picks concrete adapters.
// Application services depend only on ports, so tests can build them over fakes instead.
import { InvitesApi } from '$lib/adapters/api/invitesApi';
import { OnboardingApi } from '$lib/adapters/api/onboardingApi';
import { QuestionnaireApi } from '$lib/adapters/api/questionnaireApi';
import { UsersApi } from '$lib/adapters/api/usersApi';
import { createInvitesService } from '$lib/application/invites';
import { createOnboardingService } from '$lib/application/onboarding';
import { createQuestionnaireService } from '$lib/application/questionnaire';
import { createUsersService } from '$lib/application/users';

export const onboardingService = createOnboardingService(new OnboardingApi());
export const invitesService = createInvitesService(new InvitesApi());
export const questionnaireService = createQuestionnaireService(new QuestionnaireApi());
export const usersService = createUsersService(new UsersApi());
