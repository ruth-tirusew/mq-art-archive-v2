-- +goose Up
-- Extend featured artist with archive metadata
UPDATE artist_profiles SET
    handle = 'selamawit-abebe',
    born = 'b. 1990, Addis Ababa',
    discipline = 'Painting / Mixed media',
    tagline = 'Memory, ritual, and the textures of everyday life.',
    years_active = '2015 — present',
    featured = TRUE,
    portrait_url = 'https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=800&q=80'
WHERE slug = 'selamawit-abebe';

UPDATE art_posts SET
    year = 2024,
    dimensions = '120 × 90 cm',
    city = 'Addis Ababa',
    style = 'Contemporary',
    featured_acquisition = TRUE,
    palette = ARRAY['#1e3a5f', '#c4a574', '#8b4513', '#f5f0e8']
WHERE id = '55555555-5555-5555-5555-555555555501';

UPDATE art_posts SET year = 2024, dimensions = '80 × 60 cm', city = 'Addis Ababa', style = 'Figurative', palette = ARRAY['#8b2942', '#d4a574', '#2c1810']
WHERE id = '55555555-5555-5555-5555-555555555502';

UPDATE art_posts SET year = 2023, dimensions = '100 × 70 cm', city = 'Addis Ababa', style = 'Landscape', palette = ARRAY['#4a6741', '#87ceeb', '#f5f0e8']
WHERE id = '55555555-5555-5555-5555-555555555503';

UPDATE art_posts SET year = 2024, dimensions = '60 × 60 cm', city = 'Addis Ababa', style = 'Mixed media', featured_acquisition = TRUE, palette = ARRAY['#2c1810', '#c4a574', '#8b2942']
WHERE id = '55555555-5555-5555-5555-555555555504';

INSERT INTO art_post_media (id, art_post_id, url, mime_type, sort_order) VALUES
    ('66666666-6666-6666-6666-666666666601', '55555555-5555-5555-5555-555555555501', 'https://images.unsplash.com/photo-1579783902614-a3fb3927b6a5?w=1200&q=80', 'image/jpeg', 0),
    ('66666666-6666-6666-6666-666666666602', '55555555-5555-5555-5555-555555555502', 'https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200&q=80', 'image/jpeg', 0),
    ('66666666-6666-6666-6666-666666666603', '55555555-5555-5555-5555-555555555503', 'https://images.unsplash.com/photo-1578301978693-85fa9d0ae59c?w=1200&q=80', 'image/jpeg', 0),
    ('66666666-6666-6666-6666-666666666604', '55555555-5555-5555-5555-555555555504', 'https://images.unsplash.com/photo-1515405295579-ba7b45403062?w=1200&q=80', 'image/jpeg', 0)
ON CONFLICT (id) DO NOTHING;

