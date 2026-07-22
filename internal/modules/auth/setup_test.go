package auth

import (
	"testing"

	"github.com/stretchr/testify/require"

	testutils "mailForgeApi/internal/testUtils"
	// "mailForgeApi/internal/testutils"
)

func setup(t *testing.T) AuthRepo {

	t.Helper()

	err := testutils.CleanTables(testDB)

	require.NoError(t, err)

	return NewAuthRepository(testDB)
}
