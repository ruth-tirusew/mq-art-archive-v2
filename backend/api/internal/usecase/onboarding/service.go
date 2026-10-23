package onboarding

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/onboarding"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	applications outbound.OnboardingRepository
	users        outbound.UserRepository
	profiles     outbound.ProfileRepository
	institutions outbound.InstitutionRepository
	mailer       outbound.Mailer
	webAppURL    string
}

// Option customises optional collaborators of the onboarding Service.
type Option func(*Service)

// WithMailer enables best-effort decision emails to applicants.
func WithMailer(mailer outbound.Mailer, webAppURL string) Option {
	return func(s *Service) {
		s.mailer = mailer
		s.webAppURL = strings.TrimRight(webAppURL, "/")
	}
}

func NewService(
	applications outbound.OnboardingRepository,
	users outbound.UserRepository,
	profiles outbound.ProfileRepository,
	institutions outbound.InstitutionRepository,
	opts ...Option,
) inbound.OnboardingService {
	s := &Service{
		applications: applications,
		users:        users,
		profiles:     profiles,
		institutions: institutions,
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) ListPending(ctx context.Context) ([]domain.OnboardingApplication, error) {
	return s.applications.ListPending(ctx)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
	return s.applications.GetByID(ctx, id)
}

func (s *Service) GetMyApplication(ctx context.Context, applicantID uuid.UUID) (*domain.OnboardingApplication, error) {
	return s.applications.GetLatestByApplicantID(ctx, applicantID)
}

func (s *Service) Submit(ctx context.Context, applicantID uuid.UUID, applicantType domain.ApplicantType, displayName, notes string) (*domain.OnboardingApplication, error) {
	handle := ""
	if applicantType == domain.ApplicantTypeArtist {
		handle = strings.ReplaceAll(strings.ToLower(strings.TrimSpace(displayName)), " ", "_")
	}
	return s.SubmitWithHandle(ctx, applicantID, applicantType, displayName, handle, notes)
}

func (s *Service) SubmitWithHandle(
	ctx context.Context,
	applicantID uuid.UUID,
	applicantType domain.ApplicantType,
	displayName, requestedHandle, notes string,
) (*domain.OnboardingApplication, error) {
	displayName = strings.TrimSpace(displayName)
	if displayName == "" {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, "display_name is required")
	}
	if applicantType != domain.ApplicantTypeArtist && applicantType != domain.ApplicantTypeInstitution {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, "invalid applicant_type")
	}
	requestedHandle = strings.ToLower(strings.TrimSpace(requestedHandle))
	if applicantType == domain.ApplicantTypeArtist {
		if !validHandle(requestedHandle) {
			return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, "requested_handle must be 3-30 lowercase letters, numbers, or underscores")
		}
	} else {
		requestedHandle = ""
	}

	if _, err := s.profiles.GetArtistByUserID(ctx, applicantID); err == nil {
		return nil, apperrors.ErrConflict
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}

	existing, err := s.applications.GetLatestByApplicantID(ctx, applicantID)
	if err != nil && !errors.Is(err, apperrors.ErrNotFound) {
		return nil, err
	}
	if existing != nil && (existing.Status == domain.ApprovalStatusPending || existing.Status == domain.ApprovalStatusApproved) {
		return nil, apperrors.ErrConflict
	}
	if applicantType == domain.ApplicantTypeArtist {
		available, err := s.CheckHandleAvailable(ctx, requestedHandle)
		if err != nil {
			return nil, err
		}
		if !available {
			return nil, apperrors.ErrConflict
		}
	}

	now := time.Now().UTC()
	app := domain.OnboardingApplication{
		ID:              uuid.New(),
		ApplicantID:     applicantID,
		ApplicantType:   applicantType,
		DisplayName:     displayName,
		RequestedHandle: requestedHandle,
		Notes:           notes,
		Status:          domain.ApprovalStatusPending,
		CreatedAt:       now,
		UpdatedAt:       now,
	}
	return s.applications.Save(ctx, app)
}

