package repository

import (
	"context"
	"fmt"
	"time"

	"auction-service/constant"
	"auction-service/infrastructure"
	"auction-service/model"

	"github.com/Masterminds/squirrel"
)

type UserRepository interface {
	// create
	Insert(ctx context.Context, user *model.User) error

	// read
	Fetch(ctx context.Context, options ...model.UserQueryOption) ([]model.User, error)
	Get(ctx context.Context, id string) (*model.User, error)
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	IsExistByEmail(ctx context.Context, email string) (bool, error)

	// update
	Update(ctx context.Context, user *model.User) error

	// delete
	Delete(ctx context.Context, user *model.User) error
}

type userRepository struct {
	db          infrastructure.DBTX
	loggerStack infrastructure.LoggerStack
}

func NewUserRepository(db infrastructure.DBTX, loggerStack infrastructure.LoggerStack) UserRepository {
	return &userRepository{
		db:          db,
		loggerStack: loggerStack,
	}
}

func (r *userRepository) tableName() string {
	return model.UserTableName
}

func (r *userRepository) tableAlias() string {
	return "u"
}

func (r *userRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.tableAlias())
}

func (r *userRepository) f(field string) string {
	return fmt.Sprintf("%s.%s", r.tableAlias(), field)
}

func (r *userRepository) fetchInternal(ctx context.Context, stmt squirrel.SelectBuilder) ([]model.User, error) {
	users := []model.User{}
	if err := fetch(r.db, ctx, &users, stmt); err != nil {
		return nil, err
	}
	return users, nil
}

func (r *userRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.User, error) {
	user := model.User{}
	if err := get(r.db, ctx, &user, stmt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) Insert(ctx context.Context, user *model.User) error {
	if user.CreatedAt.IsZero() {
		user.CreatedAt = time.Now().UTC()
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = time.Now().UTC()
	}
	return insert(r.db, ctx, r.tableName(), user.ToMap())
}

func (r *userRepository) Fetch(ctx context.Context, options ...model.UserQueryOption) ([]model.User, error) {
	option := model.UserQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	stmt := stmtBuilder.Select(r.f("*")).From(r.fromTable())

	if len(option.IdIn) > 0 {
		stmt = stmt.Where(squirrel.Eq{r.f("id"): option.IdIn})
	}

	if option.Phrase != nil {
		phrase := "%" + *option.Phrase + "%"
		stmt = stmt.Where(squirrel.Or{
			squirrel.ILike{r.f("name"): phrase},
			squirrel.ILike{r.f("email"): phrase},
		})
	}

	if !option.IsCount && option.Limit != nil {
		stmt = stmt.Limit(uint64(*option.Limit))
		if option.Page != nil && *option.Page > 1 {
			stmt = stmt.Offset(uint64((*option.Page - 1) * *option.Limit))
		}
	}

	return r.fetchInternal(ctx, stmt)
}

func (r *userRepository) Get(ctx context.Context, id string) (*model.User, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id})
	return r.getInternal(ctx, stmt)
}

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.ILike{r.f("email"): email}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *userRepository) IsExistByEmail(ctx context.Context, email string) (bool, error) {
	user, err := r.GetByEmail(ctx, email)
	if err != nil && err != constant.ErrNoData {
		return false, err
	}
	return user != nil, nil
}

func (r *userRepository) Update(ctx context.Context, user *model.User) error {
	user.UpdatedAt = time.Now().UTC()
	arg := user.ToMap()
	delete(arg, "id")
	delete(arg, "created_at")
	return update(r.db, ctx, r.tableName(), arg, squirrel.Eq{"id": user.Id})
}

func (r *userRepository) Delete(ctx context.Context, user *model.User) error {
	return destroy(r.db, ctx, r.tableName(), squirrel.Eq{"id": user.Id})
}
