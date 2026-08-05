package list

import (
	testutils "mailForgeApi/internal/testUtils"
	"testing"

	"github.com/stretchr/testify/require"
)

func setup(t *testing.T) ListRepoInterface {
	t.Helper()

	err := testutils.CleanTables(t.Context(), testDB)
	require.NoError(t, err)

	return NewListRepository(testDB)
}
