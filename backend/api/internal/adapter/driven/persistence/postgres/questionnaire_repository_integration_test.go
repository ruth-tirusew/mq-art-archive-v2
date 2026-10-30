//go:build integration

package postgres_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/domain/questionnaire"
	"github.com/mq/api/internal/testutil/assist"
	"github.com/mq/api/internal/testutil/integration"
)

func draftWith(keys ...string) questionnaire.Revision {
	now := time.Now().UTC()
	rev := questionnaire.Revision{
		ID: uuid.New(), Status: questionnaire.RevisionDraft, ChangeKind: questionnaire.ChangeSubstantive,
		CreatedAt: now, UpdatedAt: now,
	}
	for i, k := range keys {
		rev.Questions = append(rev.Questions, questionnaire.Question{
			Key: k, Position: i + 1, Type: questionnaire.TypeLongText, PromptEN: "Tell us " + k,
			PromptAM: "ንገሩን", Required: i == 0, IsPublic: true, MaxLength: 500,
		})
	}
	return rev
}

func TestQuestionnaireRepository_revisionLifecycle(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewQuestionnaireRepository(pool)
	ctx := context.Background()

	_, err := repo.GetDraft(ctx)
	assist.ErrorIs(t, err, postgres.ErrNotFound)
	_, err = repo.GetLatestPublished(ctx)
	assist.ErrorIs(t, err, postgres.ErrNotFound)

	// Save, then replace the draft's questions wholesale, including a choice with options.
	draft := draftWith("bio", "influences")
	saved, err := repo.SaveDraft(ctx, draft)
	assist.NoError(t, err)
	assist.Len(t, 2, len(saved.Questions))
	assist.Equal(t, "ንገሩን", saved.Questions[0].PromptAM)

	draft.Questions = []questionnaire.Question{
		{Key: "bio", Position: 1, Type: questionnaire.TypeLongText, PromptEN: "Bio", Required: true, IsPublic: true, MaxLength: 500},
		{Key: "medium", Position: 2, Type: questionnaire.TypeChoice, PromptEN: "Medium", IsPublic: false, MaxLength: 50,
			Options: []questionnaire.Option{{Value: "oil", LabelEN: "Oil"}, {Value: "ink", LabelEN: "Ink", LabelAM: "ቀለም"}}},
	}
	saved, err = repo.SaveDraft(ctx, draft)
	assist.NoError(t, err)
	assist.Len(t, 2, len(saved.Questions))
	assist.Equal(t, "medium", saved.Questions[1].Key)
	assist.Len(t, 2, len(saved.Questions[1].Options))
	assist.Equal(t, "ቀለም", saved.Questions[1].Options[1].LabelAM)
	assist.Equal(t, false, saved.Questions[1].IsPublic)

	// Only one draft may exist at a time.
	_, err = repo.SaveDraft(ctx, draftWith("other"))
	assist.ErrorIs(t, err, apperrors.ErrConflict)

	// An invalid draft (duplicate key can't be stored, so use a choice with no options) cannot publish.
	bad := draftWith("x")
	bad.Questions[0].Type = questionnaire.TypeChoice
	bad.ID = draft.ID
	_, err = repo.SaveDraft(ctx, bad)
	assist.NoError(t, err)
	_, err = repo.Publish(ctx, draft.ID, time.Now().UTC())
	assist.ErrorIs(t, err, apperrors.ErrValidation)

	// Fix it and publish: number 1, frozen.
	_, err = repo.SaveDraft(ctx, draft)
	assist.NoError(t, err)
	pub, err := repo.Publish(ctx, draft.ID, time.Now().UTC())
	assist.NoError(t, err)
	assist.Equal(t, questionnaire.RevisionPublished, pub.Status)
	assist.NotNil(t, pub.Number)
	assist.Equal(t, 1, *pub.Number)

	_, err = repo.SaveDraft(ctx, draft)
	assist.ErrorIs(t, err, apperrors.ErrConflict) // immutable once published
	_, err = repo.Publish(ctx, draft.ID, time.Now().UTC())
	assist.ErrorIs(t, err, apperrors.ErrConflict)
	assist.ErrorIs(t, repo.DiscardDraft(ctx, draft.ID), postgres.ErrNotFound)

	// A second revision gets number 2 and becomes the latest.
	second := draftWith("bio", "new_q")
	_, err = repo.SaveDraft(ctx, second)
	assist.NoError(t, err)
	got, err := repo.GetDraft(ctx)
	assist.NoError(t, err)
	assist.Equal(t, second.ID, got.ID)
	pub2, err := repo.Publish(ctx, second.ID, time.Now().UTC())
	assist.NoError(t, err)
	assist.Equal(t, 2, *pub2.Number)

	latest, err := repo.GetLatestPublished(ctx)
	assist.NoError(t, err)
	assist.Equal(t, second.ID, latest.ID)
	assist.Len(t, 2, len(latest.Questions))

	list, err := repo.ListPublished(ctx)
	assist.NoError(t, err)
	assist.Len(t, 2, len(list))
	assist.Equal(t, 2, *list[0].Number)
	assist.Len(t, 0, len(list[0].Questions)) // summaries only

	// A fresh draft can be discarded.
	third := draftWith("tmp")
	_, err = repo.SaveDraft(ctx, third)
	assist.NoError(t, err)
	assist.NoError(t, repo.DiscardDraft(ctx, third.ID))
	_, err = repo.GetDraft(ctx)
	assist.ErrorIs(t, err, postgres.ErrNotFound)
}

