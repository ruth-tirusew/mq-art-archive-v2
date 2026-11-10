import { ApiError, apiFetch } from './client';
import type { AddWorkInput, ApplicantPortrait, ApplicantWork } from '$lib/core/domain/onboarding';
import type { ApplicantWorksPort } from '$lib/core/ports/applicantWorks';

const BASE = '/api/v1/me/application/works';
const PORTRAIT = '/api/v1/me/application/portrait';

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

	async getPortrait(): Promise<ApplicantPortrait | null> {
		try {
			return await apiFetch<ApplicantPortrait>(PORTRAIT);
		} catch (e) {
			if (e instanceof ApiError && e.status === 404) return null;
			throw e;
		}
	}

	setPortrait(mediaAssetId: string): Promise<ApplicantPortrait> {
		return apiFetch<ApplicantPortrait>(PORTRAIT, {
			method: 'PUT',
			headers: { 'Content-Type': 'application/json' },
			body: JSON.stringify({ media_asset_id: mediaAssetId })
		});
	}

	removePortrait(): Promise<void> {
		return apiFetch<void>(PORTRAIT, { method: 'DELETE' });
	}
}
