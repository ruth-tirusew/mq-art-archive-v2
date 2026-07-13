-- +goose Up
INSERT INTO users (id, email, role, created_at, updated_at)
VALUES (
    '33333333-3333-3333-3333-333333333333',
    'selamawit@example.com',
    'artist',
    '2019-03-15T00:00:00Z',
    NOW()
) ON CONFLICT (id) DO NOTHING;

INSERT INTO artist_profiles (
    id, user_id, slug, display_name, bio, contact_location,
    social_instagram, status, created_at, updated_at
) VALUES (
    '44444444-4444-4444-4444-444444444444',
    '33333333-3333-3333-3333-333333333333',
    'selamawit-abebe',
    'Selamawit Abebe',
    'Addis Ababa–based painter exploring memory, ritual, and the textures of everyday life through oil and mixed media. Her work has been shown at the National Museum and regional galleries across Ethiopia.',
    'Addis Ababa, Ethiopia',
    '@selamawit.abebe',
    'approved',
    '2019-03-15T00:00:00Z',
    NOW()
) ON CONFLICT (id) DO NOTHING;

INSERT INTO art_posts (id, artist_id, title, description, medium, status, published_at, created_at, updated_at)
VALUES
    ('55555555-5555-5555-5555-555555555501', '44444444-4444-4444-4444-444444444444', 'Blue Hour Market', 'Evening light over Merkato stalls.', 'oil on canvas', 'published', NOW() - INTERVAL '30 days', NOW(), NOW()),
    ('55555555-5555-5555-5555-555555555502', '44444444-4444-4444-4444-444444444444', 'Coffee Ceremony', 'Three figures in dialogue around the jebena.', 'acrylic', 'published', NOW() - INTERVAL '20 days', NOW(), NOW()),
    ('55555555-5555-5555-5555-555555555503', '44444444-4444-4444-4444-444444444444', 'Entoto Mist', 'Hills dissolving into morning fog.', 'oil on linen', 'published', NOW() - INTERVAL '10 days', NOW(), NOW()),
    ('55555555-5555-5555-5555-555555555504', '44444444-4444-4444-4444-444444444444', 'Thread & Ash', 'Textile fragments and charcoal on board.', 'mixed media', 'published', NOW() - INTERVAL '5 days', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM art_posts WHERE artist_id = '44444444-4444-4444-4444-444444444444';
DELETE FROM artist_profiles WHERE slug = 'selamawit-abebe';
DELETE FROM users WHERE id = '33333333-3333-3333-3333-333333333333';
