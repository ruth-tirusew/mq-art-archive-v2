package analytics

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	domain "github.com/mq/api/internal/domain/analytics"
)

type analyticsRepoStub struct {
	hash     string
	entityID *uuid.UUID
}

func (r *analyticsRepoStub) RecordUnique(_ context.Context, hash, _ string, _ uuid.UUID, _, _ time.Time) (bool, error) {
	r.hash = hash
	return true, nil
}
func (r *analyticsRepoStub) Query(_ context.Context, _ string, entityID *uuid.UUID, _, _ time.Time) ([]domain.View, error) {
	r.entityID = entityID
	return nil, nil
}
func (*analyticsRepoStub) MeOverview(context.Context, uuid.UUID, time.Time, time.Time) ([]domain.View, error) {
	return nil, nil
}

func TestRecordViewUsesStableDailyDedupeHash(t *testing.T) {
	repo := &analyticsRepoStub{}
	svc := NewService(repo).(*Service)
	svc.now = func() time.Time { return time.Date(2026, 7, 26, 12, 0, 0, 0, time.UTC) }
	id := uuid.New()
	if recorded, err := svc.RecordView(context.Background(), "visitor", "artist", id); err != nil || !recorded {
		t.Fatalf("record view: recorded=%v err=%v", recorded, err)
	}
	first := repo.hash
	svc.now = func() time.Time { return time.Date(2026, 7, 26, 23, 0, 0, 0, time.UTC) }
	_, _ = svc.RecordView(context.Background(), "visitor", "artist", id)
	if repo.hash != first {
		t.Fatal("dedupe hash changed within the same UTC day")
	}
}

func TestQueryAllowsAggregateWithoutEntityID(t *testing.T) {
	repo := &analyticsRepoStub{}
	svc := NewService(repo)
	from := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if _, err := svc.Query(context.Background(), "", nil, from, from.AddDate(0, 0, 1)); err != nil {
		t.Fatalf("aggregate query: %v", err)
	}
	if repo.entityID != nil {
		t.Fatal("aggregate query unexpectedly required an entity id")
	}
}
