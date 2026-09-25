// Package migrations embeds the goose SQL migration files so the API binary can apply
// them on boot without a shell or separate goose install in the distroless runtime image.
package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
