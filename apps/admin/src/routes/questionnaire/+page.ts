import { questionnaireService } from '$lib/application/questionnaire';
import { requireAdmin } from '$lib/utils/loadGuard';
import type { PageLoad } from './$types';

export const load: PageLoad = () =>
  requireAdmin(async () => {
    const [draft, revisions, nudge] = await Promise.all([
      questionnaireService.getDraft(),
      questionnaireService.listRevisions(),
      questionnaireService.previewNudge()
    ]);
    // The list omits questions; the newest revision is fetched in full as the baseline for "what changed".
    const latest = revisions[0] ? await questionnaireService.getRevision(revisions[0].id) : null;
    return { draft, revisions, latest, nudge };
  });
