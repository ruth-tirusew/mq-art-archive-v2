-- +goose Up
-- Align backend seed data with mq-art-archive-svelte prototype placeholders.

-- Single featured artist (Selamawit); full bios for roster artists.
UPDATE artist_profiles SET
    featured = FALSE,
    updated_at = NOW()
WHERE slug != 'selamawit-abebe';

UPDATE artist_profiles SET
    handle = 'selamawit-abebe',
    born = 'b. 1990, Addis Ababa',
    discipline = 'Painting / Mixed media',
    tagline = 'Memory, ritual, and the textures of everyday life.',
    years_active = '2015 — present',
    featured = TRUE,
    portrait_url = 'https://images.unsplash.com/photo-1544005313-94ddf0286df2?w=800&q=80',
    bio = 'Addis Ababa–based painter exploring memory, ritual, and the textures of everyday life through oil and mixed media. Her work has been shown at the National Museum and regional galleries across Ethiopia.',
    contact_location = 'Addis Ababa, Ethiopia',
    social_instagram = '@selamawit.abebe',
    updated_at = NOW()
WHERE slug = 'selamawit-abebe';

UPDATE artist_profiles SET
    bio = 'Tewodros works in oil and pigment-soaked linen, building gestural fields that recall the bold mark-making of Gebre Kristos Desta. His practice circles around memory, migration, and the residual marks of city walls.',
    born = 'b. 1987, Addis Ababa',
    discipline = 'Painting / Mixed media',
    tagline = 'Painter of residue — walls, weather, and the memory of both.',
    years_active = '2009 — present',
    portrait_url = 'https://images.unsplash.com/photo-1507003211169-0a1dd7228f2d?w=800&q=80',
    contact_location = 'Addis Ababa',
    social_instagram = '@tewodros_studio',
    social_telegram = 'https://t.me/tewodros_studio',
    featured = FALSE,
    updated_at = NOW()
WHERE slug = 'tewodros-hailu';

UPDATE artist_profiles SET
    bio = 'Meron''s canvases collide Orthodox iconography with abstract chromatic noise. She treats color as language — a system of grammar she learned from her grandmother''s church murals.',
    born = 'b. 1991, Bahir Dar',
    discipline = 'Painting / Collage',
    tagline = 'Color as language.',
    years_active = '2013 — present',
    portrait_url = 'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=800&q=80',
    contact_location = 'Berlin / Addis Ababa',
    social_instagram = '@meronalemu',
    social_telegram = 'https://t.me/meronalemu',
    featured = FALSE,
    updated_at = NOW()
WHERE slug = 'meron-alemu';

UPDATE artist_profiles SET
    bio = 'Working from a small studio above the Mercato, Abel paints monumental color fields punctuated by fidäl — the syllabic script of Amharic — broken apart into pure form.',
    born = 'b. 1984, Harar',
    discipline = 'Painting',
    tagline = 'Fidäl broken into pure form.',
    years_active = '2006 — present',
    portrait_url = 'https://images.unsplash.com/photo-1472099645785-5658abf4ff4e?w=800&q=80',
    contact_location = 'Addis Ababa',
    social_telegram = 'https://t.me/abelgetachew',
    featured = FALSE,
    updated_at = NOW()
WHERE slug = 'abel-getachew';

UPDATE artist_profiles SET
    bio = 'Hewan''s work is a quiet riot — chromatic blocks layered with woodblock prints of household objects, market scales, and her mother''s hands.',
    born = 'b. 1995, Dire Dawa',
    discipline = 'Painting / Print',
    tagline = 'A quiet riot of color and print.',
    years_active = '2017 — present',
    portrait_url = 'https://images.unsplash.com/photo-1438761681033-6461ffad8d80?w=800&q=80',
    contact_location = 'Brooklyn',
    social_instagram = '@hewan.tadesse',
    featured = FALSE,
    updated_at = NOW()
WHERE slug = 'hewan-tadesse';

-- Demo shareable profile (portfolio placeholder)
INSERT INTO users (id, email, role, created_at, updated_at) VALUES
    ('aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0005', 'demo@makdas.example', 'artist', NOW(), NOW())
ON CONFLICT (id) DO NOTHING;

