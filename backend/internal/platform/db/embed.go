package db

import "embed"

// MigrationsFS holds the platform core's Goose migrations (the Mode B session
// store). The registry merges it into the one boot-time migration sequence under
// the "platform" source; its 01xxx prefix runs before the feature modules.
//
//go:embed migrations/*.sql
var MigrationsFS embed.FS