func TestQuestionnaireAnswersAndNudges_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	revs := postgres.NewQuestionnaireRepository(pool)
	answers := postgres.NewQuestionnaireAnswerRepository(pool)
	nudges := postgres.NewQuestionnaireNudgeRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC()

	r1 := draftWith("bio", "influences")
	_, err := revs.SaveDraft(ctx, r1)
	assist.NoError(t, err)
	_, err = revs.Publish(ctx, r1.ID, now)
	assist.NoError(t, err)

	_, ada := integration.InsertArtistProfile(t, pool, "ada", "Ada")
	_, bez := integration.InsertArtistProfile(t, pool, "bez", "Bez")

	put := func(user uuid.UUID, rev uuid.UUID, key, body string) questionnaire.Answer {
		return questionnaire.Answer{UserID: user, QuestionKey: key, RevisionID: rev, Body: body,
			Format: questionnaire.FormatRich, Lang: questionnaire.LangAmharic, CreatedAt: now, UpdatedAt: now}
	}
	assist.NoError(t, answers.Upsert(ctx, []questionnaire.Answer{put(ada, r1.ID, "bio", "<p>ሰላም</p>"), put(ada, r1.ID, "influences", "x"), put(bez, r1.ID, "bio", "hi")}))

	// Upsert overwrites in place rather than duplicating.
	assist.NoError(t, answers.Upsert(ctx, []questionnaire.Answer{put(ada, r1.ID, "bio", "<p>updated</p>")}))
	got, err := answers.ListByUser(ctx, ada)
	assist.NoError(t, err)
	assist.Len(t, 2, len(got))
	assist.Equal(t, "<p>updated</p>", got[0].Body)
	assist.Equal(t, questionnaire.LangAmharic, got[0].Lang)

	both, err := answers.ListByUsers(ctx, []uuid.UUID{ada, bez})
	assist.NoError(t, err)
	assist.Len(t, 3, len(both))

	// An unknown revision fails the whole batch, leaving nothing half-written.
	err = answers.Upsert(ctx, []questionnaire.Answer{put(bez, r1.ID, "influences", "ok"), put(bez, uuid.New(), "zzz", "bad")})
	assist.Error(t, err)
	bezAnswers, _ := answers.ListByUser(ctx, bez)
	assist.Len(t, 1, len(bezAnswers))

	// Revision 2 adds a question. Ada answered everything she was asked; Bez is missing "influences".
	r2 := draftWith("bio", "influences", "new_q")
	_, err = revs.SaveDraft(ctx, r2)
	assist.NoError(t, err)
	_, err = revs.Publish(ctx, r2.ID, now)
	assist.NoError(t, err)

	// Migrations seed other approved artists, so only look at the two created here.
	mine := func() []uuid.UUID {
		all, err := nudges.ListUsersMissingAnswers(ctx, r2.ID)
		assist.NoError(t, err)
		var out []uuid.UUID
		for _, id := range all {
			if id == ada || id == bez {
				out = append(out, id)
			}
		}
		return out
	}
	assist.Len(t, 2, len(mine())) // both lack new_q

	assist.NoError(t, answers.Upsert(ctx, []questionnaire.Answer{put(ada, r2.ID, "new_q", "   ")}))
	assist.Len(t, 2, len(mine())) // whitespace-only is not an answer
	assist.NoError(t, answers.Upsert(ctx, []questionnaire.Answer{put(ada, r2.ID, "new_q", "real")}))
	missing := mine()
	assist.Len(t, 1, len(missing))
	assist.Equal(t, bez, missing[0])

	// Nudge: created once, never re-sent, dismissible.
	admin := uuid.New()
	integration.InsertUser(t, pool, admin, identity.RoleAdmin)
	created, err := nudges.Create(ctx, r2.ID, &admin, missing, now)
	assist.NoError(t, err)
	assist.Len(t, 1, len(created))
	again, err := nudges.Create(ctx, r2.ID, &admin, missing, now)
	assist.NoError(t, err)
	assist.Len(t, 0, len(again))

	n, err := nudges.GetActive(ctx, bez)
	assist.NoError(t, err)
	assist.Equal(t, r2.ID, n.RevisionID)
	_, err = nudges.GetActive(ctx, ada)
	assist.ErrorIs(t, err, postgres.ErrNotFound)

	assist.NoError(t, nudges.Dismiss(ctx, r2.ID, bez, now))
	_, err = nudges.GetActive(ctx, bez)
	assist.ErrorIs(t, err, postgres.ErrNotFound)
	assist.ErrorIs(t, nudges.Dismiss(ctx, r2.ID, ada, now), postgres.ErrNotFound)
}

