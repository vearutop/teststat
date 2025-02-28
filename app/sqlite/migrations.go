// Package sqlite provides migrations.
package sqlite

import (
	"embed"
)

// migrations provide database migrations.
//
//go:embed migrations/*.sql
var migrations embed.FS
