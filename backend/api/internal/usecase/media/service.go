package media

import (
	"context"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/media"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	signer  outbound.MediaSigner
	storage outbound.MediaStorage
	assets  outbound.MediaAssetRepository
	folder  string
}

func NewService(signer outbound.MediaSigner, storage outbound.MediaStorage, assets outbound.MediaAssetRepository, folder string) inbound.MediaService {
	return &Service{signer: signer, storage: storage, assets: assets, folder: strings.Trim(folder, "/")}
}

func (s *Service) SignForUser(ctx context.Context, _ uuid.UUID) (*domain.UploadSignature, error) {
	publicID := s.folder + "/" + uuid.NewString()
	return s.signer.SignUpload(ctx, domain.UploadOptions{PublicID: publicID, Folder: s.folder})
}

func (s *Service) CompleteUpload(ctx context.Context, userID uuid.UUID, in domain.Completion) (*domain.Asset, error) {
	if !strings.HasPrefix(in.PublicID, s.folder+"/") || in.ResourceType != "image" {
		return nil, fmt.Errorf("%w: invalid media identifier or resource type", apperrors.ErrValidation)
	}
	format := strings.ToLower(in.Format)
	if format != "jpg" && format != "jpeg" && format != "png" && format != "webp" {
		return nil, fmt.Errorf("%w: image format must be jpeg, png, or webp", apperrors.ErrValidation)
	}
	u, err := url.Parse(in.SecureURL)
	if err != nil || u.Scheme != "https" || !strings.HasSuffix(strings.ToLower(u.Hostname()), "cloudinary.com") {
		return nil, fmt.Errorf("%w: invalid secure_url", apperrors.ErrValidation)
	}
	if in.Bytes <= 0 || in.Bytes > domain.MaxImageBytes || in.Width <= 0 || in.Height <= 0 {
		return nil, fmt.Errorf("%w: invalid image dimensions or size", apperrors.ErrValidation)
	}
	now := time.Now().UTC()
	asset := domain.Asset{
		ID: uuid.New(), OwnerUserID: userID, PublicID: in.PublicID, SecureURL: in.SecureURL,
		ResourceType: "image", Width: &in.Width, Height: &in.Height, Bytes: &in.Bytes,
		Folder: s.folder, CreatedAt: now, UpdatedAt: now,
	}
	return s.assets.Create(ctx, asset)
}

func (s *Service) DeleteOwned(ctx context.Context, userID, assetID uuid.UUID) error {
	asset, err := s.assets.GetByID(ctx, assetID)
	if err != nil {
		return err
	}
	if asset.OwnerUserID != userID {
		return apperrors.ErrForbidden
	}
	if err := s.assets.Delete(ctx, asset.ID); err != nil {
		return err
	}
	// Remote deletion is deliberately best-effort after database success.
	_ = s.storage.Delete(ctx, asset.PublicID)
	return nil
}
