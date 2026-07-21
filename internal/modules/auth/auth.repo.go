package auth

import (
	"context"
	"database/sql"
	"errors"
	"mailForgeApi/internal/apperrors"
	"mailForgeApi/internal/models"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
)

// define the interfaces which will be implemented by the Service

type AuthRepo interface {
	CreateUser(ctx context.Context, user *models.User) error
	FindByPublicID(ctx context.Context, publicId string) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	UpdateLastLogin(ctx context.Context, publicId string) error
	IncrementFailedAttempts(ctx context.Context, publicId string) error
}

// define the struct which wont be exposed to the service
type authRepository struct {
	db *bun.DB
}

// write the constructor
func NewAuthRepository(db *bun.DB) AuthRepo {
	return &authRepository{
		db: db,
	}
}

// write the methods that will be based on the contructor

func (r *authRepository) CreateUser(ctx context.Context, user *models.User) error {
	// user

	// Insert with ON CONFLICT handling
	_, err := r.db.NewInsert().
		Model(user).
		Exec(ctx)

	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return apperrors.ErrDuplicate
	}
	return err
}

func (r *authRepository) FindByPublicID(ctx context.Context, publicId string) (*models.User, error) {
	user := new(models.User)

	err := r.db.NewSelect().Model(user).Where("public_id = ?", publicId).
		Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}

	return user, nil
}

func (r *authRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {

	user := new(models.User)

	err := r.db.NewSelect().Model(user).Where("email = ?", email).Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return user, nil
}

func (r *authRepository) UpdateLastLogin(ctx context.Context, publicId string) error {

	result, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("last_login_at = ?", time.Now().UTC()).
		Set("failed_login_attempts = 0").
		Where("public_id = ?", publicId).Exec(ctx)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}

func (r *authRepository) IncrementFailedAttempts(ctx context.Context, publicId string) error {
	result, err := r.db.NewUpdate().
		Model((*models.User)(nil)).
		Set("failed_login_attempts = failed_login_attempts + 1").
		Where("public_id = ?", publicId).
		Exec(ctx)

	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}

	if rows == 0 {
		return apperrors.ErrNotFound
	}

	return nil
}
