package onboarding

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/identity"
	institutiondomain "github.com/mq/api/internal/domain/institution"
	domain "github.com/mq/api/internal/domain/onboarding"
	profiledomain "github.com/mq/api/internal/domain/profile"
	"github.com/mq/api/internal/testutil/assist"
)

type mockOnboardingRepo struct {
	listPending            func(ctx context.Context) ([]domain.OnboardingApplication, error)
	getByID                func(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error)
	getLatestByApplicantID func(ctx context.Context, applicantID uuid.UUID) (*domain.OnboardingApplication, error)
	save                   func(ctx context.Context, app domain.OnboardingApplication) (*domain.OnboardingApplication, error)
	// transition defaults to save, so tests that only care about the stored result need set just one hook.
	transition func(ctx context.Context, app domain.OnboardingApplication, from domain.ApprovalStatus) (*domain.OnboardingApplication, error)
}

func (m *mockOnboardingRepo) ListPending(ctx context.Context) ([]domain.OnboardingApplication, error) {
	if m.listPending == nil {
		return []domain.OnboardingApplication{}, nil
	}
	return m.listPending(ctx)
}

func (m *mockOnboardingRepo) List(ctx context.Context, filter domain.ListFilter) ([]domain.OnboardingApplication, int, error) {
	apps, err := m.ListPending(ctx)
	return apps, len(apps), err
}

func (m *mockOnboardingRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
	return m.getByID(ctx, id)
}

func (m *mockOnboardingRepo) GetLatestByApplicantID(ctx context.Context, applicantID uuid.UUID) (*domain.OnboardingApplication, error) {
	if m.getLatestByApplicantID != nil {
		return m.getLatestByApplicantID(ctx, applicantID)
	}
	return nil, apperrors.ErrNotFound
}

func (m *mockOnboardingRepo) Save(ctx context.Context, app domain.OnboardingApplication) (*domain.OnboardingApplication, error) {
	return m.save(ctx, app)
}

func (m *mockOnboardingRepo) Transition(ctx context.Context, app domain.OnboardingApplication, from domain.ApprovalStatus) (*domain.OnboardingApplication, error) {
	if m.transition != nil {
		return m.transition(ctx, app, from)
	}
	return m.save(ctx, app)
}

type mockUserRepo struct {
	updateRole func(ctx context.Context, id uuid.UUID, role identity.Role) error
}

func (m *mockUserRepo) GetByID(ctx context.Context, id uuid.UUID) (*identity.User, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*identity.User, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockUserRepo) GetAuthByEmail(ctx context.Context, email string) (*identity.User, *string, error) {
	return nil, nil, apperrors.ErrNotFound
}
func (m *mockUserRepo) GetAuthByID(ctx context.Context, id uuid.UUID) (*identity.User, *string, error) {
	return nil, nil, apperrors.ErrNotFound
}
func (m *mockUserRepo) Create(ctx context.Context, user identity.User) error {
	return nil
}
func (m *mockUserRepo) CreateWithPassword(ctx context.Context, user identity.User, passwordHash string) error {
	return nil
}
func (m *mockUserRepo) UpdatePassword(ctx context.Context, id uuid.UUID, passwordHash string) error {
	return nil
}
func (m *mockUserRepo) UpdateEmail(ctx context.Context, id uuid.UUID, email string) error {
	return nil
}
func (m *mockUserRepo) UpdateProfile(ctx context.Context, id uuid.UUID, displayName, avatarURL string) error {
	return nil
}
func (m *mockUserRepo) UpdateRole(ctx context.Context, id uuid.UUID, role identity.Role) error {
	if m.updateRole != nil {
		return m.updateRole(ctx, id, role)
	}
	return nil
}
func (m *mockUserRepo) PromoteToArtist(ctx context.Context, id uuid.UUID) error {
	return m.UpdateRole(ctx, id, identity.RoleArtist)
}
func (m *mockUserRepo) PromoteToInstitution(ctx context.Context, id uuid.UUID) error {
	return m.UpdateRole(ctx, id, identity.RoleInstitution)
}

