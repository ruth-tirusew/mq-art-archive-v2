// cmd/backfill-blurhash computes a blur hash for every image that doesn't have one yet: uploaded
// assets (read as a tiny Cloudinary rendition) and images stored only as URLs (seed data, other
// CDNs). It only fills in blanks and is safe to run again. Run it once after migration 00045 and
// again whenever images were loaded some other way than the upload flow.
//
//	go run ./cmd/backfill-blurhash            # do it
//	go run ./cmd/backfill-blurhash -dry-run   # only report how many images are missing a hash
package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"time"

	"github.com/mq/api/config"
	mediaadapter "github.com/mq/api/internal/adapter/driven/media"
	"github.com/mq/api/internal/adapter/driven/persistence/postgres"
)

// target is a column holding an image URL and, beside it, that image's blur hash.
type target struct {
	table, urlColumn, hashColumn string
}

var targets = []target{
	{"art_post_media", "url", "blurhash"},
	{"artist_profiles", "portrait_url", "portrait_blurhash"},
	{"events", "image_url", "image_blurhash"},
}

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

	if *dryRun {
		report(ctx, pool)
		return
	}

	uploaded := backfillAssets(ctx, pool)
	propagated := propagate(ctx, pool)
	external := backfillExternal(ctx, pool)
	fmt.Printf("done: %d uploaded assets hashed, %d image rows filled from them, %d external images hashed\n", uploaded, propagated, external)
	report(ctx, pool)
}

func report(ctx context.Context, pool *postgres.Pool) {
	var assets int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM media_assets WHERE blurhash = ''`).Scan(&assets); err != nil {
		log.Fatalf("count assets: %v", err)
	}
	fmt.Printf("still missing a blur hash: %d uploaded assets", assets)
	for _, t := range targets {
		var n int
		q := fmt.Sprintf(`SELECT count(*) FROM %s WHERE COALESCE(%s,'') <> '' AND %s = ''`, t.table, t.urlColumn, t.hashColumn)
		if err := pool.QueryRow(ctx, q).Scan(&n); err != nil {
			log.Fatalf("count %s: %v", t.table, err)
		}
		fmt.Printf(", %d %s.%s", n, t.table, t.urlColumn)
	}
	fmt.Println()
}

// backfillAssets hashes uploaded assets from their Cloudinary thumbnail.
func backfillAssets(ctx context.Context, pool *postgres.Pool) int {
	rows, err := pool.Query(ctx, `SELECT id, secure_url FROM media_assets WHERE blurhash = ''`)
	if err != nil {
		log.Fatalf("list assets: %v", err)
	}
	type asset struct{ id, url string }
	var assets []asset
	for rows.Next() {
		var a asset
		if err := rows.Scan(&a.id, &a.url); err != nil {
			log.Fatalf("scan asset: %v", err)
		}
		assets = append(assets, a)
	}
	rows.Close()

	hasher := mediaadapter.NewBlurhasher()
	done := 0
	for _, a := range assets {
		thumb, ok := mediaadapter.ThumbnailFromURL(a.url)
		if !ok {
			log.Printf("skip asset %s: not a Cloudinary image URL (%s)", a.id, a.url)
			continue
		}
		hash, err := hasher.EncodeURL(ctx, thumb)
		if err != nil {
			log.Printf("skip asset %s: %v", a.id, err)
			continue
		}
		if _, err := pool.Exec(ctx, `UPDATE media_assets SET blurhash = $2, updated_at = NOW() WHERE id = $1`, a.id, hash); err != nil {
			log.Fatalf("update asset %s: %v", a.id, err)
		}
		done++
	}
	return done
}

// propagate copies a hash from an uploaded asset to every row that shows its URL.
func propagate(ctx context.Context, pool *postgres.Pool) int {
	total := 0
	for _, t := range targets {
		q := fmt.Sprintf(`UPDATE %[1]s x SET %[3]s = m.blurhash FROM media_assets m
			WHERE x.%[2]s = m.secure_url AND x.%[3]s = '' AND m.blurhash <> ''`, t.table, t.urlColumn, t.hashColumn)
		tag, err := pool.Exec(ctx, q)
		if err != nil {
			log.Fatalf("propagate to %s: %v", t.table, err)
		}
		total += int(tag.RowsAffected())
	}
	return total
}

// backfillExternal hashes images that are only known by URL, once per distinct URL.
func backfillExternal(ctx context.Context, pool *postgres.Pool) int {
	hasher := mediaadapter.NewRemoteBlurhasher()
	done := 0
	for _, t := range targets {
		q := fmt.Sprintf(`SELECT DISTINCT %[2]s FROM %[1]s WHERE COALESCE(%[2]s,'') <> '' AND %[3]s = ''`, t.table, t.urlColumn, t.hashColumn)
		rows, err := pool.Query(ctx, q)
		if err != nil {
			log.Fatalf("list %s: %v", t.table, err)
		}
		var urls []string
		for rows.Next() {
			var u string
			if err := rows.Scan(&u); err != nil {
				log.Fatalf("scan %s: %v", t.table, err)
			}
			urls = append(urls, u)
		}
		rows.Close()

		for _, u := range urls {
			fetch := u
			if thumb, ok := mediaadapter.ThumbnailFromURL(u); ok {
				fetch = thumb
			}
			hash, err := hasher.EncodeURL(ctx, fetch)
			if err != nil {
				log.Printf("skip %s %s: %v", t.table, u, err)
				continue
			}
			up := fmt.Sprintf(`UPDATE %s SET %s = $2 WHERE %s = $1 AND %s = ''`, t.table, t.hashColumn, t.urlColumn, t.hashColumn)
			tag, err := pool.Exec(ctx, up, u, hash)
			if err != nil {
				log.Fatalf("update %s: %v", t.table, err)
			}
			done += int(tag.RowsAffected())
		}
	}
	return done
}