-- Additional artists
INSERT INTO users (id, email, role, created_at, updated_at) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0001', 'tewodros@example.com', 'artist', NOW(), NOW()),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0002', 'meron@example.com', 'artist', NOW(), NOW()),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0003', 'abel@example.com', 'artist', NOW(), NOW()),
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0004', 'hewan@example.com', 'artist', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO artist_profiles (id, user_id, slug, handle, display_name, bio, born, discipline, tagline, years_active, portrait_url, featured, contact_location, social_instagram, status, created_at, updated_at) VALUES
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0001', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0001', 'tewodros-hailu', 'tewodros-hailu', 'Tewodros Hailu', 'Painter exploring memory, migration, and the residual marks of city walls.', 'b. 1987, Addis Ababa', 'Painting / Mixed media', 'Painter of residue — walls, weather, and the memory of both.', '2009 — present', 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=800&q=80', TRUE, 'Addis Ababa, Ethiopia', '@tewodros_studio', 'approved', NOW(), NOW()),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0002', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0002', 'meron-alemu', 'meron-alemu', 'Meron Alemu', 'Collides Orthodox iconography with abstract chromatic noise.', 'b. 1991, Bahir Dar', 'Painting / Collage', 'Color as language.', '2013 — present', 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=800&q=80', FALSE, 'Berlin / Addis Ababa', '@meronalemu', 'approved', NOW(), NOW()),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0003', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0003', 'abel-getachew', 'abel-getachew', 'Abel Getachew', 'Monumental color fields punctuated by fidäl broken into pure form.', 'b. 1984, Harar', 'Painting', '', '2006 — present', 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=800&q=80', FALSE, 'Addis Ababa', '', 'approved', NOW(), NOW()),
    ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0004', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0004', 'hewan-tadesse', 'hewan-tadesse', 'Hewan Tadesse', 'Chromatic blocks layered with woodblock prints of household objects.', 'b. 1995, Dire Dawa', 'Painting / Print', '', '2017 — present', 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=800&q=80', FALSE, 'Brooklyn', '@hewan.tadesse', 'approved', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO art_posts (id, artist_id, title, description, medium, year, dimensions, city, style, featured_acquisition, palette, status, published_at, created_at, updated_at) VALUES
    ('cccccccc-cccc-cccc-cccc-cccccccc0001', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0001', 'Ascension, after the rains', 'Oil and gold leaf on linen.', 'oil on linen', 2023, '150 × 120 cm', 'Addis Ababa', 'Abstract', TRUE, ARRAY['#1e3a5f', '#d4a853', '#8b2942'], 'published', NOW() - INTERVAL '3 days', NOW(), NOW()),
    ('cccccccc-cccc-cccc-cccc-cccccccc0002', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0002', 'Four Saints in Indigo', 'Acrylic and collage on canvas.', 'acrylic', 2024, '110 × 90 cm', 'Berlin', 'Figurative', TRUE, ARRAY['#2c3e6b', '#c4a574', '#f5f0e8'], 'published', NOW() - INTERVAL '7 days', NOW(), NOW()),
    ('cccccccc-cccc-cccc-cccc-cccccccc0003', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0003', 'Fidäl, dissolving', 'Oil on canvas.', 'oil on canvas', 2024, '90 × 90 cm', 'Addis Ababa', 'Abstract', FALSE, ARRAY['#8b2942', '#2c1810', '#d4a574'], 'published', NOW() - INTERVAL '14 days', NOW(), NOW()),
    ('cccccccc-cccc-cccc-cccc-cccccccc0004', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0004', 'Conversations in Pink', 'Woodblock and acrylic.', 'mixed media', 2023, '76 × 56 cm', 'Brooklyn', 'Contemporary', FALSE, ARRAY['#e8a0bf', '#2c1810', '#f5f0e8'], 'published', NOW() - INTERVAL '21 days', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO art_post_media (id, art_post_id, url, mime_type, sort_order) VALUES
    ('dddddddd-dddd-dddd-dddd-dddddddd0001', 'cccccccc-cccc-cccc-cccc-cccccccc0001', 'https://images.unsplash.com/photo-1547891654-e66ed7ebb968?w=1200&q=80', 'image/jpeg', 0),
    ('dddddddd-dddd-dddd-dddd-dddddddd0002', 'cccccccc-cccc-cccc-cccc-cccccccc0002', 'https://images.unsplash.com/photo-1578301978693-85fa9d0ae59c?w=1200&q=80', 'image/jpeg', 0),
    ('dddddddd-dddd-dddd-dddd-dddddddd0003', 'cccccccc-cccc-cccc-cccc-cccccccc0003', 'https://images.unsplash.com/photo-1579783902614-a3fb3927b6a5?w=1200&q=80', 'image/jpeg', 0),
    ('dddddddd-dddd-dddd-dddd-dddddddd0004', 'cccccccc-cccc-cccc-cccc-cccccccc0004', 'https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200&q=80', 'image/jpeg', 0)
ON CONFLICT (id) DO NOTHING;

-- Wiki articles
INSERT INTO articles (id, slug, title, body, category, excerpt, reading_time, difficulty, verified, contributors, status, author_id, created_at, updated_at) VALUES
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeee0001', 'copyright-basics', 'Copyright Basics for Artists', 'Ethiopian copyright law protects original works from the moment of creation. Register with the Intellectual Property Office for additional protection.', 'Legal', 'What every Ethiopian artist should know about protecting their work.', 8, 'Beginner', TRUE, 12, 'published', '33333333-3333-3333-3333-333333333333', NOW(), NOW()),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeee0002', 'oil-painting-techniques', 'Oil Painting Techniques', 'Layering, glazing, and impasto methods used by contemporary Ethiopian painters.', 'Technique', 'Foundational oil techniques from Addis studios.', 15, 'Intermediate', TRUE, 8, 'published', '33333333-3333-3333-3333-333333333333', NOW(), NOW()),
    ('eeeeeeee-eeee-eeee-eeee-eeeeeeee0003', 'gallery-contracts', 'Understanding Gallery Contracts', 'Key clauses to review before signing with a gallery or dealer.', 'Legal', 'Commission rates, consignment terms, and exit clauses explained.', 12, 'Advanced', FALSE, 5, 'published', '33333333-3333-3333-3333-333333333333', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- Events
INSERT INTO events (id, title, description, source_url, slug, event_type, venue, city, starts_at, scraped_at, status, created_at, updated_at) VALUES
    ('ffffffff-ffff-ffff-ffff-ffffffff0001', 'The Long Conversation', 'Two-person show at Zoma Museum featuring Tewodros Hailu and Meron Alemu.', 'https://makdas.example/events/long-conversation', 'long-conversation', 'Exhibition', 'Zoma Museum', 'Addis Ababa', NOW() + INTERVAL '14 days', NOW(), 'approved', NOW(), NOW()),
    ('ffffffff-ffff-ffff-ffff-ffffffff0002', 'Addis Art Week Opening', 'Annual opening night for Addis Art Week.', 'https://makdas.example/events/addis-art-week', 'addis-art-week', 'Opening', 'National Museum', 'Addis Ababa', NOW() + INTERVAL '30 days', NOW(), 'approved', NOW(), NOW()),
    ('ffffffff-ffff-ffff-ffff-ffffffff0003', 'Studio Visit: Piazza', 'Open studio day in the Piazza district.', 'https://makdas.example/events/piazza-studio', 'piazza-studio', 'Studio visit', 'Piazza Studios', 'Addis Ababa', NOW() + INTERVAL '7 days', NOW(), 'approved', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM events WHERE id IN ('ffffffff-ffff-ffff-ffff-ffffffff0001', 'ffffffff-ffff-ffff-ffff-ffffffff0002', 'ffffffff-ffff-ffff-ffff-ffffffff0003');
DELETE FROM articles WHERE id IN ('eeeeeeee-eeee-eeee-eeee-eeeeeeee0001', 'eeeeeeee-eeee-eeee-eeee-eeeeeeee0002', 'eeeeeeee-eeee-eeee-eeee-eeeeeeee0003');
DELETE FROM art_post_media WHERE art_post_id IN ('cccccccc-cccc-cccc-cccc-cccccccc0001', 'cccccccc-cccc-cccc-cccc-cccccccc0002', 'cccccccc-cccc-cccc-cccc-cccccccc0003', 'cccccccc-cccc-cccc-cccc-cccccccc0004', '55555555-5555-5555-5555-555555555501', '55555555-5555-5555-5555-555555555502', '55555555-5555-5555-5555-555555555503', '55555555-5555-5555-5555-555555555504');
DELETE FROM art_posts WHERE id IN ('cccccccc-cccc-cccc-cccc-cccccccc0001', 'cccccccc-cccc-cccc-cccc-cccccccc0002', 'cccccccc-cccc-cccc-cccc-cccccccc0003', 'cccccccc-cccc-cccc-cccc-cccccccc0004');
DELETE FROM artist_profiles WHERE id IN ('bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0001', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0002', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0003', 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0004');
DELETE FROM users WHERE id IN ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0001', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0002', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0003', 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0004');
