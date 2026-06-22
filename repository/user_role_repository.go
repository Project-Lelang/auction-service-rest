package repository

import (
	"context"

	"auction-service/infrastructure"
	"auction-service/model"

	"github.com/Masterminds/squirrel"
)

type UserRoleRepository interface {
	Insert(ctx context.Context, userRole *model.UserRole) error
	Delete(ctx context.Context, userId int64, role string) error
	GetByUserIdAndRole(ctx context.Context, userId int64, role string) (*model.UserRole, error)
	FetchByUserIds(ctx context.Context, userIds []int64) ([]model.UserRole, error)
}

type userRoleRepository struct {
	db infrastructure.DBTX
}

func NewUserRoleRepository(db infrastructure.DBTX) UserRoleRepository {
	return &userRoleRepository{db: db}
}

func (r *userRoleRepository) Insert(ctx context.Context, userRole *model.UserRole) error {
	return defaultInsert(r.db, ctx, userRole)
}

func (r *userRoleRepository) FetchByUserIds(ctx context.Context, userIds []int64) ([]model.UserRole, error) {
	if len(userIds) == 0 {
		return []model.UserRole{}, nil
	}
	stmt := stmtBuilder.Select("*").
		From(model.UserRoleTableName).
		Where(squirrel.Eq{"user_id": userIds})

	roles := []model.UserRole{}
	if err := fetch(r.db, ctx, &roles, stmt); err != nil {
		return nil, err
	}
	return roles, nil
}

func (r *userRoleRepository) Delete(ctx context.Context, userId int64, role string) error {
	stmt := stmtBuilder.Delete(model.UserRoleTableName).
		Where(squirrel.Eq{"user_id": userId, "role": role})
	_, err := exec(r.db, ctx, stmt)
	return err
}

func (r *userRoleRepository) GetByUserIdAndRole(ctx context.Context, userId int64, role string) (*model.UserRole, error) {
	stmt := stmtBuilder.Select("*").
		From(model.UserRoleTableName).
		Where(squirrel.Eq{"user_id": userId, "role": role}).
		Limit(1)

	ur := model.UserRole{}
	if err := get(r.db, ctx, &ur, stmt); err != nil {
		return nil, err
	}
	return &ur, nil
}