type mockInstitutionRepo struct {
	getByUserID        func(ctx context.Context, userID uuid.UUID) (*institutiondomain.Institution, error)
	createDraftForUser func(ctx context.Context, userID uuid.UUID, displayName string) (*institutiondomain.Institution, error)
}

func (m *mockInstitutionRepo) GetBySlug(ctx context.Context, slug string) (*institutiondomain.Institution, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockInstitutionRepo) GetByID(ctx context.Context, id uuid.UUID) (*institutiondomain.Institution, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockInstitutionRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*institutiondomain.Institution, error) {
	if m.getByUserID != nil {
		return m.getByUserID(ctx, userID)
	}
	return nil, apperrors.ErrNotFound
}
func (m *mockInstitutionRepo) Save(ctx context.Context, inst institutiondomain.Institution) (*institutiondomain.Institution, error) {
	return &inst, nil
}
func (m *mockInstitutionRepo) CreateDraftForUser(ctx context.Context, userID uuid.UUID, displayName string) (*institutiondomain.Institution, error) {
	if m.createDraftForUser != nil {
		return m.createDraftForUser(ctx, userID, displayName)
	}
	return &institutiondomain.Institution{UserID: userID, Name: displayName, Slug: "draft"}, nil
}

type mockProfileRepo struct {
	getByUserID        func(ctx context.Context, userID uuid.UUID) (*profiledomain.ArtistProfile, error)
	save               func(ctx context.Context, p profiledomain.ArtistProfile) (*profiledomain.ArtistProfile, error)
	createDraftForUser func(ctx context.Context, userID uuid.UUID, displayName string) (*profiledomain.ArtistProfile, error)
}

func (m *mockProfileRepo) GetArtistBySlug(ctx context.Context, slug string) (*profiledomain.ArtistProfile, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockProfileRepo) GetArtistByHandle(ctx context.Context, handle string) (*profiledomain.ArtistProfile, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockProfileRepo) ListApproved(ctx context.Context, filter profiledomain.ListFilter) ([]profiledomain.ArtistProfile, error) {
	return nil, nil
}
func (m *mockProfileRepo) ListAll(ctx context.Context, status *profiledomain.ProfileStatus, limit, offset int) ([]profiledomain.ArtistProfile, error) {
	return nil, nil
}
func (m *mockProfileRepo) GetArtistByID(ctx context.Context, id uuid.UUID) (*profiledomain.ArtistProfile, error) {
	return nil, apperrors.ErrNotFound
}
func (m *mockProfileRepo) GetArtistByUserID(ctx context.Context, userID uuid.UUID) (*profiledomain.ArtistProfile, error) {
	if m.getByUserID != nil {
		return m.getByUserID(ctx, userID)
	}
	return nil, apperrors.ErrNotFound
}
func (m *mockProfileRepo) SaveArtist(ctx context.Context, p profiledomain.ArtistProfile) (*profiledomain.ArtistProfile, error) {
	if m.save != nil {
		return m.save(ctx, p)
	}
	return &p, nil
}
func (m *mockProfileRepo) CreateDraftForUser(ctx context.Context, userID uuid.UUID, displayName string) (*profiledomain.ArtistProfile, error) {
	if m.createDraftForUser != nil {
		return m.createDraftForUser(ctx, userID, displayName)
	}
	return &profiledomain.ArtistProfile{UserID: userID, DisplayName: displayName, Slug: "studio-x"}, nil
}
func (m *mockProfileRepo) Delete(ctx context.Context, id uuid.UUID) error {
	return nil
}

