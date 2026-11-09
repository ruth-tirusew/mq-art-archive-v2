// Composition root for the applicant onboarding flow: the one place that picks concrete adapters.
// Application services depend only on ports, so tests can build them over fakes instead.
import { ApplicantWorksApi } from '$lib/adapters/api/applicantWorksApi';
import { InviteApi } from '$lib/adapters/api/inviteApi';
import { OnboardingApi } from '$lib/adapters/api/onboardingApi';
import { QuestionnaireApi } from '$lib/adapters/api/questionnaireApi';
import { createApplicantWorksService } from '$lib/application/applicantWorks';
import { createInviteService } from '$lib/application/invites';
import { createOnboardingService } from '$lib/application/onboarding';
import { createQuestionnaireService } from '$lib/application/questionnaire';

export const onboardingService = createOnboardingService(new OnboardingApi());
export const inviteService = createInviteService(new InviteApi());
export const questionnaireService = createQuestionnaireService(new QuestionnaireApi());
export const applicantWorksService = createApplicantWorksService(new ApplicantWorksApi());

// Chosen-but-not-uploaded works, kept in this browser so a refresh doesn't lose them.
export { workDraftStore } from '$lib/adapters/browser/workDraftStore';
