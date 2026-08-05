package list

import (
	"context"
	"mailForgeApi/internal/models"
	testutils "mailForgeApi/internal/testUtils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCreateList(t *testing.T) {
	//setup
	repo := setup(t)
	ctx := context.Background()

	// arrange

	user, err := testutils.CreateTestUser(testDB)
	require.NoError(t, err)

	list, err := testutils.BuildTestList()
	require.NoError(t, err)

	// act
	err = repo.CreateList(ctx, user.ID, list)
	require.NoError(t, err)

	// verify insertion worked
	stored := new(models.List)
	err = testDB.NewSelect().Model(stored).Where("public_id = ?", list.PublicID).Scan(ctx)
	require.NoError(t, err)

	// assert
	require.NotZero(t, stored.ID)
	assert.Equal(t, list.PublicID, stored.PublicID)
	assert.Equal(t, list.Name, stored.Name)
	assert.Equal(t, list.Description, stored.Description)
	assert.Equal(t, list.Status, stored.Status)
}