INSERT INTO artist_profiles (
    id, user_id, slug, handle, display_name, bio, born, discipline, tagline, years_active,
    portrait_url, featured, contact_location, social_instagram, social_telegram, status, created_at, updated_at
) VALUES (
    'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0005',
    'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0005',
    'demo-yordanos-kebede',
    'demo',
    'Yordanos Kebede',
    'Placeholder profile for the Mäkdäs shareable handle demo. Swap in your portrait, bio, contact links, and works when you claim your own @handle.',
    'b. 1992, Hawassa',
    'Ceramics / Installation',
    'This is what your link-in-bio could look like.',
    '2018 — present',
    'https://images.unsplash.com/photo-1494790108377-be9c29b29330?w=800&q=80',
    FALSE,
    'Addis Ababa',
    '@makdas.demo',
    'https://t.me/makdas_demo',
    'approved',
    NOW(),
    NOW()
) ON CONFLICT (id) DO NOTHING;

-- Roster works (prototype archive.ts)
UPDATE art_posts SET
    title = 'Ascension, after the rains',
    description = 'Oil and gold leaf on linen.',
    medium = 'Oil on linen',
    year = 2023,
    dimensions = '180 × 150 cm',
    city = 'Addis Ababa',
    style = 'Abstract',
    featured_acquisition = TRUE,
    palette = ARRAY['#1a2f6b', '#d94e1f', '#f5d76e', '#f1ead8'],
    published_at = NOW() - INTERVAL '3 days',
    updated_at = NOW()
WHERE id = 'cccccccc-cccc-cccc-cccc-cccccccc0001';

UPDATE art_posts SET
    title = 'Four Saints in Indigo',
    description = 'Tempera on board.',
    medium = 'Tempera on board',
    year = 2024,
    dimensions = '60 × 45 cm',
    city = 'Berlin',
    style = 'Figurative',
    featured_acquisition = TRUE,
    palette = ARRAY['#1a2f6b', '#c2410c', '#f5d76e', '#e8d5b0'],
    published_at = NOW() - INTERVAL '7 days',
    updated_at = NOW()
WHERE id = 'cccccccc-cccc-cccc-cccc-cccccccc0002';

UPDATE art_posts SET
    title = 'Fidäl, dissolving',
    description = 'Mixed media on jute.',
    medium = 'Mixed media on jute',
    year = 2024,
    dimensions = '90 × 70 cm',
    city = 'Addis Ababa',
    style = 'Text-based',
    featured_acquisition = FALSE,
    palette = ARRAY['#b91c1c', '#f59e0b', '#f5e6c0', '#1c1917'],
    published_at = NOW() - INTERVAL '14 days',
    updated_at = NOW()
WHERE id = 'cccccccc-cccc-cccc-cccc-cccccccc0003';

UPDATE art_posts SET
    title = 'Conversations in Pink',
    description = 'Acrylic on canvas.',
    medium = 'Acrylic on canvas',
    year = 2023,
    dimensions = '150 × 120 cm',
    city = 'Brooklyn',
    style = 'Abstract',
    featured_acquisition = FALSE,
    palette = ARRAY['#ec4899', '#0d9488', '#0a0a0a', '#f97316'],
    published_at = NOW() - INTERVAL '21 days',
    updated_at = NOW()
WHERE id = 'cccccccc-cccc-cccc-cccc-cccccccc0004';

INSERT INTO art_posts (
    id, artist_id, title, description, medium, year, dimensions, city, style,
    featured_acquisition, palette, status, published_at, created_at, updated_at
) VALUES
    (
        'cccccccc-cccc-cccc-cccc-cccccccc0005',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0003',
        'Three Squares for Meskel',
        'Acrylic on canvas.',
        'Acrylic on canvas',
        2024,
        '120 × 90 cm',
        'Addis Ababa',
        'Geometric',
        FALSE,
        ARRAY['#d62828', '#fcbf49', '#0a0a0a', '#f1ead8'],
        'published',
        NOW() - INTERVAL '10 days',
        NOW(),
        NOW()
    ),
    (
        'cccccccc-cccc-cccc-cccc-cccccccc0006',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0001',
        'Field study, Entoto',
        'Oil and gold leaf.',
        'Oil and gold leaf',
        2022,
        '100 × 80 cm',
        'Addis Ababa',
        'Landscape',
        FALSE,
        ARRAY['#2d4a2b', '#c47a3d', '#f4c430', '#e8e1c6'],
        'published',
        NOW() - INTERVAL '18 days',
        NOW(),
        NOW()
    ),
    (
        'cccccccc-cccc-cccc-cccc-cccccccc0007',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0005',
        'Vessel study I',
        'Stoneware vessel study.',
        'Stoneware',
        2024,
        '40 × 30 cm',
        'Addis Ababa',
        'Ceramic',
        FALSE,
        ARRAY['#0d9488', '#f97316', '#1c1917', '#f5e6c0'],
        'published',
        NOW() - INTERVAL '5 days',
        NOW(),
        NOW()
    ),
    (
        'cccccccc-cccc-cccc-cccc-cccccccc0008',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0005',
        'Mercato floor tiles',
        'Glazed ceramic installation.',
        'Glazed ceramic',
        2023,
        'Installation',
        'Addis Ababa',
        'Installation',
        FALSE,
        ARRAY['#d62828', '#fcbf49', '#0a0a0a', '#f1ead8'],
        'published',
        NOW() - INTERVAL '12 days',
        NOW(),
        NOW()
    ),
    (
        'cccccccc-cccc-cccc-cccc-cccccccc0009',
        'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0005',
        'Hawassa light',
        'Terracotta sculpture.',
        'Terracotta',
        2022,
        '50 × 40 cm',
        'Hawassa',
        'Sculpture',
        FALSE,
        ARRAY['#c2410c', '#f5d76e', '#1a2f6b', '#e8d5b0'],
        'published',
        NOW() - INTERVAL '25 days',
        NOW(),
        NOW()
    )
