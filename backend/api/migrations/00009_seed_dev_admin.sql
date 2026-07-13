-- +goose Up
INSERT INTO users (id, email, role, created_at, updated_at)
VALUES (
    '00000000-0000-4000-8000-000000000001',
    'admin@mq.local',
    'admin',
    NOW(),
    NOW()
)
ON CONFLICT (email) DO NOTHING;

-- +goose Down
DELETE FROM users WHERE email = 'admin@mq.local';
