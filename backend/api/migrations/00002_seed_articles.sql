-- +goose Up
INSERT INTO articles (id, slug, title, body, status, author_id, created_at, updated_at)
VALUES (
    '11111111-1111-1111-1111-111111111111',
    'welcome-to-mq',
    'Welcome to mq',
    'A community wiki and archive for Ethiopian artists.',
    'published',
    '22222222-2222-2222-2222-222222222222',
    NOW(),
    NOW()
);

-- +goose Down
DELETE FROM articles WHERE slug = 'welcome-to-mq';
