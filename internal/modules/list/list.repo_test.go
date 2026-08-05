package list

import (
	"context"
	"mailForgeApi/internal/models"
	"mailForgeApi/internal/shared/apperrors"
	testutils "mailForgeApi/internal/testUtils"
	"testing"
	"time"

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

	list, err := testutils.BuildTestList("wedding invites")
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
	require.NotZero(t, stored.CreatedAt)
	assert.Equal(t, list.PublicID, stored.PublicID)
	assert.Equal(t, list.Name, stored.Name)
	assert.Equal(t, list.Description, stored.Description)
	assert.Equal(t, list.Status, stored.Status)
}

func TestFindAll_ReturnsOnlyUsersLists(t *testing.T) {
	// setup
	repo := setup(t)
	ctx := context.Background()

	// arrange
	userA, err := testutils.CreateTestUser(testDB)
	require.NoError(t, err)

	userB, err := testutils.CreateTestUser(testDB)
	require.NoError(t, err)

	list1, err := testutils.CreateTestList(testDB, userA.ID, "school stuff")
	require.NoError(t, err)

	list2, err := testutils.CreateTestList(testDB, userA.ID, "graphic customers")
	require.NoError(t, err)

	otherList, err := testutils.CreateTestList(testDB, userB.ID, "fabric customers")
	require.NoError(t, err)

	// act
	lists, err := repo.FindAll(ctx, userA.ID)
	require.NoError(t, err)

	// assert
	require.Len(t, lists, 2)

	for _, list := range lists {
		assert.Equal(t, userA.ID, list.UserID)
	}

	ids := []string{
		lists[0].PublicID,
		lists[1].PublicID,
	}

	assert.Contains(t, ids, list1.PublicID)
	assert.Contains(t, ids, list2.PublicID)

	assert.NotContains(t, ids, otherList.PublicID)
}

func TestFindByID_DoesNotReturnAnotherUsersList(t *testing.T) {
	// setup
	repo := setup(t)
	ctx := context.Background()

	// arrange
	userA, err := testutils.CreateTestUser(testDB)
	require.NoError(t, err)

	list1, err := testutils.CreateTestList(testDB, userA.ID, "web dev")
	require.NoError(t, err)

	userB, err := testutils.CreateTestUser(testDB)
	require.NoError(t, err)

	list2, err := testutils.CreateTestList(testDB, userB.ID, "eminem lovers")
	require.NoError(t, err)

	// act
	list, err := repo.FindByPublicID(ctx, userA.ID, list2.PublicID)

	// assert
	require.Error(t, err)
	assert.Nil(t, list)
	assert.NotEqual(t, list, list1)
	assert.NotEqual(t, list, list2)

}

func TestUpdateList(t *testing.T) {
	// setup
	repo := setup(t)
	ctx := context.Background()

	// arrange
	user, err := testutils.CreateTestUser(testDB)
	require.NoError(t, err)

	original, err := testutils.CreateTestList(testDB, user.ID, "old name")
	require.NoError(t, err)

	updates := &models.List{
		Name:        "new name",
		Description: "updated description",
		UpdatedAt:   time.Now().UTC(),
	}

	// act
	err = repo.Update(ctx, user.ID, original.PublicID, updates)
	require.NoError(t, err)

	// verify independently — not through the repo method under test
	stored := new(models.List)
	err = testDB.NewSelect().Model(stored).Where("public_id = ?", original.PublicID).Scan(ctx)
	require.NoError(t, err)

	// assert
	assert.Equal(t, "new name", stored.Name)
	assert.Equal(t, "updated description", stored.Description)
	assert.Equal(t, original.PublicID, stored.PublicID) // unchanged — Update must never touch identity
	assert.Equal(t, original.Status, stored.Status)     // unchanged — Update must never touch status; that's Archive's job
}

func TestUpdateList_ReturnsNotFound_ForAnotherUsersList(t *testing.T) {
	// setup
	repo := setup(t)
	ctx := context.Background()

	// arrange
	userA, err := testutils.CreateTestUser(testDB)
	require.NoError(t, err)

	userB, err := testutils.CreateTestUser(testDB)
	require.NoError(t, err)

	list, err := testutils.CreateTestList(testDB, userB.ID, "userB's list")
	require.NoError(t, err)

	updates := &models.List{
		Name:      "hijacked name",
		UpdatedAt: time.Now().UTC(),
	}

	// act — userA tries to update userB's list
	err = repo.Update(ctx, userA.ID, list.PublicID, updates)

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)

	// confirm the row was genuinely untouched, not just that we got an error
	stored := new(models.List)
	dbErr := testDB.NewSelect().Model(stored).Where("public_id = ?", list.PublicID).Scan(ctx)
	require.NoError(t, dbErr)
	assert.Equal(t, "userB's list", stored.Name)
}

func TestArchiveList(t *testing.T) {
	// setup
	repo := setup(t)
	ctx := context.Background()

	// arrange
	user, err := testutils.CreateTestUser(testDB)
	require.NoError(t, err)

	list, err := testutils.CreateTestList(testDB, user.ID, "to be archived")
	require.NoError(t, err)
	require.Equal(t, "active", list.Status) // sanity check on the fixture's default

	// act
	err = repo.Archive(ctx, user.ID, list.PublicID)
	require.NoError(t, err)

	// verify independently
	stored := new(models.List)
	err = testDB.NewSelect().Model(stored).Where("public_id = ?", list.PublicID).Scan(ctx)
	require.NoError(t, err)

	// assert
	assert.Equal(t, "archived", stored.Status)
	assert.Equal(t, list.Name, stored.Name) // row still exists with data intact — this is a status flip, not a delete
}

func TestArchiveList_ReturnsNotFound_ForAnotherUsersList(t *testing.T) {
	// setup
	repo := setup(t)
	ctx := context.Background()

	// arrange
	userA, err := testutils.CreateTestUser(testDB)
	require.NoError(t, err)

	userB, err := testutils.CreateTestUser(testDB)
	require.NoError(t, err)

	list, err := testutils.CreateTestList(testDB, userB.ID, "userB's list")
	require.NoError(t, err)

	// act — userA tries to archive userB's list
	err = repo.Archive(ctx, userA.ID, list.PublicID)

	// assert
	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)

	stored := new(models.List)
	dbErr := testDB.NewSelect().Model(stored).Where("public_id = ?", list.PublicID).Scan(ctx)
	require.NoError(t, dbErr)
	assert.Equal(t, "active", stored.Status) // untouched
}
