package auth

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	testutils "mailForgeApi/internal/testUtils"
	tokens "mailForgeApi/pkg/token"
)

func setup(t *testing.T) AuthRepo {

	t.Helper()
	// ctx := context.Background()

	err := testutils.CleanTables(t.Context(), testDB)

	require.NoError(t, err)

	return NewAuthRepository(testDB)
}

func setupService(t *testing.T) *Service {

	t.Helper()
	err := testutils.CleanTables(t.Context(), testDB)
	require.NoError(t, err)

	repo := NewAuthRepository(testDB)
	refreshManager := tokens.NewRefreshTokenManager(testRedis, 7*24*time.Hour)
	privatekey, _, _ := testutils.TestKeyPair()

	return NewService(repo, refreshManager, privatekey, time.Hour)
}