func TestService_Review_updatesApplication(t *testing.T) {
	appID := uuid.New()
	reviewerID := uuid.New()
	applicantID := uuid.New()
	existing := &domain.OnboardingApplication{
		ID:            appID,
		ApplicantID:   applicantID,
		ApplicantType: domain.ApplicantTypeArtist,
		DisplayName:   "Studio X",
		Notes:         "instagram.com/studiox",
		Status:        domain.ApprovalStatusPending,
	}

	repo := &mockOnboardingRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
			return existing, nil
		},
		save: func(ctx context.Context, app domain.OnboardingApplication) (*domain.OnboardingApplication, error) {
			assist.Equal(t, domain.ApprovalStatusApproved, app.Status)
			assist.Equal(t, "looks good", app.ReviewNotes)
			assist.Equal(t, "instagram.com/studiox", app.Notes)
			assist.Equal(t, reviewerID, *app.ReviewedBy)
			assist.NotNil(t, app.ReviewedAt)
			return &app, nil
		},
	}
	users := &mockUserRepo{
		updateRole: func(ctx context.Context, id uuid.UUID, role identity.Role) error {
			assist.Equal(t, applicantID, id)
			assist.Equal(t, identity.RoleArtist, role)
			return nil
		},
	}
	profiles := &mockProfileRepo{
		getByUserID: func(ctx context.Context, userID uuid.UUID) (*profiledomain.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
		createDraftForUser: func(ctx context.Context, userID uuid.UUID, displayName string) (*profiledomain.ArtistProfile, error) {
			assist.Equal(t, applicantID, userID)
			assist.Equal(t, "Studio X", displayName)
			return &profiledomain.ArtistProfile{UserID: userID, Slug: "studio-x", Status: profiledomain.ProfileStatusDraft}, nil
		},
	}
	svc := NewService(repo, users, profiles, &mockInstitutionRepo{})

	got, err := svc.Review(context.Background(), appID, reviewerID, domain.ApprovalStatusApproved, "looks good")
	assist.NoError(t, err)
	assist.Equal(t, domain.ApprovalStatusApproved, got.Status)
	assist.WithinDuration(t, time.Now().UTC(), got.UpdatedAt, time.Second)
}

func TestService_Review_provisionsInstitution(t *testing.T) {
	appID := uuid.New()
	reviewerID := uuid.New()
	applicantID := uuid.New()
	existing := &domain.OnboardingApplication{
		ID:            appID,
		ApplicantID:   applicantID,
		ApplicantType: domain.ApplicantTypeInstitution,
		DisplayName:   "National Gallery",
		Status:        domain.ApprovalStatusPending,
	}

	repo := &mockOnboardingRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
			return existing, nil
		},
		save: func(ctx context.Context, app domain.OnboardingApplication) (*domain.OnboardingApplication, error) {
			return &app, nil
		},
	}
	users := &mockUserRepo{
		updateRole: func(ctx context.Context, id uuid.UUID, role identity.Role) error {
			assist.Equal(t, applicantID, id)
			assist.Equal(t, identity.RoleInstitution, role)
			return nil
		},
	}
	institutions := &mockInstitutionRepo{
		getByUserID: func(ctx context.Context, userID uuid.UUID) (*institutiondomain.Institution, error) {
			return nil, apperrors.ErrNotFound
		},
		createDraftForUser: func(ctx context.Context, userID uuid.UUID, displayName string) (*institutiondomain.Institution, error) {
			assist.Equal(t, applicantID, userID)
			assist.Equal(t, "National Gallery", displayName)
			return &institutiondomain.Institution{UserID: userID, Slug: "national-gallery", Status: institutiondomain.InstitutionStatusDraft}, nil
		},
	}
	svc := NewService(repo, users, &mockProfileRepo{}, institutions)

	got, err := svc.Review(context.Background(), appID, reviewerID, domain.ApprovalStatusApproved, "approved")
	assist.NoError(t, err)
	assist.Equal(t, domain.ApprovalStatusApproved, got.Status)
}

