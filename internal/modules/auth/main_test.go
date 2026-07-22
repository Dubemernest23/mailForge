package auth

import (
	"log"
	"os"
	"testing"

	"github.com/uptrace/bun"

	testutils "mailForgeApi/internal/testUtils"
	// "mailForgeApi/internal/testutils"
)

var testDB *bun.DB

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
