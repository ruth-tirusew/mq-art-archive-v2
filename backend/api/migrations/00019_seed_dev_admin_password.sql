-- +goose Up
-- Dev admin password: admin123 (local/dev only — change in non-dev environments)
UPDATE users
SET password_hash = '$2a$10$Tv677q70S11wHwlos.tmMe1.75sqIF2bp7pe2T1Bb7baLf79CxRT.',
    updated_at = NOW()
WHERE email = 'admin@mq.local';

-- +goose Down
UPDATE users
SET password_hash = NULL,
    updated_at = NOW()
WHERE email = 'admin@mq.local';
