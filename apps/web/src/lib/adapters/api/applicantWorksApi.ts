import { apiFetch } from './client';
import type { AddWorkInput, ApplicantWork } from '$lib/core/domain/onboarding';
import type { ApplicantWorksPort } from '$lib/core/ports/applicantWorks';

const BASE = '/api/v1/me/application/works';

export class ApplicantWorksApi implements ApplicantWorksPort {
	list(): Promise<ApplicantWork[]> {
		return apiFetch<ApplicantWork[]>(BASE);
	}

	add(input: AddWorkInput): Promise<ApplicantWork> {
		return apiFetch<ApplicantWork>(BASE, {
			method: 'POST',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify(input)
		});
	}

	remove(id: string): Promise<void> {
		return apiFetch<void>(`${BASE}/${encodeURIComponent(id)}`, { method: 'DELETE' });
	}
}
