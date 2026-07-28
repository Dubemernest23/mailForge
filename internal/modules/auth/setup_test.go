package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	testutils "mailForgeApi/internal/testUtils"
	tokens "mailForgeApi/pkg/token"
	// "mailForgeApi/internal/testutils"
)

func setup(t *testing.T) AuthRepo {

	t.Helper()

	err := testutils.CleanTables(testDB)

	require.NoError(t, err)

	return NewAuthRepository(testDB)
}

func setupService(t *testing.T) *Service {

	t.Helper()
	err := testutils.CleanTables(testDB)
	require.NoError(t, err)

	repo := NewAuthRepository(testDB)
	refreshManager := tokens.NewRefreshTokenManager(testRedis, 7*24*time.Hour)
	privatekey, _, _ := testutils.TestKeyPair()

	return NewService(repo, refreshManager, privatekey, time.Hour)
}
