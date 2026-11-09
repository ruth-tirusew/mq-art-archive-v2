-- +goose Up
-- Bios and artwork descriptions are now rich text: HTML cleaned by the server's sanitizer on every
-- write. Until now they were plain text, so every existing value is converted to the same HTML the
-- editor would have produced (escaped, blank-line-separated paragraphs, single newlines as <br>)
-- and keeps reading correctly.
UPDATE artist_profiles
SET bio = '<p>' || replace(
        regexp_replace(
            replace(replace(replace(replace(trim(bio), '&', '&amp;'), '<', '&lt;'), '>', '&gt;'), E'\r\n', E'\n'),
            E'\n{2,}', '</p><p>', 'g'),
        E'\n', '<br>') || '</p>'
WHERE trim(bio) <> '';

UPDATE art_posts
SET description = '<p>' || replace(
        regexp_replace(
            replace(replace(replace(replace(trim(description), '&', '&amp;'), '<', '&lt;'), '>', '&gt;'), E'\r\n', E'\n'),
            E'\n{2,}', '</p><p>', 'g'),
        E'\n', '<br>') || '</p>'
WHERE trim(description) <> '';

-- +goose Down
-- Not reversible: once edited with formatting, a value can't be turned back into the plain text it
-- came from. The HTML is still readable text, so leaving it in place is harmless.
SELECT 1;
