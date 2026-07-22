package testutils

import (
	"context"

	"github.com/uptrace/bun"
)

func CleanTables(db *bun.DB) error {

	ctx := context.Background()

	_, err := db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 0")

	if err != nil {
		return err
	}

	defer func() {
		_, _ = db.ExecContext(ctx, "SET FOREIGN_KEY_CHECKS = 1")
	}()

	_, err = db.ExecContext(ctx, `
		TRUNCATE TABLE users;
	`)

	if err != nil {
		return err
	}

	return nil
}