ON CONFLICT (id) DO NOTHING;

INSERT INTO art_post_media (id, art_post_id, url, mime_type, sort_order) VALUES
    ('dddddddd-dddd-dddd-dddd-dddddddd0005', 'cccccccc-cccc-cccc-cccc-cccccccc0005', 'https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200&q=80', 'image/jpeg', 0),
    ('dddddddd-dddd-dddd-dddd-dddddddd0006', 'cccccccc-cccc-cccc-cccc-cccccccc0006', 'https://images.unsplash.com/photo-1579783902614-a3fb3927b6a5?w=1200&q=80', 'image/jpeg', 0),
    ('dddddddd-dddd-dddd-dddd-dddddddd0007', 'cccccccc-cccc-cccc-cccc-cccccccc0007', 'https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200&q=80', 'image/jpeg', 0),
    ('dddddddd-dddd-dddd-dddd-dddddddd0008', 'cccccccc-cccc-cccc-cccc-cccccccc0008', 'https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200&q=80', 'image/jpeg', 0),
    ('dddddddd-dddd-dddd-dddd-dddddddd0009', 'cccccccc-cccc-cccc-cccc-cccccccc0009', 'https://images.unsplash.com/photo-1578301978693-85fa9d0ae59c?w=1200&q=80', 'image/jpeg', 0)
ON CONFLICT (id) DO NOTHING;

UPDATE art_post_media SET url = 'https://images.unsplash.com/photo-1547891654-e66ed7ebb968?w=1200&q=80' WHERE art_post_id = 'cccccccc-cccc-cccc-cccc-cccccccc0001';
UPDATE art_post_media SET url = 'https://images.unsplash.com/photo-1578301978693-85fa9d0ae59c?w=1200&q=80' WHERE art_post_id = 'cccccccc-cccc-cccc-cccc-cccccccc0002';
UPDATE art_post_media SET url = 'https://images.unsplash.com/photo-1579783902614-a3fb3927b6a5?w=1200&q=80' WHERE art_post_id = 'cccccccc-cccc-cccc-cccc-cccccccc0003';
UPDATE art_post_media SET url = 'https://images.unsplash.com/photo-1541961017774-22349e4a1262?w=1200&q=80' WHERE art_post_id = 'cccccccc-cccc-cccc-cccc-cccccccc0004';

