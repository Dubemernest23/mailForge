package auth

import (
	"context"
	"mailForgeApi/internal/apperrors"
	"mailForgeApi/internal/models"
	testutils "mailForgeApi/internal/testUtils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFindUserByPublicId(t *testing.T) {
	// test setup - you are using ctx here instead of in setup_test.go because you want to control the cpntext
	// some test will still use context.WithTimeout() and stuff
	repo := setup(t)
	ctx := context.Background()

	// arrange - why user require() here? it is because if the fixture counldn't be inserted the the test cannot continue
	expectedUser, err := testutils.CreateTestUser(testDB)
	require.NoError(t, err)

	// act - everything b4 this is a setup and everything after is a verification
	actualUser, err := repo.FindByPublicID(ctx, expectedUser.PublicId)
	require.NoError(t, err) // verify the repo didnt return an error

	require.NotNil(t, actualUser)

	assert.Equal(t, expectedUser.ID, actualUser.ID)
	assert.Equal(t, expectedUser.PublicId, actualUser.PublicId)
	assert.Equal(t, expectedUser.Email, actualUser.Email)
	assert.Equal(t, expectedUser.Username, actualUser.Username)
	assert.Equal(t, expectedUser.Role, actualUser.Role)
	assert.Equal(t, expectedUser.FailedLoginAttempts, actualUser.FailedLoginAttempts)
	assert.Equal(t, expectedUser.Status, actualUser.Status)
}

func TestFindByPublicID_NotFound(t *testing.T) {
	repo := setup(t)
	ctx := context.Background()

	// act
	user, err := repo.FindByPublicID(ctx, "this-doesnt-exist-in-test-db-343")

	require.Error(t, err)
	assert.Nil(t, user)
	assert.ErrorIs(t, err, apperrors.ErrNotFound)
}

func TestFindUserByEmail(t *testing.T) {
	// test table
	tests := []struct {
		name        string
		setup       func(t *testing.T) string
		expectedErr error
	}{
		{
			name: "Existing_user",
			setup: func(t *testing.T) string {
				user, err := testutils.CreateTestUser(testDB)

				require.NoError(t, err)
				return user.Email
			},
			expectedErr: nil,
		},
		{
			name: "user_not_found",
			setup: func(t *testing.T) string {
				return "missing-mail@mail.com"
			},
			expectedErr: apperrors.ErrNotFound,
		},
	}

	for _, tc := range tests {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			// setup
			repo := setup(t)
			ctx := context.Background()

			// arrange
			email := tc.setup(t)

			user, err := repo.FindByEmail(ctx, email)
			// assert
			if tc.expectedErr != nil {
				require.Error(t, err)
				assert.ErrorIs(t, err, tc.expectedErr)
				assert.Nil(t, user)
				return
			}

			require.NoError(t, err)
			require.NotNil(t, user)

			assert.Equal(t, email, user.Email)
		})
	}

}

func TestCreateUser(t *testing.T) {
	// arrange
	repo := setup(t)
	ctx := context.Background()

	user, err := testutils.BuildTestUser()
	require.NoError(t, err)

	// act
	err = repo.CreateUser(ctx, user)
	require.NoError(t, err)

	//verification - verify that the insertion actually worked b4 asserting
	stored := new(models.User)

	err = testDB.NewSelect().Model(stored).Where("public_id = ?", user.PublicId).Scan(ctx)
	require.NoError(t, err)

	assert.Equal(t, user.ID, stored.ID)
	assert.Equal(t, user.PublicId, stored.PublicId)
	assert.Equal(t, user.Username, stored.Username)
	assert.Equal(t, user.Email, stored.Email)
	assert.Equal(t, user.PasswordHash, stored.PasswordHash)
	assert.Equal(t, user.Role, stored.Role)
	assert.Equal(t, user.Status, stored.Status)
	assert.Equal(t, user.FailedLoginAttempts, stored.FailedLoginAttempts)
}

func TestCreateUser_DuplicateEmail(t *testing.T) {
	repo := setup(t)
	ctx := context.Background()

	user1, err := testutils.BuildTestUser()
	require.NoError(t, err)

	require.NoError(t, repo.CreateUser(ctx, user1))

	user2, err := testutils.BuildTestUser()
	require.NoError(t, err)

	user2.Email = user1.Email
	err = repo.CreateUser(ctx, user2)

	require.Error(t, err)
	assert.ErrorIs(t, err, apperrors.ErrDuplicate)
}

func TestUpdateLastLogin(t *testing.T) {

}
