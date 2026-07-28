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
}

func (m *mockOnboardingRepo) ListPending(ctx context.Context) ([]domain.OnboardingApplication, error) {
	if m.listPending == nil {
		return []domain.OnboardingApplication{}, nil
	}
	return m.listPending(ctx)
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
		Status:        domain.ApprovalStatusPending,
	}

	repo := &mockOnboardingRepo{
		getByID: func(ctx context.Context, id uuid.UUID) (*domain.OnboardingApplication, error) {
			return existing, nil
		},
		save: func(ctx context.Context, app domain.OnboardingApplication) (*domain.OnboardingApplication, error) {
			assist.Equal(t, domain.ApprovalStatusApproved, app.Status)
			assist.Equal(t, "looks good", app.Notes)
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
