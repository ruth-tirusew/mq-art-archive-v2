package postgres

// SQL expressions for the blur hash of the uploaded image at the URL in a given $n placeholder, or
// ” when the URL isn't one of ours. Writes use them so a row's hash always comes from the asset's
// own record, never from the caller. They are constants, not built by a function, so query text
// can never include caller input (see the architecture test).
const (
	blurhashForArg2  = `COALESCE((SELECT m.blurhash FROM media_assets m WHERE m.secure_url = $2 LIMIT 1), '')`
	blurhashForArg3  = `COALESCE((SELECT m.blurhash FROM media_assets m WHERE m.secure_url = $3 LIMIT 1), '')`
	blurhashForArg11 = `COALESCE((SELECT m.blurhash FROM media_assets m WHERE m.secure_url = $11 LIMIT 1), '')`
)
