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
