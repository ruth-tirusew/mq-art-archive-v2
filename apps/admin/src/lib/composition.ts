// Composition root for the admin onboarding screens: the one place that picks concrete adapters.
// Application services depend only on ports, so tests can build them over fakes instead.
import { InvitesApi } from '$lib/adapters/api/invitesApi';
import { OnboardingApi } from '$lib/adapters/api/onboardingApi';
import { QuestionnaireApi } from '$lib/adapters/api/questionnaireApi';
import { StaffInvitesApi } from '$lib/adapters/api/staffInvitesApi';
import { UsersApi } from '$lib/adapters/api/usersApi';
import { createInvitesService } from '$lib/application/invites';
import { createOnboardingService } from '$lib/application/onboarding';
import { createQuestionnaireService } from '$lib/application/questionnaire';
import { createStaffInvitesService } from '$lib/application/staffInvites';
import { createUsersService } from '$lib/application/users';

export const onboardingService = createOnboardingService(new OnboardingApi());
export const invitesService = createInvitesService(new InvitesApi());
export const questionnaireService = createQuestionnaireService(new QuestionnaireApi());
export const usersService = createUsersService(new UsersApi());
export const staffInvitesService = createStaffInvitesService(new StaffInvitesApi());
