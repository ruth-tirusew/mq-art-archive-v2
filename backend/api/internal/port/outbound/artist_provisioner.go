package outbound

import (
	"context"

	"github.com/mq/api/internal/domain/onboarding"
)

// ArtistProvisioner turns an approved artist application into a live artist: a public profile with
// the handle they asked for, and a published artwork for each work they submitted. It must be safe
// to repeat, because a failure after it ran sends the application back to pending and the next
// approval runs it again.
type ArtistProvisioner interface {
	ProvisionApproved(ctx context.Context, app onboarding.OnboardingApplication) error
}
