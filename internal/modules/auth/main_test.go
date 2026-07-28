package auth

import (
	"crypto/rsa"
	"log"
	"os"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"

	testutils "mailForgeApi/internal/testUtils"
)

var (
	testDB     *bun.DB
	testRedis  *redis.Client
	testPrvKey *rsa.PrivateKey
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

	testRedis, err = testutils.SetupTestRedis()
	if err != nil {
		log.Fatalf("failed to connect to redis: %v", err)
	}

	testPrvKey, _, err = testutils.TestKeyPair()
	if err != err {
		log.Fatalf("failed to generate private key: %v", err)
	}

	code := m.Run()

	_ = testDB.Close()

	os.Exit(code)
}
