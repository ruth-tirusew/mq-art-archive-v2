-- +goose Up
CREATE TABLE media_assets (
    id UUID PRIMARY KEY,
    owner_user_id UUID NOT NULL REFERENCES users(id),
    public_id TEXT NOT NULL UNIQUE,
    secure_url TEXT NOT NULL,
    resource_type TEXT NOT NULL DEFAULT 'image' CHECK (resource_type = 'image'),
    width INT,
    height INT,
    bytes INT,
    folder TEXT,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);
CREATE INDEX idx_media_assets_owner ON media_assets(owner_user_id);
ALTER TABLE art_post_media ADD COLUMN media_asset_id UUID REFERENCES media_assets(id) ON DELETE SET NULL;
ALTER TABLE artist_profiles ADD COLUMN portrait_media_asset_id UUID REFERENCES media_assets(id) ON DELETE SET NULL;

-- +goose Down
ALTER TABLE artist_profiles DROP COLUMN IF EXISTS portrait_media_asset_id;
ALTER TABLE art_post_media DROP COLUMN IF EXISTS media_asset_id;
DROP TABLE IF EXISTS media_assets;
