import type { AddWorkInput, ApplicantWork } from '$lib/core/domain/onboarding';

export interface ApplicantWorksPort {
	list(): Promise<ApplicantWork[]>;
	add(input: AddWorkInput): Promise<ApplicantWork>;
	remove(id: string): Promise<void>;
}