func TestInviteRepository_integration(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	repo := postgres.NewInviteRepository(pool)
	apps := postgres.NewOnboardingRepository(pool)
	revs := postgres.NewQuestionnaireRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC()

	mk := func(hash, email, handle string, expires time.Time) onboarding.Invite {
		return onboarding.Invite{ID: uuid.New(), TokenHash: onboarding.HashInviteToken(hash), Email: email,
			DisplayName: "Ada", RequestedHandle: handle, ExpiresAt: expires, CreatedAt: now}
	}

	a, err := repo.Create(ctx, mk("t1", "ada@example.com", "ada_1", now.Add(time.Hour)))
	assist.NoError(t, err)

	// Handle is reserved case-insensitively; same token hash is also rejected.
	_, err = repo.Create(ctx, mk("t2", "b@example.com", "ADA_1", now.Add(time.Hour)))
	assist.ErrorIs(t, err, apperrors.ErrConflict)
	_, err = repo.Create(ctx, mk("t1", "c@example.com", "", now.Add(time.Hour)))
	assist.ErrorIs(t, err, apperrors.ErrConflict)

	byHash, err := repo.GetByTokenHash(ctx, onboarding.HashInviteToken("t1"))
	assist.NoError(t, err)
	assist.Equal(t, a.ID, byHash.ID)
	live, err := repo.GetLiveByHandle(ctx, "Ada_1", now)
	assist.NoError(t, err)
	assist.Equal(t, a.ID, live.ID)

	// Claim is single-use.
	user := uuid.New()
	integration.InsertUser(t, pool, user, identity.RolePublic)
	claimed, err := repo.Claim(ctx, a.ID, user, now)
	assist.NoError(t, err)
	assist.Equal(t, onboarding.InviteClaimed, claimed.Status(now))
	_, err = repo.Claim(ctx, a.ID, user, now)
	assist.ErrorIs(t, err, apperrors.ErrConflict)
	_, err = repo.Claim(ctx, uuid.New(), user, now)
	assist.ErrorIs(t, err, postgres.ErrNotFound)
	_, err = repo.GetLiveByHandle(ctx, "ada_1", now)
	assist.ErrorIs(t, err, postgres.ErrNotFound)
	assist.ErrorIs(t, repo.Revoke(ctx, a.ID, now), apperrors.ErrConflict) // cannot revoke a claimed invite

	// An expired invite frees its handle for a new invite, and cannot itself be claimed.
	expired, err := repo.Create(ctx, mk("t3", "d@example.com", "bez_2", now.Add(-time.Minute)))
	assist.NoError(t, err)
	_, err = repo.Claim(ctx, expired.ID, user, now)
	assist.ErrorIs(t, err, apperrors.ErrConflict)
	reuse, err := repo.Create(ctx, mk("t4", "e@example.com", "bez_2", now.Add(time.Hour)))
	assist.NoError(t, err)
	old, err := repo.GetByID(ctx, expired.ID)
	assist.NoError(t, err)
	assert := old.RevokedAt != nil
	assist.Equal(t, true, assert)

	// Revoke frees a live handle too.
	assist.NoError(t, repo.Revoke(ctx, reuse.ID, now))
	_, err = repo.Create(ctx, mk("t5", "f@example.com", "bez_2", now.Add(time.Hour)))
	assist.NoError(t, err)

	all, err := repo.List(ctx)
	assist.NoError(t, err)
	assist.Len(t, 4, len(all))

	// The application round-trips its revision and invite links.
	rev := draftWith("bio")
	_, err = revs.SaveDraft(ctx, rev)
	assist.NoError(t, err)
	app, err := apps.Save(ctx, onboarding.OnboardingApplication{
		ID: uuid.New(), ApplicantID: user, ApplicantType: onboarding.ApplicantTypeArtist, DisplayName: "Ada",
		Status: onboarding.ApprovalStatusPending, QuestionnaireRevisionID: &rev.ID, InviteID: &a.ID, CreatedAt: now, UpdatedAt: now,
	})
	assist.NoError(t, err)
	got, err := apps.GetByID(ctx, app.ID)
	assist.NoError(t, err)
	assist.NotNil(t, got.QuestionnaireRevisionID)
	assist.Equal(t, rev.ID, *got.QuestionnaireRevisionID)
	assist.Equal(t, a.ID, *got.InviteID)
}

