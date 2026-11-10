import { uploadImage } from '$lib/application/media';
import type { UploadedMedia } from '$lib/core/domain/media';
import type { AddWorkInput, ApplicantPortrait, ApplicantWork, WorkDraft } from '$lib/core/domain/onboarding';
import { MAX_APPLICATION_WORKS } from '$lib/core/domain/onboarding';
import type { ApplicantWorksPort } from '$lib/core/ports/applicantWorks';

type Upload = (file: File, options: { scope: 'applicant'; signal?: AbortSignal }) => Promise<UploadedMedia>;

export function createApplicantWorksService(port: ApplicantWorksPort, upload: Upload = uploadImage) {
	return {
		/** Works already stored for the applicant (only a re-applicant has any). */
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
		/** The portrait already stored for the applicant (only a re-applicant has one). */
		getPortrait(): Promise<ApplicantPortrait | null> {
			return port.getPortrait();
		},
		removePortrait(): Promise<void> {
			return port.removePortrait();
		},
		/**
		 * Uploads the chosen portrait and makes it the applicant's portrait, replacing any earlier
		 * one. Like the works, this runs only when the applicant submits.
		 */
		async uploadPortrait(file: File, options: { signal?: AbortSignal } = {}): Promise<ApplicantPortrait> {
			const media = await upload(file, { scope: 'applicant', signal: options.signal });
			if (!media.id) throw new Error('The upload did not return an image id.');
			return port.setPortrait(media.id);
		},
		/**
		 * Uploads each draft that isn't stored yet, in order, and attaches it as a work. This is the
		 * only place images go to Cloudinary, and it runs when the applicant submits. A draft that
		 * is already done is skipped, so a retry after a failure only does the remaining ones. It
		 * stops at the first failure (marking that draft) so the applicant can fix it or try again.
		 */
		async uploadDrafts(
			drafts: WorkDraft[],
			options: { signal?: AbortSignal; onProgress?: (done: number, total: number) => void } = {}
		): Promise<void> {
			const todo = drafts.filter((d) => d.status !== 'done');
			let done = 0;
			options.onProgress?.(0, todo.length);
			for (const draft of todo) {
				draft.status = 'uploading';
				draft.error = undefined;
				try {
					const media = await upload(draft.file, { scope: 'applicant', signal: options.signal });
					if (!media.id) throw new Error('The upload did not return an image id.');
					const work = await this.add({
						media_asset_id: media.id,
						title: draft.title,
						medium: draft.medium,
						year: draft.year.trim() ? Number(draft.year) : null
					});
					draft.workId = work.id;
					draft.status = 'done';
					done++;
					options.onProgress?.(done, todo.length);
				} catch (cause) {
					draft.status = 'failed';
					draft.error = cause instanceof Error ? cause.message : 'Upload failed.';
					throw cause;
				}
			}
		},
		maxWorks: MAX_APPLICATION_WORKS
	};
}
