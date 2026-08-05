package testutils

import (
	"context"
	"mailForgeApi/internal/models"

	"github.com/uptrace/bun"
)

func CreateTestUser(db *bun.DB) (*models.User, error) {
	user, err := BuildTestUser()
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	_, err = db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}

func CreateTestList(
	db *bun.DB,
	userID uint64,
) (*models.List, error) {

	list, err := BuildTestList()
	if err != nil {
		return nil, err
	}

	list.UserID = userID

	ctx := context.Background()

	_, err = db.NewInsert().
		Model(list).
		Exec(ctx)

	if err != nil {
		return nil, err
	}

	return list, nil
}

func CreateTestUserWithPassword(db *bun.DB, plaintextPassword string) (*models.User, error) {
	user, err := BuildTestUserWithPassword(plaintextPassword)
	if err != nil {
		return nil, err
	}

	ctx := context.Background()

	_, err = db.NewInsert().Model(user).Exec(ctx)
	if err != nil {
		return nil, err
	}

	return user, nil
}
