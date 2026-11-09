// Package provisioning creates the live artist (public profile and artworks) when an application
// is approved. It composes the profile and art repositories, so it owns no SQL of its own.
package provisioning

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/art"
	"github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/port/outbound"
)

// postNamespace makes an artwork's id a pure function of the work it came from, so provisioning
// the same application twice finds the artworks it already made instead of duplicating them.
var postNamespace = uuid.MustParse("6f1d3a52-9c1e-4d6f-a4c0-5b0e2f7a8c11")

type ArtistProvisioner struct {
	profiles outbound.ProfileRepository
	posts    outbound.ArtPostRepository
	now      func() time.Time
}

func NewArtistProvisioner(profiles outbound.ProfileRepository, posts outbound.ArtPostRepository) *ArtistProvisioner {
	return &ArtistProvisioner{profiles: profiles, posts: posts, now: func() time.Time { return time.Now().UTC() }}
}

// WorkPostID is the id of the artwork created from an applicant's work.
func WorkPostID(workID uuid.UUID) uuid.UUID {
	return uuid.NewSHA1(postNamespace, workID[:])
}

func (p *ArtistProvisioner) ProvisionApproved(ctx context.Context, app onboarding.OnboardingApplication) error {
	prof, err := p.profiles.GetArtistByUserID(ctx, app.ApplicantID)
	if errors.Is(err, apperrors.ErrNotFound) {
		prof, err = p.profiles.CreateDraftForUser(ctx, app.ApplicantID, app.DisplayName)
	}
	if err != nil {
		return err
	}

	now := p.now()
	changed := false
	if app.RequestedHandle != "" && prof.Handle != app.RequestedHandle {
		prof.Handle, prof.Slug = app.RequestedHandle, app.RequestedHandle
		changed = true
	}
	// An admin has reviewed this artist and their work, so the profile goes live with the approval.
	if prof.Status != profile.ProfileStatusApproved {
		prof.Status = profile.ProfileStatusApproved
		changed = true
	}
	if prof.ApprovedAt == nil {
		prof.ApprovedAt = &now
		changed = true
	}
	if changed {
		prof.UpdatedAt = now
		if prof, err = p.profiles.SaveArtist(ctx, *prof); err != nil {
			return err
		}
	}

	return p.PublishWorks(ctx, prof.ID, app.Works)
}

// PublishWorks creates a published artwork for each submitted work that doesn't have one yet. It
// never touches the profile, so it is also what repairs an artist who was approved before approval
// published their work (see cmd/backfill-artworks). Repeating it is harmless.
func (p *ArtistProvisioner) PublishWorks(ctx context.Context, artistID uuid.UUID, works []onboarding.ApplicantWork) error {
	now := p.now()
	for i, work := range works {
		id := WorkPostID(work.ID)
		if _, err := p.posts.GetByID(ctx, id); err == nil {
			continue
		} else if !errors.Is(err, apperrors.ErrNotFound) {
			return err
		}
		// The first work is the newest, so it leads the artist's page.
		publishedAt := now.Add(-time.Duration(i) * time.Millisecond)
		if _, err := p.posts.Create(ctx, art.ArtPost{
			ID: id, ArtistID: artistID, Title: work.Title, Medium: work.Medium, Year: work.Year,
			Status: art.ArtStatusPublished, PublishedAt: &publishedAt, CreatedAt: now, UpdatedAt: now,
			Media: []art.MediaAsset{{ID: uuid.New(), URL: work.URL, SortOrder: 0}},
		}); err != nil {
			return fmt.Errorf("create artwork for %q: %w", work.Title, err)
		}
	}
	return nil
}