func TestService_Submit_rejectsExistingPending(t *testing.T) {
	applicantID := uuid.New()
	repo := &mockOnboardingRepo{
		getLatestByApplicantID: func(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
			return &domain.OnboardingApplication{Status: domain.ApprovalStatusPending}, nil
		},
	}
	svc := NewService(repo, &mockUserRepo{}, &mockProfileRepo{
		getByUserID: func(ctx context.Context, userID uuid.UUID) (*profiledomain.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
	}, &mockInstitutionRepo{})

	_, err := svc.Submit(context.Background(), applicantID, domain.ApplicantTypeArtist, "Name", "")
	assist.ErrorIs(t, err, apperrors.ErrConflict)
}

func TestService_Submit_createsApplication(t *testing.T) {
	applicantID := uuid.New()
	repo := &mockOnboardingRepo{
		getLatestByApplicantID: func(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
			return nil, apperrors.ErrNotFound
		},
		save: func(ctx context.Context, app domain.OnboardingApplication) (*domain.OnboardingApplication, error) {
			assist.Equal(t, applicantID, app.ApplicantID)
			assist.Equal(t, domain.ApprovalStatusPending, app.Status)
			assist.Equal(t, "studio_x", app.RequestedHandle)
			return &app, nil
		},
	}
	svc := NewService(repo, &mockUserRepo{}, &mockProfileRepo{
		getByUserID: func(ctx context.Context, userID uuid.UUID) (*profiledomain.ArtistProfile, error) {
			return nil, apperrors.ErrNotFound
		},
	}, &mockInstitutionRepo{})

	got, err := svc.SubmitWithHandle(context.Background(), applicantID, domain.ApplicantTypeArtist, "Studio X", "studio_x", "portfolio link")
	assist.NoError(t, err)
	assist.Equal(t, "Studio X", got.DisplayName)
}

func TestService_Review_rejectRequiresReason(t *testing.T) {
	appID := uuid.New()
	repo := &mockOnboardingRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
			return &domain.OnboardingApplication{ID: appID, ApplicantID: uuid.New(), Status: domain.ApprovalStatusPending}, nil
		},
	}
	svc := NewService(repo, &mockUserRepo{}, &mockProfileRepo{}, &mockInstitutionRepo{})

	_, err := svc.Review(context.Background(), appID, uuid.New(), domain.ApprovalStatusRejected, "  ")
	assist.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestDecisionEmail(t *testing.T) {
	rejected := &domain.OnboardingApplication{DisplayName: "Sam", Status: domain.ApprovalStatusRejected, ReviewNotes: "Please add a link to your work"}
	subject, body := decisionEmail(rejected, "https://artiv.example")
	assist.Contains(t, subject, "Update")
	assist.Contains(t, body, "Please add a link to your work")
	assist.Contains(t, body, "https://artiv.example/apply")

	approved := &domain.OnboardingApplication{DisplayName: "Sam", Status: domain.ApprovalStatusApproved}
	subject, body = decisionEmail(approved, "https://artiv.example")
	assist.Contains(t, subject, "approved")
	assist.Contains(t, body, "https://artiv.example/studio")
}

func pendingArtistApp(handle string) *domain.OnboardingApplication {
	return &domain.OnboardingApplication{
		ID: uuid.New(), ApplicantID: uuid.New(), ApplicantType: domain.ApplicantTypeArtist,
		DisplayName: "Studio X", RequestedHandle: handle, Status: domain.ApprovalStatusPending,
	}
}

func TestService_Review_rejectsAlreadyDecidedApplication(t *testing.T) {
	for _, status := range []domain.ApprovalStatus{domain.ApprovalStatusApproved, domain.ApprovalStatusRejected} {
		app := pendingArtistApp("studio_x")
		app.Status = status
		promoted := false
		repo := &mockOnboardingRepo{
			getByID: func(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) { return app, nil },
			transition: func(ctx context.Context, a domain.OnboardingApplication, from domain.ApprovalStatus) (*domain.OnboardingApplication, error) {
				t.Fatal("a decided application must not be written again")
				return nil, nil
			},
		}
		users := &mockUserRepo{updateRole: func(ctx context.Context, id uuid.UUID, role identity.Role) error {
			promoted = true
			return nil
		}}
		svc := NewService(repo, users, &mockProfileRepo{}, &mockInstitutionRepo{})

		_, err := svc.Review(context.Background(), app.ID, uuid.New(), domain.ApprovalStatusApproved, "")
		assist.ErrorIs(t, err, apperrors.ErrConflict)
		if promoted {
			t.Fatal("re-review must not provision")
		}
	}
}

