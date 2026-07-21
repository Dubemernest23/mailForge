package migrations

import (
	"context"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/migrate"
)

func NewMigrator(db *bun.DB) (*migrate.Migrator, error) {
	m := migrate.NewMigrations()

	if err := m.Discover(SQLMigrations); err != nil {
		return nil, err
	}

	return migrate.NewMigrator(db, m), nil
}

func Run(ctx context.Context, db *bun.DB) error {
	migrator, err := NewMigrator(db)
	if err != nil {
		return err
	}

	if err := migrator.Init(ctx); err != nil {
		return err
	}

	_, err = migrator.Migrate(ctx)
	return err
}
