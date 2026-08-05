package list

import (
	"log"
	testutils "mailForgeApi/internal/testUtils"
	"os"
	"testing"

	"github.com/uptrace/bun"
)

var (
	testDB *bun.DB
)

func TestMain(m *testing.M) {
	var err error

	testDB, err = testutils.SetupTestDB()

	if err != nil {
		log.Fatalf("failed to connect to test database: %v", err)
	}

	if err := testutils.RunTestMigrations(testDB); err != nil {
		log.Fatalf("failed to run migrations: %v", err)
	}

	code := m.Run()

	_ = testDB.Close()

	os.Exit(code)
}