func TestService_Review_losingTheRaceDoesNotProvision(t *testing.T) {
	app := pendingArtistApp("studio_x")
	promoted := false
	repo := &mockOnboardingRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
			c := *app
			return &c, nil
		},
		transition: func(ctx context.Context, a domain.OnboardingApplication, from domain.ApprovalStatus) (*domain.OnboardingApplication, error) {
			return nil, apperrors.ErrConflict // another admin decided between our read and write
		},
	}
	users := &mockUserRepo{updateRole: func(ctx context.Context, id uuid.UUID, role identity.Role) error {
		promoted = true
		return nil
	}}
	svc := NewService(repo, users, &mockProfileRepo{}, &mockInstitutionRepo{})

	_, err := svc.Review(context.Background(), app.ID, uuid.New(), domain.ApprovalStatusApproved, "")
	assist.ErrorIs(t, err, apperrors.ErrConflict)
	if promoted {
		t.Fatal("the losing admin must not provision")
	}
}

func TestService_Review_rejectsInvalidTargetStatus(t *testing.T) {
	svc := NewService(&mockOnboardingRepo{}, &mockUserRepo{}, &mockProfileRepo{}, &mockInstitutionRepo{})
	_, err := svc.Review(context.Background(), uuid.New(), uuid.New(), domain.ApprovalStatusPending, "")
	assist.ErrorIs(t, err, apperrors.ErrValidation)
}

func TestService_Review_failedProvisioningReopensApplicationAndRetryAppliesHandle(t *testing.T) {
	app := pendingArtistApp("studio_x")
	stored := *app
	repo := &mockOnboardingRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
			c := stored
			return &c, nil
		},
		transition: func(ctx context.Context, a domain.OnboardingApplication, from domain.ApprovalStatus) (*domain.OnboardingApplication, error) {
			if stored.Status != from {
				return nil, apperrors.ErrConflict
			}
			stored.Status = a.Status
			c := stored
			return &c, nil
		},
	}
	var profile *profiledomain.ArtistProfile
	saveFails := true
	role := identity.RolePublic
	users := &mockUserRepo{updateRole: func(ctx context.Context, id uuid.UUID, r identity.Role) error { role = r; return nil }}
	profiles := &mockProfileRepo{
		getByUserID: func(ctx context.Context, userID uuid.UUID) (*profiledomain.ArtistProfile, error) {
			if profile == nil {
				return nil, apperrors.ErrNotFound
			}
			c := *profile
			return &c, nil
		},
		createDraftForUser: func(ctx context.Context, userID uuid.UUID, displayName string) (*profiledomain.ArtistProfile, error) {
			profile = &profiledomain.ArtistProfile{UserID: userID, Handle: "draft-1", Slug: "draft-1"}
			c := *profile
			return &c, nil
		},
		save: func(ctx context.Context, p profiledomain.ArtistProfile) (*profiledomain.ArtistProfile, error) {
			if saveFails {
				return nil, apperrors.ErrConflict // handle taken in the meantime
			}
			profile = &p
			return &p, nil
		},
	}
	svc := NewService(repo, users, profiles, &mockInstitutionRepo{})

	_, err := svc.Review(context.Background(), app.ID, uuid.New(), domain.ApprovalStatusApproved, "")
	assist.ErrorIs(t, err, apperrors.ErrConflict)
	assist.Equal(t, domain.ApprovalStatusPending, stored.Status)
	assist.Equal(t, identity.RolePublic, role) // never promoted without a finished profile

	saveFails = false
	got, err := svc.Review(context.Background(), app.ID, uuid.New(), domain.ApprovalStatusApproved, "")
	assist.NoError(t, err)
	assist.Equal(t, domain.ApprovalStatusApproved, got.Status)
	assist.Equal(t, "studio_x", profile.Handle) // the retry still applied the requested handle
	assist.Equal(t, identity.RoleArtist, role)
}