-- Wiki articles (prototype wiki.ts)
INSERT INTO articles (
    id, slug, title, body, category, excerpt, reading_time, difficulty, verified, contributors,
    status, author_id, created_at, updated_at
) VALUES
    (
        'eeeeeeee-eeee-eeee-eeee-eeeeeeee0011',
        'eipa-registration',
        'Registering your work with the EIPA',
        'Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring. Register early to establish a clear record of your work.',
        'Legal',
        'Step-by-step on filing with the Ethiopian Intellectual Property Authority — fees, forms, and what to bring.',
        8,
        'Beginner',
        TRUE,
        11,
        'published',
        '33333333-3333-3333-3333-333333333333',
        NOW(),
        NOW()
    ),
    (
        'eeeeeeee-eeee-eeee-eeee-eeeeeeee0012',
        'addis-pigment-sources',
        'Where to buy pigments, linen and stretchers in Addis',
        'A working list of Piassa, Mercato and Bole suppliers — what they stock, who imports, and price ranges. Updated by artists who buy materials weekly.',
        'Materials',
        'A working list of Piassa, Mercato and Bole suppliers — what they stock, who imports, and price ranges.',
        12,
        'Beginner',
        TRUE,
        23,
        'published',
        '33333333-3333-3333-3333-333333333333',
        NOW(),
        NOW()
    ),
    (
        'eeeeeeee-eeee-eeee-eeee-eeeeeeee0013',
        'pricing-local-vs-international',
        'Pricing commissions: local vs. international clients',
        'How to set a base rate in ETB and USD, account for FX volatility, and avoid common underpricing traps when working across borders.',
        'Pricing',
        'How to set a base rate in ETB and USD, account for FX volatility, and avoid common underpricing traps.',
        15,
        'Intermediate',
        FALSE,
        7,
        'published',
        '33333333-3333-3333-3333-333333333333',
        NOW(),
        NOW()
    ),
    (
        'eeeeeeee-eeee-eeee-eeee-eeeeeeee0014',
        'shipping-works-abroad',
        'Shipping works abroad from Addis',
        'Crating, customs paperwork, DHL vs. freight forwarders, and what collectors actually pay for when importing Ethiopian art.',
        'Distribution',
        'Crating, customs paperwork, DHL vs. freight forwarders, and what collectors actually pay for.',
        10,
        'Beginner',
        FALSE,
        5,
        'published',
        '33333333-3333-3333-3333-333333333333',
        NOW(),
        NOW()
    ),
    (
        'eeeeeeee-eeee-eeee-eeee-eeeeeeee0015',
        'fair-use-amharic',
        'Fair use and licensing in an Ethiopian context',
        'What Ethiopian copyright law actually says about derivative works, sampling, and reference photography.',
        'Legal',
        'What Ethiopian copyright law actually says about derivative works, sampling, and reference photography.',
        14,
        'Advanced',
        TRUE,
        14,
        'published',
        '33333333-3333-3333-3333-333333333333',
        NOW(),
        NOW()
    )
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    body = EXCLUDED.body,
    category = EXCLUDED.category,
    excerpt = EXCLUDED.excerpt,
    reading_time = EXCLUDED.reading_time,
    difficulty = EXCLUDED.difficulty,
    verified = EXCLUDED.verified,
    contributors = EXCLUDED.contributors,
    status = EXCLUDED.status,
    updated_at = NOW();

UPDATE articles SET
    title = 'Reading a gallery consignment contract',
    body = 'Commission splits, insurance, exclusivity windows, and the clauses Ethiopian artists keep getting burned by.',
    category = 'Contracts',
    excerpt = 'Commission splits, insurance, exclusivity windows, and the clauses Ethiopian artists keep getting burned by.',
    reading_time = 18,
    difficulty = 'Advanced',
    verified = TRUE,
    contributors = 9,
    updated_at = NOW()
WHERE slug = 'gallery-contracts';

-- Events (prototype events.ts)
DELETE FROM events WHERE id IN (
    'ffffffff-ffff-ffff-ffff-ffffffff0001',
    'ffffffff-ffff-ffff-ffff-ffffffff0002',
    'ffffffff-ffff-ffff-ffff-ffffffff0003'
);

