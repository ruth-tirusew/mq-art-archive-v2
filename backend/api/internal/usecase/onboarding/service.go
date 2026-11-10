package onboarding

import (
	"context"
	"errors"
	"fmt"
	"log"
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
	applications  outbound.OnboardingRepository
	users         outbound.UserRepository
	profiles      outbound.ProfileRepository
	institutions  outbound.InstitutionRepository
	mailer        outbound.Mailer
	webAppURL     string
	invites       outbound.InviteRepository
	questionnaire inbound.ApplicationQuestionnaireGate
	registrar     inbound.InvitedRegistrar
	works         outbound.ApplicantWorkRepository
	provisioner   outbound.ArtistProvisioner
	now           func() time.Time
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

// WithInvites enables artist invite links, and makes handle checks respect handles that
// live invites have reserved.
func WithInvites(invites outbound.InviteRepository) Option {
	return func(s *Service) { s.invites = invites }
}

// WithInvitedRegistrar enables creating an invitee's account straight from their invite.
func WithInvitedRegistrar(r inbound.InvitedRegistrar) Option {
	return func(s *Service) { s.registrar = r }
}

// WithQuestionnaire requires artist applicants to have completed the published questionnaire
// before they can apply. Without it (or before a questionnaire is published) applying works as before.
func WithQuestionnaire(gate inbound.ApplicationQuestionnaireGate) Option {
	return func(s *Service) { s.questionnaire = gate }
}

// WithArtistProvisioner makes approving an artist application publish them: a public profile and a
// published artwork for each submitted work. Without it approval creates a private draft profile
// the artist publishes themselves.
func WithArtistProvisioner(p outbound.ArtistProvisioner) Option {
	return func(s *Service) { s.provisioner = p }
}

// WithApplicantWorks requires artist applicants to have uploaded at least MinApplicationWorks
// works before they can apply, and freezes those works onto the application. Without it nothing is required.
func WithApplicantWorks(works outbound.ApplicantWorkRepository) Option {
	return func(s *Service) { s.works = works }
}

func NewService(
	applications outbound.OnboardingRepository,
	users outbound.UserRepository,
	profiles outbound.ProfileRepository,
	institutions outbound.InstitutionRepository,
	opts ...Option,
) *Service {
	s := &Service{
		applications: applications,
		users:        users,
		profiles:     profiles,
		institutions: institutions,
		now:          func() time.Time { return time.Now().UTC() },
	}
	for _, opt := range opts {
		opt(s)
	}
	return s
}

func (s *Service) List(ctx context.Context, filter domain.ListFilter) ([]domain.OnboardingApplication, int, error) {
	if filter.Limit <= 0 {
		filter.Limit = 50
	}
	if filter.Offset < 0 {
		filter.Offset = 0
	}
	return s.applications.List(ctx, filter)
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
	app, err := s.prepareApplication(ctx, applicantID, applicantType, displayName, requestedHandle, notes, nil)
	if err != nil {
		return nil, err
	}
	return s.applications.Save(ctx, *app)
}

// prepareApplication validates a submission and builds the pending application without saving it.
// exceptInvite, when set, is the invite the applicant is claiming: the handle it reserved counts
// as available to them.
func (s *Service) prepareApplication(
	ctx context.Context,
	applicantID uuid.UUID,
	applicantType domain.ApplicantType,
	displayName, requestedHandle, notes string,
	exceptInvite *uuid.UUID,
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
		available, err := s.handleAvailable(ctx, requestedHandle, exceptInvite)
		if err != nil {
			return nil, err
		}
		if !available {
			return nil, apperrors.ErrConflict
		}
	}

	var submitted *inbound.SubmittedQuestionnaire
	if applicantType == domain.ApplicantTypeArtist && s.questionnaire != nil {
		submitted, err = s.questionnaire.CheckComplete(ctx, applicantID)
		if err != nil {
			return nil, err
		}
	}
	var revisionID *uuid.UUID
	var answers []byte
	if submitted != nil {
		revisionID, answers = &submitted.RevisionID, submitted.Answers
	}

	var works []domain.ApplicantWork
	var portrait *domain.ApplicantPortrait
	if applicantType == domain.ApplicantTypeArtist && s.works != nil {
		works, err = s.works.ListByUser(ctx, applicantID)
		if err != nil {
			return nil, err
		}
		if short := domain.WorksShortfall(len(works)); short > 0 {
			return nil, fmt.Errorf("%w: add at least %d works to apply (you have %d, add %d more)",
				apperrors.ErrValidation, domain.MinApplicationWorks, len(works), short)
		}
		// A profile needs a face on it, and approval publishes straight away, so the portrait is
		// collected here instead of leaving the artist to add it afterwards.
		portrait, err = s.works.GetPortrait(ctx, applicantID)
		if errors.Is(err, apperrors.ErrNotFound) {
			return nil, fmt.Errorf("%w: add a portrait photo of yourself to apply", apperrors.ErrValidation)
		}
		if err != nil {
			return nil, err
		}
	}

	now := s.now()
	return &domain.OnboardingApplication{
		ID:                      uuid.New(),
		ApplicantID:             applicantID,
		ApplicantType:           applicantType,
		DisplayName:             displayName,
		RequestedHandle:         requestedHandle,
		Notes:                   notes,
		Status:                  domain.ApprovalStatusPending,
		QuestionnaireRevisionID: revisionID,
		QuestionnaireAnswers:    answers,
		Works:                   works,
		Portrait:                portrait,
		CreatedAt:               now,
		UpdatedAt:               now,
	}, nil
}