func TestQuestionnaireAnswerRepository_isLockedForReview(t *testing.T) {
	pool := integration.SetupPostgresPool(t)
	answers := postgres.NewQuestionnaireAnswerRepository(pool)
	apps := postgres.NewOnboardingRepository(pool)
	ctx := context.Background()
	now := time.Now().UTC()

	user := uuid.New()
	integration.InsertUser(t, pool, user, identity.RolePublic)

	locked, err := answers.IsLockedForReview(ctx, user)
	assist.NoError(t, err)
	assist.Equal(t, false, locked) // never applied

	app, err := apps.Save(ctx, onboarding.OnboardingApplication{
		ID: uuid.New(), ApplicantID: user, ApplicantType: onboarding.ApplicantTypeArtist, DisplayName: "Ada",
		Status: onboarding.ApprovalStatusPending, CreatedAt: now, UpdatedAt: now,
	})
	assist.NoError(t, err)
	locked, _ = answers.IsLockedForReview(ctx, user)
	assist.Equal(t, true, locked) // pending: frozen for the reviewer

	app.Status = onboarding.ApprovalStatusRejected
	_, err = apps.Save(ctx, *app)
	assist.NoError(t, err)
	locked, _ = answers.IsLockedForReview(ctx, user)
	assist.Equal(t, false, locked) // rejected: they may edit and resubmit

	// Only the LATEST application counts.
	_, err = apps.Save(ctx, onboarding.OnboardingApplication{
		ID: uuid.New(), ApplicantID: user, ApplicantType: onboarding.ApplicantTypeArtist, DisplayName: "Ada",
		Status: onboarding.ApprovalStatusPending, CreatedAt: now.Add(time.Minute), UpdatedAt: now.Add(time.Minute),
	})
	assist.NoError(t, err)
	locked, _ = answers.IsLockedForReview(ctx, user)
	assist.Equal(t, true, locked)
}
