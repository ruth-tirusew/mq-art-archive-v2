-- Removes the SAMPLE artists (and everything that hangs off them) from a database.
--
--   1. Take a backup first:   pg_dump "$DATABASE_URL" -Fc -f before-sample-cleanup.dump
--   2. Run this file as it is. It ends in ROLLBACK, so it only PREVIEWS: read the three result tables.
--   3. Happy with the list? Change the last line to COMMIT and run it again.
--
--   psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f scripts/remove_sample_artists.sql
--
-- Who counts as "sample" is decided in ONE place, the sample_users query below:
--   * the people the seed migrations create (fixed ids 33333333-…-3333 and aaaaaaaa-…-000N), and
--   * artist/institution accounts on placeholder addresses (@example.com, @example.test, @makdas.example).
-- Admins and ordinary users are never matched. Add or remove a line there to widen or narrow it.

BEGIN;

CREATE TEMP TABLE sample_users ON COMMIT DROP AS
SELECT u.id, u.email,
       CASE WHEN u.id = '33333333-3333-3333-3333-333333333333' OR u.id::text LIKE 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa000_'
            THEN 'seed migration' ELSE 'placeholder email' END AS why
FROM users u
WHERE u.role IN ('artist', 'institution')
  AND (   u.id = '33333333-3333-3333-3333-333333333333'
       OR u.id::text LIKE 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa000_'
       OR u.email ILIKE '%@example.com'
       OR u.email ILIKE '%@example.test'
       OR u.email ILIKE '%@makdas.example');

CREATE TEMP TABLE sample_profiles ON COMMIT DROP AS
SELECT p.id FROM artist_profiles p JOIN sample_users s ON s.id = p.user_id;

CREATE TEMP TABLE sample_posts ON COMMIT DROP AS
SELECT a.id FROM art_posts a JOIN sample_profiles p ON p.id = a.artist_id;

-- ---- PREVIEW 1: who will be removed -------------------------------------------------------------
SELECT s.email, s.why, p.slug, p.status, u.created_at::date AS created,
       (SELECT count(*) FROM art_posts a WHERE a.artist_id = p.id) AS works
FROM sample_users s
JOIN users u ON u.id = s.id
LEFT JOIN artist_profiles p ON p.user_id = s.id
ORDER BY u.created_at;

-- ---- PREVIEW 2: how much goes with them ---------------------------------------------------------
SELECT 'users' AS what, count(*) AS rows FROM sample_users
UNION ALL SELECT 'artist profiles', count(*) FROM sample_profiles
UNION ALL SELECT 'artworks', count(*) FROM sample_posts
UNION ALL SELECT 'uploaded images', count(*) FROM media_assets WHERE owner_user_id IN (SELECT id FROM sample_users)
UNION ALL SELECT 'applications', count(*) FROM onboarding_applications WHERE applicant_id IN (SELECT id FROM sample_users)
UNION ALL SELECT 'invites to those emails', count(*) FROM artist_invites WHERE lower(email) IN (SELECT lower(email) FROM sample_users)
UNION ALL SELECT 'page-view rows', count(*) FROM page_view_daily WHERE entity_id IN (SELECT id FROM sample_profiles UNION SELECT id FROM sample_posts)
UNION ALL SELECT 'wiki articles they authored (kept)', count(*) FROM articles WHERE author_id IN (SELECT id FROM sample_users);

-- ---- PREVIEW 3: Cloudinary files to delete by hand afterwards (this SQL only removes the database rows)
SELECT public_id FROM media_assets WHERE owner_user_id IN (SELECT id FROM sample_users) ORDER BY public_id;

-- ---- THE CLEAN-UP (children first; the order matters for the foreign keys that do not cascade) ----
DELETE FROM page_view_daily
 WHERE entity_id IN (SELECT id FROM sample_profiles UNION SELECT id FROM sample_posts);

DELETE FROM art_posts WHERE id IN (SELECT id FROM sample_posts);            -- their images' rows go with them

DELETE FROM onboarding_applications WHERE applicant_id IN (SELECT id FROM sample_users);
UPDATE onboarding_applications SET reviewed_by = NULL WHERE reviewed_by IN (SELECT id FROM sample_users);
UPDATE events               SET reviewed_by = NULL WHERE reviewed_by IN (SELECT id FROM sample_users);
UPDATE article_submissions  SET reviewed_by = NULL WHERE reviewed_by IN (SELECT id FROM sample_users);
DELETE FROM article_submissions WHERE submitter_id IN (SELECT id FROM sample_users);

DELETE FROM artist_invites WHERE lower(email) IN (SELECT lower(email) FROM sample_users);
DELETE FROM institution_profiles WHERE user_id IN (SELECT id FROM sample_users);
DELETE FROM artist_profiles WHERE id IN (SELECT id FROM sample_profiles);
DELETE FROM media_assets WHERE owner_user_id IN (SELECT id FROM sample_users);

-- Deleting the user also removes (ON DELETE CASCADE): sign-in links, tokens, notification settings,
-- favourites, highlights, comments, questionnaire answers and nudges, applicant works, digest deliveries.
DELETE FROM users WHERE id IN (SELECT id FROM sample_users);

-- ---- CHECK: nothing sample should be left (every number should be 0) -----------------------------
SELECT 'sample users left' AS check, count(*) AS rows FROM users
 WHERE role IN ('artist', 'institution')
   AND (id = '33333333-3333-3333-3333-333333333333' OR id::text LIKE 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa000_'
        OR email ILIKE '%@example.com' OR email ILIKE '%@example.test' OR email ILIKE '%@makdas.example')
UNION ALL SELECT 'profiles without a user', count(*) FROM artist_profiles p WHERE NOT EXISTS (SELECT 1 FROM users u WHERE u.id = p.user_id)
UNION ALL SELECT 'artworks without an artist', count(*) FROM art_posts a WHERE NOT EXISTS (SELECT 1 FROM artist_profiles p WHERE p.id = a.artist_id);

-- Preview only. After reading the results above, change ROLLBACK to COMMIT and run the file again.
ROLLBACK;