func (s *Service) Review(ctx context.Context, id uuid.UUID, reviewerID uuid.UUID, status domain.ApprovalStatus, notes string) (*domain.OnboardingApplication, error) {
	return s.review(ctx, id, &reviewerID, status, notes)
}

// review is Review with an optional reviewer, so a decision taken by the system (an invited
// artist, vetted when the invite was issued) can be recorded without a reviewing user.
func (s *Service) review(ctx context.Context, id uuid.UUID, reviewerID *uuid.UUID, status domain.ApprovalStatus, notes string) (*domain.OnboardingApplication, error) {
	if status != domain.ApprovalStatusApproved && status != domain.ApprovalStatusRejected {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, "status must be approved or rejected")
	}
	app, err := s.applications.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if app.Status != domain.ApprovalStatusPending {
		return nil, fmt.Errorf("%w: this application was already %s", apperrors.ErrConflict, app.Status)
	}

	reviewNotes := strings.TrimSpace(notes)
	if status == domain.ApprovalStatusRejected && reviewNotes == "" {
		return nil, fmt.Errorf("%w: %s", apperrors.ErrValidation, "a reason is required when rejecting an application")
	}

	now := s.now()
	decided := *app
	decided.Status = status
	decided.ReviewNotes = reviewNotes
	decided.ReviewedBy = reviewerID
	decided.ReviewedAt = &now
	decided.UpdatedAt = now

	// Take the decision first, conditioned on the application still being pending. Of two admins
	// deciding at once only one gets past this point, so only one provisions.
	saved, err := s.applications.Transition(ctx, decided, domain.ApprovalStatusPending)
	if err != nil {
		return nil, err
	}

	if status == domain.ApprovalStatusApproved {
		var provisionErr error
		switch app.ApplicantType {
		case domain.ApplicantTypeArtist:
			provisionErr = s.provisionArtist(ctx, app)
		case domain.ApplicantTypeInstitution:
			provisionErr = s.provisionInstitution(ctx, app)
		}
		if provisionErr != nil {
			// Put the application back to pending so the admin can retry; provisioning is safe to repeat.
			revert := *app
			revert.Status = domain.ApprovalStatusPending
			revert.UpdatedAt = s.now()
			if _, revErr := s.applications.Transition(context.WithoutCancel(ctx), revert, domain.ApprovalStatusApproved); revErr != nil {
				return nil, errors.Join(provisionErr, fmt.Errorf("reopen application after failed approval: %w", revErr))
			}
			return nil, provisionErr
		}
	}

	s.notifyDecision(ctx, saved)
	return saved, nil
}

