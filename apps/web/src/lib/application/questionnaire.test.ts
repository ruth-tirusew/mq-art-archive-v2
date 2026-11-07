import { describe, expect, it, vi } from 'vitest';
import { createQuestionnaireService } from '$lib/application/questionnaire';
import type { QuestionnairePort } from '$lib/core/ports/questionnaire';

function fakePort(overrides: Partial<QuestionnairePort> = {}): QuestionnairePort {
	return {
		listArtistAnswers: vi.fn().mockResolvedValue([]),
		getForm: vi.fn().mockResolvedValue(null),
		saveAnswers: vi.fn().mockResolvedValue(undefined),
		getNudge: vi.fn().mockResolvedValue(null),
		dismissNudge: vi.fn().mockResolvedValue(undefined),
		...overrides
	};
}

describe('questionnaireService over a fake port', () => {
	it('never lets a failed public-answers load break the profile page', async () => {
		const service = createQuestionnaireService(
			fakePort({ listArtistAnswers: vi.fn().mockRejectedValue(new Error('boom')) })
		);
		await expect(service.listArtistAnswers('ada')).resolves.toEqual([]);
	});

	it('treats a failed nudge load as no nudge', async () => {
		const service = createQuestionnaireService(fakePort({ getNudge: vi.fn().mockRejectedValue(new Error('boom')) }));
		await expect(service.getNudge()).resolves.toBeNull();
	});

	it('lets form and save failures through, so the apply flow can react to them', async () => {
		const service = createQuestionnaireService(
			fakePort({
				getForm: vi.fn().mockRejectedValue(new Error('down')),
				saveAnswers: vi.fn().mockRejectedValue(new Error('offline'))
			})
		);
		await expect(service.getForm()).rejects.toThrow('down');
		await expect(service.saveAnswers([])).rejects.toThrow('offline');
	});
});
