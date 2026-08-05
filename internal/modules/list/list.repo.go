package list

import (
	"context"
	"database/sql"
	"errors"
	"mailForgeApi/internal/models"
	"mailForgeApi/internal/shared/apperrors"

	"github.com/go-sql-driver/mysql"
	"github.com/uptrace/bun"
)

// FindAll(ctx, userID)`, `FindByID(ctx, userID, publicID)`, `Update`, `Archive`
type ListRepoInterface interface {
	CreateList(ctx context.Context, userID uint64, list *models.List) error
	FindAll(ctx context.Context, userID uint64) ([]models.List, error)
	FindByPublicID(ctx context.Context, userID uint64, publicID string) (*models.List, error)
	Update(ctx context.Context, userID uint64, publicID string, updates *models.List) error
	Archive(ctx context.Context, userID uint64, publicID string) error
}

type listRepository struct {
	db *bun.DB
}

func NewListRepository(db *bun.DB) ListRepoInterface {
	return &listRepository{
		db: db,
	}
}

func (r *listRepository) CreateList(ctx context.Context, userID uint64, list *models.List) error {
	list.UserID = userID

	_, err := r.db.NewInsert().Model(list).Exec(ctx)
	var mysqlErr *mysql.MySQLError
	if errors.As(err, &mysqlErr) && mysqlErr.Number == 1062 {
		return apperrors.ErrDuplicate
	}
	return err
}

func (r *listRepository) FindAll(ctx context.Context, userID uint64) ([]models.List, error) {
	var lists []models.List

	err := r.db.NewSelect().Model(&lists).Where("user_id = ?", userID).Order("created_at DESC").Scan(ctx)

	if err != nil {
		return nil, err
	}

	return lists, nil
}

func (r *listRepository) FindByPublicID(ctx context.Context, userID uint64, publicID string) (*models.List, error) {
	list := new(models.List)

	err := r.db.NewSelect().Model(list).Where("user_id = ? AND public_id = ?", userID, publicID).Scan(ctx)

	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, err
	}
	return list, nil
}

func (r *listRepository) Update(ctx context.Context, userID uint64, publicID string, updates *models.List) error {

	// Update with WHERE clause
	res, err := r.db.NewUpdate().
		Model(updates).
		Column("name", "description", "updated_at").
		Where("user_id = ? AND public_id = ?", userID, publicID).
		Exec(ctx)

	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func (r *listRepository) Archive(ctx context.Context, userID uint64, publicID string) error {
	res, err := r.db.NewUpdate().
		Model((*models.List)(nil)).
		Set("status = ?", "archived").
		Where("user_id = ? AND public_id = ?", userID, publicID).
		Exec(ctx)

	if err != nil {
		return err
	}
	return checkRowsAffected(res)
}

func checkRowsAffected(res sql.Result) error {
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return apperrors.ErrNotFound
	}
	return nil
}
