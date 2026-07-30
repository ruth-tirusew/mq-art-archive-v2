package settings

import (
	"context"
	"errors"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/internal/domain/apperrors"
	domain "github.com/mq/api/internal/domain/settings"
	"github.com/mq/api/internal/port/inbound"
	"github.com/mq/api/internal/port/outbound"
)

type Service struct {
	repo     outbound.ScrapeSettingsRepository
	reloader outbound.EventSourceReloader
	session  string
}

func NewService(
	repo outbound.ScrapeSettingsRepository,
	reloader outbound.EventSourceReloader,
	sessionPath string,
) inbound.SettingsService {
	return &Service{
		repo:     repo,
		reloader: reloader,
		session:  sessionPath,
	}
}

func (s *Service) GetScrapeSettings(ctx context.Context) (*domain.ScrapeSettingsView, error) {
	cfg, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}
	return s.toView(cfg), nil
}

func (s *Service) UpdateScrapeSettings(ctx context.Context, updatedBy uuid.UUID, update domain.ScrapeSettingsUpdate) (*domain.ScrapeSettingsView, error) {
	current, err := s.repo.Get(ctx)
	if err != nil {
		return nil, err
	}

	applyUpdate(current, update)
	if current.ScrapeTimeoutSeconds <= 0 {
		return nil, apperrors.ErrValidation
	}
	if current.ScrapeIntervalSeconds <= 0 {
		return nil, apperrors.ErrValidation
	}
	if current.TelegramFetchLimit <= 0 {
		return nil, apperrors.ErrValidation
	}

	current.UpdatedAt = time.Now().UTC()
	current.UpdatedBy = &updatedBy

	if err := s.repo.Upsert(ctx, *current); err != nil {
		return nil, err
	}
	if s.reloader != nil {
		if err := s.reloader.Reload(*current); err != nil {
			return nil, err
		}
	}
	return s.toView(current), nil
}

func applyUpdate(current *domain.ScrapeSettings, update domain.ScrapeSettingsUpdate) {
	if update.ScrapeEnabled != nil {
		current.ScrapeEnabled = *update.ScrapeEnabled
	}
	if update.ScrapeSources != nil {
		current.ScrapeSources = *update.ScrapeSources
	}
	if update.ScrapeUserAgent != nil {
		current.ScrapeUserAgent = strings.TrimSpace(*update.ScrapeUserAgent)
	}
	if update.ScrapeTimeoutSeconds != nil {
		current.ScrapeTimeoutSeconds = *update.ScrapeTimeoutSeconds
	}
	if update.ScrapeIntervalSeconds != nil {
		current.ScrapeIntervalSeconds = *update.ScrapeIntervalSeconds
	}
	if update.TelegramEnabled != nil {
		current.TelegramEnabled = *update.TelegramEnabled
	}
	if update.TelegramAPIID != nil {
		current.TelegramAPIID = *update.TelegramAPIID
	}
	if update.TelegramAPIHash != nil {
		// empty string means leave unchanged (UI sends blank to keep secret)
		if strings.TrimSpace(*update.TelegramAPIHash) != "" {
			current.TelegramAPIHash = strings.TrimSpace(*update.TelegramAPIHash)
		}
	}
	if update.TelegramChannels != nil {
		current.TelegramChannels = *update.TelegramChannels
	}
	if update.TelegramKeywords != nil {
		current.TelegramKeywords = *update.TelegramKeywords
	}
	if update.TelegramFetchLimit != nil {
		current.TelegramFetchLimit = *update.TelegramFetchLimit
	}
}

func (s *Service) toView(cfg *domain.ScrapeSettings) *domain.ScrapeSettingsView {
	return &domain.ScrapeSettingsView{
		ScrapeEnabled:         cfg.ScrapeEnabled,
		ScrapeSources:         append([]string(nil), cfg.ScrapeSources...),
		ScrapeUserAgent:       cfg.ScrapeUserAgent,
		ScrapeTimeoutSeconds:  cfg.ScrapeTimeoutSeconds,
		ScrapeIntervalSeconds: cfg.ScrapeIntervalSeconds,
		TelegramEnabled:       cfg.TelegramEnabled,
		TelegramAPIID:         cfg.TelegramAPIID,
		TelegramAPIHashSet:    cfg.TelegramAPIHash != "",
		TelegramChannels:      append([]string(nil), cfg.TelegramChannels...),
		TelegramKeywords:      append([]string(nil), cfg.TelegramKeywords...),
		TelegramFetchLimit:    cfg.TelegramFetchLimit,
		SessionAuthorized:     sessionAuthorized(s.session),
		UpdatedAt:             cfg.UpdatedAt,
	}
}

func sessionAuthorized(path string) bool {
	if path == "" {
		return false
	}
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.Size() > 0
}

// EnsureSeed loads DB settings or seeds from the provided defaults when missing.
func EnsureSeed(ctx context.Context, repo outbound.ScrapeSettingsRepository, seed domain.ScrapeSettings) error {
	_, err := repo.Get(ctx)
	if err == nil {
		return nil
	}
	if !errors.Is(err, apperrors.ErrNotFound) {
		return err
	}
	seed.UpdatedAt = time.Now().UTC()
	return repo.Upsert(ctx, seed)
}