func (s *Service) Review(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, status domain.ApprovalStatus, notes string) (*domain.OnboardingApplication, error) {
	app, err := s.applications.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	reviewNotes := strings.TrimSpace(notes)
	if status == domain.ApprovalStatusRejected && reviewNotes == "" {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, "a reason is required when rejecting an application")
	}

	now := time.Now().UTC()
	app.Status = status
	app.ReviewNotes = reviewNotes
	app.ReviewedBy = &reviewerID
	app.ReviewedAt = &now
	app.UpdatedAt = now

	if status == domain.ApprovalStatusApproved {
		switch app.ApplicantType {
		case domain.ApplicantTypeArtist:
			if err := s.provisionArtist(ctx, app); err != nil {
				return nil, err
			}
		case domain.ApplicantTypeInstitution:
			if err := s.provisionInstitution(ctx, app); err != nil {
				return nil, err
			}
		}
	}

	saved, err := s.applications.Save(ctx, *app)
	if err != nil {
		return nil, err
	}
	s.notifyDecision(ctx, saved)
	return saved, nil
}

// notifyDecision emails the applicant about the outcome. Failures are ignored: the
// review itself has already been saved and must not fail because of a mail outage.
func (s *Service) notifyDecision(ctx context.Context, app *domain.OnboardingApplication) {
	if s.mailer == nil {
		return
	}
	user, err := s.users.GetByID(ctx, app.ApplicantID)
	if err != nil || user == nil || user.Email == "" {
		return
	}
	subject, body := decisionEmail(app, s.webAppURL)
	_ = s.mailer.Send(ctx, user.Email, subject, body)
}

func decisionEmail(app *domain.OnboardingApplication, webAppURL string) (string, string) {
	var b strings.Builder
	fmt.Fprintf(&b, "Hi %s,\n\n", app.DisplayName)
	if app.Status == domain.ApprovalStatusApproved {
		b.WriteString("Good news: your Artiv application was approved.\n\n")
		if app.ReviewNotes != "" {
			fmt.Fprintf(&b, "A note from the team:\n%s\n\n", app.ReviewNotes)
		}
		fmt.Fprintf(&b, "Next, open your studio, add a portrait and a short bio, and publish your profile:\n%s/studio\n", webAppURL)
		return "Welcome to Artiv — your application was approved", b.String()
	}
	b.WriteString("Thanks for applying to Artiv. We couldn't approve your application this time.\n\n")
	if app.ReviewNotes != "" {
		fmt.Fprintf(&b, "Reason:\n%s\n\n", app.ReviewNotes)
	}
	fmt.Fprintf(&b, "You can update your details and apply again:\n%s/apply\n", webAppURL)
	return "Update on your Artiv application", b.String()
}

func (s *Service) provisionArtist(ctx context.Context, app *domain.OnboardingApplication) error {
	if err := s.users.PromoteToArtist(ctx, app.ApplicantID); err != nil {
		return err
	}

	if _, err := s.profiles.GetArtistByUserID(ctx, app.ApplicantID); err == nil {
		return nil
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return err
	}

	draft, err := s.profiles.CreateDraftForUser(ctx, app.ApplicantID, app.DisplayName)
	if err != nil {
		return err
	}
	if app.RequestedHandle != "" {
		draft.Handle = app.RequestedHandle
		draft.Slug = app.RequestedHandle
		_, err = s.profiles.SaveArtist(ctx, *draft)
	}
	return err
}

func (s *Service) provisionInstitution(ctx context.Context, app *domain.OnboardingApplication) error {
	if err := s.users.PromoteToInstitution(ctx, app.ApplicantID); err != nil {
		return err
	}

	if _, err := s.institutions.GetByUserID(ctx, app.ApplicantID); err == nil {
		return nil
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return err
	}

	_, err := s.institutions.CreateDraftForUser(ctx, app.ApplicantID, app.DisplayName)
	return err
}

var handlePattern = regexp.MustCompile(`^[a-z0-9_]{3,30}$`)

func validHandle(handle string) bool { return handlePattern.MatchString(handle) }

func (s *Service) CheckHandleAvailable(ctx context.Context, handle string) (bool, error) {
	handle = strings.ToLower(strings.TrimSpace(handle))
	if !validHandle(handle) {
		return false, nil
	}
	if _, err := s.profiles.GetArtistByHandle(ctx, handle); err == nil {
		return false, nil
	} else if !errors.Is(err, apperrors.ErrNotFound) {
		return false, err
	}
	pending, err := s.applications.ListPending(ctx)
	if err != nil {
		return false, err
	}
	for _, app := range pending {
		if strings.EqualFold(app.RequestedHandle, handle) {
			return false, nil
		}
	}
	return true, nil
}