func (s *Service) notifyDecision(ctx context.Context, app *domain.OnboardingApplication) {
	if s.mailer == nil {
		return
	}
	// Best-effort: the decision is already saved and shows on the status page. But an applicant
	// who never hears back is the failure that matters here, so every miss is logged.
	user, err := s.users.GetByID(ctx, app.ApplicantID)
	if err != nil || user == nil || user.Email == "" {
		log.Printf("onboarding: application %s %s, but no decision email (applicant %s not found or has no email): %v", app.ID, app.Status, app.ApplicantID, err)
		return
	}
	subject, body := decisionEmail(app, s.webAppURL)
	if err := s.mailer.Send(ctx, user.Email, subject, body); err != nil {
		log.Printf("onboarding: application %s %s, but the decision email to %s was not sent: %v", app.ID, app.Status, user.Email, err)
	}
}

func decisionEmail(app *domain.OnboardingApplication, webAppURL string) (string, string) {
	var b strings.Builder
	fmt.Fprintf(&b, "Hi %s,\n\n", app.DisplayName)
	if app.Status == domain.ApprovalStatusApproved {
		b.WriteString("Good news: your Artiv application was approved.\n\n")
		if app.ReviewNotes != "" {
			fmt.Fprintf(&b, "A note from the team:\n%s\n\n", app.ReviewNotes)
		}
		if len(app.Works) > 0 && app.RequestedHandle != "" {
			fmt.Fprintf(&b, "Your profile and your %d works are now live on Artiv:\n%s/artists/%s\n\n", len(app.Works), webAppURL, app.RequestedHandle)
			fmt.Fprintf(&b, "Next, open your studio to add a portrait and a short bio:\n%s/studio\n", webAppURL)
		} else {
			fmt.Fprintf(&b, "Next, open your studio, add a portrait and a short bio, and publish your profile:\n%s/studio\n", webAppURL)
		}
		return "Welcome to Artiv — your application was approved", b.String()
	}
	b.WriteString("Thanks for applying to Artiv. We couldn't approve your application this time.\n\n")
	if app.ReviewNotes != "" {
		fmt.Fprintf(&b, "Reason:\n%s\n\n", app.ReviewNotes)
	}
	fmt.Fprintf(&b, "You can update your details and apply again:\n%s/apply\n", webAppURL)
	return "Update on your Artiv application", b.String()
}

// provisionArtist is safe to repeat: a retry after a partial failure finishes the remaining steps,
// including applying the requested handle to a draft profile an earlier attempt already created.
// The role is promoted last so a failure never leaves an artist without their profile.
func (s *Service) provisionArtist(ctx context.Context, app *domain.OnboardingApplication) error {
	if s.provisioner != nil {
		if err := s.provisioner.ProvisionApproved(ctx, *app); err != nil {
			return err
		}
		return s.users.PromoteToArtist(ctx, app.ApplicantID)
	}
	draft, err := s.profiles.GetArtistByUserID(ctx, app.ApplicantID)
	if errors.Is(err, apperrors.ErrNotFound) {
		draft, err = s.profiles.CreateDraftForUser(ctx, app.ApplicantID, app.DisplayName)
	}
	if err != nil {
		return err
	}
	if app.RequestedHandle != "" && draft.Handle != app.RequestedHandle {
		draft.Handle = app.RequestedHandle
		draft.Slug = app.RequestedHandle
		if _, err := s.profiles.SaveArtist(ctx, *draft); err != nil {
			return err
		}
	}
	return s.users.PromoteToArtist(ctx, app.ApplicantID)
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
	return s.handleAvailable(ctx, handle, nil)
}

func (s *Service) handleAvailable(ctx context.Context, handle string, exceptInvite *uuid.UUID) (bool, error) {
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
	if s.invites != nil {
		held, err := s.invites.GetLiveByHandle(ctx, handle, s.now())
		switch {
		case err == nil:
			if exceptInvite == nil || held.ID != *exceptInvite {
				return false, nil
			}
		case !errors.Is(err, apperrors.ErrNotFound):
			return false, err
		}
	}
	return true, nil
}
