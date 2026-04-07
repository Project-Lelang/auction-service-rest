package repository

import (
	"context"
	"fmt"
	"time"

	"auction-service/infrastructure"
	"auction-service/model"

	"github.com/Masterminds/squirrel"
)

type UserAccessTokenRepository interface {
	// create
	Insert(ctx context.Context, token *model.UserAccessToken) error

	// read
	GetById(ctx context.Context, id string) (*model.UserAccessToken, error)
	GetByToken(ctx context.Context, token string) (*model.UserAccessToken, error)

	// delete
	Delete(ctx context.Context, token *model.UserAccessToken) error
	DeleteByUserId(ctx context.Context, userId string) error
}

type userAccessTokenRepository struct {
	db          infrastructure.DBTX
	loggerStack infrastructure.LoggerStack
}

func NewUserAccessTokenRepository(db infrastructure.DBTX, loggerStack infrastructure.LoggerStack) UserAccessTokenRepository {
	return &userAccessTokenRepository{
		db:          db,
		loggerStack: loggerStack,
	}
}

func (r *userAccessTokenRepository) tableName() string {
	return model.UserAccessTokenTableName
}

func (r *userAccessTokenRepository) tableAlias() string {
	return "uat"
}

func (r *userAccessTokenRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.tableAlias())
}

func (r *userAccessTokenRepository) f(field string) string {
	return fmt.Sprintf("%s.%s", r.tableAlias(), field)
}

func (r *userAccessTokenRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.UserAccessToken, error) {
	token := model.UserAccessToken{}
	if err := get(r.db, ctx, &token, stmt); err != nil {
		return nil, err
	}
	return &token, nil
}

func (r *userAccessTokenRepository) Insert(ctx context.Context, token *model.UserAccessToken) error {
	if token.CreatedAt.IsZero() {
		token.CreatedAt = time.Now().UTC()
	}
	return insert(r.db, ctx, r.tableName(), token.ToMap())
}

func (r *userAccessTokenRepository) GetById(ctx context.Context, id string) (*model.UserAccessToken, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id})
	return r.getInternal(ctx, stmt)
}

func (r *userAccessTokenRepository) GetByToken(ctx context.Context, token string) (*model.UserAccessToken, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("token"): token})
	return r.getInternal(ctx, stmt)
}

func (r *userAccessTokenRepository) Delete(ctx context.Context, token *model.UserAccessToken) error {
	return destroy(r.db, ctx, r.tableName(), squirrel.Eq{"id": token.Id})
}

func (r *userAccessTokenRepository) DeleteByUserId(ctx context.Context, userId string) error {
	return destroy(r.db, ctx, r.tableName(), squirrel.Eq{"user_id": userId})
}
