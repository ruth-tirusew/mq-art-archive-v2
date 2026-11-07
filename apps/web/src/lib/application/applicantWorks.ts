import type { AddWorkInput, ApplicantWork } from '$lib/core/domain/onboarding';
import { MAX_APPLICATION_WORKS } from '$lib/core/domain/onboarding';
import type { ApplicantWorksPort } from '$lib/core/ports/applicantWorks';

export function createApplicantWorksService(port: ApplicantWorksPort) {
	return {
		list(): Promise<ApplicantWork[]> {
			return port.list();
		},
		/** Trims what was typed and refuses an empty title before bothering the server. */
		add(input: AddWorkInput): Promise<ApplicantWork> {
			const title = input.title.trim();
			if (!title) return Promise.reject(new Error('Give this work a title.'));
			return port.add({
				media_asset_id: input.media_asset_id,
				title,
				medium: input.medium?.trim() || undefined,
				year: input.year ?? null
			});
		},
		remove(id: string): Promise<void> {
			return port.remove(id);
		},
		maxWorks: MAX_APPLICATION_WORKS
	};
}
