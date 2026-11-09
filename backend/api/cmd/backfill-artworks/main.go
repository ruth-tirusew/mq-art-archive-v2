// cmd/backfill-artworks publishes the work of artists who were approved before approval did that
// itself: for every approved artist application that carries submitted works, it creates the
// published artwork for each one that is missing. It never changes a profile (a profile someone
// took offline stays offline) and only fills gaps, so it is safe to run again.
//
//	go run ./cmd/backfill-artworks            # do it
//	go run ./cmd/backfill-artworks -dry-run   # only list what would be created
package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/mq/api/config"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
	provisioningadapter "github.com/mq/api/internal/adapter/driven/provisioning"
	"github.com/mq/api/internal/domain/apperrors"
	"github.com/mq/api/internal/domain/onboarding"
)

func main() {
	dryRun := flag.Bool("dry-run", false, "only report what is missing")
	flag.Parse()

	cfg := config.Load()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	pool, err := postgres.NewPool(ctx, cfg)
	cancel()
	if err != nil {
		log.Fatalf("database: %v", err)
	}
	defer pool.Close()
	ctx = context.Background()

	profiles := postgres.NewProfileRepository(pool)
	posts := postgres.NewArtPostRepository(pool)
	provisioner := provisioningadapter.NewArtistProvisioner(profiles, posts)

	rows, err := pool.Query(ctx, `
		SELECT id, applicant_id, display_name, works FROM onboarding_applications
		WHERE status = 'approved' AND applicant_type = 'artist' AND works IS NOT NULL AND jsonb_array_length(works) > 0
		ORDER BY reviewed_at
	`)
	if err != nil {
		log.Fatalf("list applications: %v", err)
	}
	type application struct {
		id, applicant uuid.UUID
		name          string
		works         []onboarding.ApplicantWork
	}
	var apps []application
	for rows.Next() {
		var a application
		var raw []byte
		if err := rows.Scan(&a.id, &a.applicant, &a.name, &raw); err != nil {
			log.Fatalf("scan application: %v", err)
		}
		if err := json.Unmarshal(raw, &a.works); err != nil {
			log.Fatalf("decode works of application %s: %v", a.id, err)
		}
		apps = append(apps, a)
	}
	rows.Close()

	created := 0
	for _, a := range apps {
		profile, err := profiles.GetArtistByUserID(ctx, a.applicant)
		if errors.Is(err, apperrors.ErrNotFound) {
			log.Printf("skip %s: no profile for this artist", a.name)
			continue
		}
		if err != nil {
			log.Fatalf("load profile for %s: %v", a.name, err)
		}
		missing := 0
		for _, w := range a.works {
			if _, err := posts.GetByID(ctx, provisioningadapter.WorkPostID(w.ID)); errors.Is(err, apperrors.ErrNotFound) {
				missing++
			} else if err != nil {
				log.Fatalf("check artwork for %s: %v", a.name, err)
			}
		}
		fmt.Printf("%s (@%s): %d of %d works have no artwork yet\n", a.name, profile.Handle, missing, len(a.works))
		if missing == 0 || *dryRun {
			continue
		}
		if err := provisioner.PublishWorks(ctx, profile.ID, a.works); err != nil {
			log.Fatalf("publish works for %s: %v", a.name, err)
		}
		created += missing
	}
	if *dryRun {
		fmt.Println("dry run: nothing was changed")
		return
	}
	fmt.Printf("done: %d artworks created\n", created)
}
