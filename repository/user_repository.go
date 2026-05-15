package repository

import (
	"context"
	"fmt"

	"auction-service/constant"
	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

// UserRepository defines all persistence operations for the User entity.
type UserRepository interface {
	// create
	Insert(ctx context.Context, user *model.User) error

	// read
	GetByPhone(ctx context.Context, phone string) (*model.User, error)
	GetById(ctx context.Context, id string) (*model.User, error)
	FindAdminByPhone(ctx context.Context, phone string) (*model.User, error)
	FetchByIds(ctx context.Context, ids []string) ([]model.User, error)
	Fetch(ctx context.Context, options ...model.UserQueryOption) ([]model.User, error)
	Count(ctx context.Context, options ...model.UserQueryOption) (int64, error)

	// update
	SoftDelete(ctx context.Context, id string) error
}

type userRepository struct {
	db infrastructure.DBTX
}

func NewUserRepository(db infrastructure.DBTX) UserRepository {
	return &userRepository{db: db}
}

// ------------------------------------------------------------------ helpers

func (r *userRepository) tableName() string { return model.UserTableName }
func (r *userRepository) alias() string     { return "u" }
func (r *userRepository) fromTable() string { return fmt.Sprintf("%s %s", r.tableName(), r.alias()) }
func (r *userRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

// buildBaseStmt returns a SelectBuilder with FROM, JOIN, and WHERE filters applied.
// Callers must add the desired columns and call model.Prepare() before executing.
func (r *userRepository) buildBaseStmt(option model.UserQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.Select().
		From(r.fromTable()).
		LeftJoin("user_roles ur ON ur.user_id = u.id")

	if option.Role != nil {
		stmt = stmt.Where(squirrel.ILike{"ur.role": *option.Role})
	}

	return stmt
}

func (r *userRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.User, error) {
	user := model.User{}
	if err := get(r.db, ctx, &user, stmt); err != nil {
		return nil, err
	}
	return &user, nil
}

func (r *userRepository) fetchInternal(ctx context.Context, stmt squirrel.SelectBuilder) ([]model.User, error) {
	users := []model.User{}
	if err := fetch(r.db, ctx, &users, stmt); err != nil {
		return nil, err
	}
	return users, nil
}

// ------------------------------------------------------------------ create

func (r *userRepository) Insert(ctx context.Context, user *model.User) error {
	return defaultInsert(r.db, ctx, user)
}

// ------------------------------------------------------------------ read

func (r *userRepository) GetByPhone(ctx context.Context, phone string) (*model.User, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("phone"): phone}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *userRepository) GetById(ctx context.Context, id string) (*model.User, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *userRepository) FindAdminByPhone(ctx context.Context, phone string) (*model.User, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Join("user_roles ur ON ur.user_id = u.id").
		Where(squirrel.Eq{r.f("phone"): phone}).
		Where(squirrel.Eq{r.f("is_deleted"): false}).
		Where(squirrel.Or{
			squirrel.Eq{"ur.role": constant.RoleAdmin},
			squirrel.Eq{"ur.role": constant.RoleSuperAdmin},
		}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *userRepository) Fetch(ctx context.Context, options ...model.UserQueryOption) ([]model.User, error) {
	option := model.UserQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	stmt := r.buildBaseStmt(option).Column("DISTINCT " + r.f("*"))
	stmt = model.Prepare(stmt, &option)

	return r.fetchInternal(ctx, stmt)
}

func (r *userRepository) FetchByIds(ctx context.Context, ids []string) ([]model.User, error) {
	if len(ids) == 0 {
		return []model.User{}, nil
	}
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): ids})
	return r.fetchInternal(ctx, stmt)
}

func (r *userRepository) Count(ctx context.Context, options ...model.UserQueryOption) (int64, error) {
	option := model.UserQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	stmt := r.buildBaseStmt(option).Column("COUNT(DISTINCT u.id)")

	var count int64
	if err := get(r.db, ctx, &count, stmt); err != nil {
		return 0, err
	}
	return count, nil
}

// ------------------------------------------------------------------ update

func (r *userRepository) SoftDelete(ctx context.Context, id string) error {
	return update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"is_deleted": true,
			"updated_at": util.CurrentDateTime(),
		},
		squirrel.Eq{"id": id},
	)
}
