// internal/migrations/embed.go
package migrations

import "embed"

//go:embed *.sql
var SQLMigrations embed.FS