INSERT INTO events (
    id, title, description, source_url, slug, event_type, venue, city,
    starts_at, scraped_at, status, created_at, updated_at
) VALUES
    (
        'ffffffff-ffff-ffff-ffff-ffffffff0011',
        'After the Rains — Tewodros Hailu solo',
        E'Tewodros Hailu''s first solo at Addis Fine Art in three years — nine new pigment-soaked linens hung unstretched, pinned like drying laundry. The work circles residue: walls, weather, and the memory of both.\n\nOpening night includes a walkthrough with the artist at 19:00. Works remain on view through August.',
        'https://makdas.example/events/after-the-rains-tewodros-hailu',
        'after-the-rains-tewodros-hailu',
        'Opening',
        'Addis Fine Art',
        'Addis Ababa',
        '2026-06-27 18:00:00+03',
        NOW(),
        'approved',
        NOW(),
        NOW()
    ),
    (
        'ffffffff-ffff-ffff-ffff-ffffffff0012',
        'Tobiya Poetic Jazz Night',
        E'A monthly evening of Amharic poetry, live jazz, and open mic at Fendika. This month''s theme: migration and return.\n\nDoors at 20:00. Entry 150 ETB at the door.',
        'https://makdas.example/events/tobiya-poetic-jazz-night',
        'tobiya-poetic-jazz-night',
        'Poetry',
        'Fendika Cultural Center',
        'Addis Ababa',
        '2026-06-28 20:00:00+03',
        NOW(),
        'approved',
        NOW(),
        NOW()
    ),
    (
        'ffffffff-ffff-ffff-ffff-ffffffff0013',
        'Curator''s Walkthrough — Skunder''s Cosmos',
        E'A guided tour of the Skunder Boghossian retrospective with Alle School faculty. Focus on diasporic dream-spaces and the artist''s chromatic grammar.\n\nFree with museum admission. Meet in the main hall.',
        'https://makdas.example/events/curators-walkthrough-skunders-cosmos',
        'curators-walkthrough-skunders-cosmos',
        'Talk',
        'Modern Art Museum / Gebre Kristos Desta Center',
        'Addis Ababa',
        '2026-07-02 15:00:00+03',
        NOW(),
        'approved',
        NOW(),
        NOW()
    ),
    (
        'ffffffff-ffff-ffff-ffff-ffffffff0014',
        'Design Pop-up: Habesha Futures',
        E'A one-weekend pop-up at Zoma Museum featuring textile designers, furniture makers, and illustrators reimagining Ethiopian craft for contemporary interiors.\n\nSaturday and Sunday, 10:00–18:00.',
        'https://makdas.example/events/design-pop-up-habesha-futures',
        'design-pop-up-habesha-futures',
        'Pop-up',
        'Zoma Museum',
        'Addis Ababa',
        '2026-07-05 10:00:00+03',
        NOW(),
        'approved',
        NOW(),
        NOW()
    ),
    (
        'ffffffff-ffff-ffff-ffff-ffffffff0015',
        'Yenegew Sew — Theatre revival',
        E'A revival of the classic Ethiopian play at the National Theatre — directed by a new generation of Alle School graduates. Subtitles in English for select performances.\n\nEvening shows Thursday through Sunday.',
        'https://makdas.example/events/yenegew-sew-theatre-revival',
        'yenegew-sew-theatre-revival',
        'Theatre',
        'National Theatre',
        'Addis Ababa',
        '2026-07-09 19:30:00+03',
        NOW(),
        'approved',
        NOW(),
        NOW()
    ),
    (
        'ffffffff-ffff-ffff-ffff-ffffffff0016',
        'Goethe Film Series: Diaspora Editions',
        E'Monthly screening of contemporary African and diaspora cinema, followed by a moderated discussion. This edition features work from Ethiopian filmmakers in Berlin and London.\n\nSeating is limited — arrive early.',
        'https://makdas.example/events/goethe-film-series-diaspora-editions',
        'goethe-film-series-diaspora-editions',
        'Screening',
        'Goethe-Institut',
        'Addis Ababa',
        '2026-07-11 18:30:00+03',
        NOW(),
        'approved',
        NOW(),
        NOW()
    ),
    (
        'ffffffff-ffff-ffff-ffff-ffffffff0017',
        'Bahir Dar Lake Sessions — Open studios',
        E'Open studios along Lake Tana with painters, weavers, and ceramicists working in a shared boathouse space. A chance to meet artists outside the Addis circuit.\n\nFree entry. Works available for direct purchase from studios.',
        'https://makdas.example/events/bahir-dar-lake-sessions-open-studios',
        'bahir-dar-lake-sessions-open-studios',
        'Pop-up',
        'Lake Tana Boathouse',
        'Bahir Dar',
        '2026-07-13 11:00:00+03',
        NOW(),
        'approved',
        NOW(),
        NOW()
    )
ON CONFLICT (id) DO NOTHING;

-- +goose Down
DELETE FROM events WHERE id BETWEEN 'ffffffff-ffff-ffff-ffff-ffffffff0011' AND 'ffffffff-ffff-ffff-ffff-ffffffff0017';
DELETE FROM art_post_media WHERE art_post_id IN (
    'cccccccc-cccc-cccc-cccc-cccccccc0005',
    'cccccccc-cccc-cccc-cccc-cccccccc0006',
    'cccccccc-cccc-cccc-cccc-cccccccc0007',
    'cccccccc-cccc-cccc-cccc-cccccccc0008',
    'cccccccc-cccc-cccc-cccc-cccccccc0009'
);
DELETE FROM art_posts WHERE id IN (
    'cccccccc-cccc-cccc-cccc-cccccccc0005',
    'cccccccc-cccc-cccc-cccc-cccccccc0006',
    'cccccccc-cccc-cccc-cccc-cccccccc0007',
    'cccccccc-cccc-cccc-cccc-cccccccc0008',
    'cccccccc-cccc-cccc-cccc-cccccccc0009'
);
DELETE FROM artist_profiles WHERE id = 'bbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbb0005';
DELETE FROM users WHERE id = 'aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaa0005';
DELETE FROM articles WHERE slug IN (
    'eipa-registration',
    'addis-pigment-sources',
    'pricing-local-vs-international',
    'shipping-works-abroad',
    'fair-use-amharic'
);
