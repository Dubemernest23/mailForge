package testutils

import (
	"database/sql"
	"fmt"
	"os"

	"github.com/uptrace/bun"
	"github.com/uptrace/bun/dialect/mysqldialect"
)

func SetupTestDB() (*bun.DB, error) {
	dsn := os.Getenv("TEST_DB_DSN")
	if dsn == "" {
		dsn = "root:secret@tcp(localhost:3307)/mailforge_test?parseTime=true&loc=UTC"
	}

	sqldb, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to open test database: %w", err)
	}

	if err := sqldb.Ping(); err != nil {
		_ = sqldb.Close()
		return nil, fmt.Errorf("failed to reach database: %w", err)
	}

	return bun.NewDB(sqldb, mysqldialect.New()), nil
}
