package testutils

import (
	"context"

	"github.com/uptrace/bun"

	"mailForgeApi/internal/migrations"
)

func RunTestMigrations(db *bun.DB) error {
	return migrations.Run(context.Background(), db)
}
