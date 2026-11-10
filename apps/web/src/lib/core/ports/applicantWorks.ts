import type { AddWorkInput, ApplicantPortrait, ApplicantWork } from '$lib/core/domain/onboarding';

export interface ApplicantWorksPort {
	list(): Promise<ApplicantWork[]>;
	add(input: AddWorkInput): Promise<ApplicantWork>;
	remove(id: string): Promise<void>;
	/** The applicant's portrait, or null if they have not added one. */
	getPortrait(): Promise<ApplicantPortrait | null>;
	setPortrait(mediaAssetId: string): Promise<ApplicantPortrait>;
	removePortrait(): Promise<void>;
}
