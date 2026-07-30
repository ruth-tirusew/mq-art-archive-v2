-- +goose Up
CREATE INDEX idx_artist_profiles_search ON artist_profiles USING GIN (
    to_tsvector('simple', COALESCE(display_name,'') || ' ' || COALESCE(bio,'') || ' ' ||
        COALESCE(handle,'') || ' ' || COALESCE(discipline,''))
);
CREATE INDEX idx_art_posts_search ON art_posts USING GIN (
    to_tsvector('simple', COALESCE(title,'') || ' ' || COALESCE(description,'') || ' ' ||
        COALESCE(medium,'') || ' ' || COALESCE(city,'') || ' ' || COALESCE(style,''))
);

-- +goose Down
DROP INDEX IF EXISTS idx_art_posts_search;
DROP INDEX IF EXISTS idx_artist_profiles_search;
