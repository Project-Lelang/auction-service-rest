package repository

import (
	"context"
	"fmt"

	"auction-service/constant"
	"auction-service/data_type"
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
	GetByEmail(ctx context.Context, email string) (*model.User, error)
	GetById(ctx context.Context, id int64) (*model.User, error)
	FindAdminByEmail(ctx context.Context, email string) (*model.User, error)
	FetchByIds(ctx context.Context, ids []int64) ([]model.User, error)
	Fetch(ctx context.Context, options ...model.UserQueryOption) ([]model.User, error)
	Count(ctx context.Context, options ...model.UserQueryOption) (int64, error)

	// update
	Update(ctx context.Context, id int64, fullname string, birth data_type.DateTime, gender *string) (*model.User, error)
	UpdateIdentityInfo(ctx context.Context, id int64, nik string, identityImagePath string, selfieIdentityImagePath string) error
	UpdateBankAccountNumber(ctx context.Context, id int64, bankAccountNumber string) error
	DepositBalance(ctx context.Context, id int64, amount float64) (*model.User, error)
	WithdrawBalance(ctx context.Context, id int64, amount float64) (*model.User, error)
	SoftDelete(ctx context.Context, id int64) error
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
		stmt = stmt.Where(squirrel.Eq{"ur.role": *option.Role})
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

func (r *userRepository) GetByEmail(ctx context.Context, email string) (*model.User, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("email"): email}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *userRepository) GetById(ctx context.Context, id int64) (*model.User, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *userRepository) FindAdminByEmail(ctx context.Context, email string) (*model.User, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Join("user_roles ur ON ur.user_id = u.id").
		Where(squirrel.Eq{r.f("email"): email}).
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

func (r *userRepository) FetchByIds(ctx context.Context, ids []int64) ([]model.User, error) {
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

func (r *userRepository) Update(ctx context.Context, id int64, fullname string, birth data_type.DateTime, gender *string) (*model.User, error) {
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"fullname":   fullname,
			"birth":      birth,
			"gender":     gender,
			"updated_at": util.CurrentDateTime(),
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *userRepository) SoftDelete(ctx context.Context, id int64) error {
	return update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"is_deleted": true,
			"updated_at": util.CurrentDateTime(),
		},
		squirrel.Eq{"id": id},
	)
}

func (r *userRepository) UpdateIdentityInfo(ctx context.Context, id int64, nik string, identityImagePath string, selfieIdentityImagePath string) error {
	return update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"nik":                        nik,
			"identity_image_path":        identityImagePath,
			"selfie_identity_image_path": selfieIdentityImagePath,
			"updated_at":                 util.CurrentDateTime(),
		},
		squirrel.Eq{"id": id},
	)
}

func (r *userRepository) UpdateBankAccountNumber(ctx context.Context, id int64, bankAccountNumber string) error {
	return update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"bank_account_number": bankAccountNumber,
			"updated_at":          util.CurrentDateTime(),
		},
		squirrel.Eq{"id": id},
	)
}

func (r *userRepository) DepositBalance(ctx context.Context, id int64, amount float64) (*model.User, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("deposit amount must be positive")
	}

	stmt := stmtBuilder.Update(r.tableName()).Set("balance", squirrel.Expr("balance + ?", amount)).SetMap(map[string]interface{}{
		"updated_at": util.CurrentDateTime(),
	}).Where(squirrel.Eq{"id": id})

	if _, err := exec(r.db, ctx, stmt); err != nil {
		return nil, err
	}

	return r.GetById(ctx, id)
}

func (r *userRepository) WithdrawBalance(ctx context.Context, id int64, amount float64) (*model.User, error) {
	if amount <= 0 {
		return nil, fmt.Errorf("withdraw amount must be positive")
	}

	stmt := stmtBuilder.Update(r.tableName()).Set("balance", squirrel.Expr("balance - ?", amount)).SetMap(map[string]interface{}{
		"updated_at": util.CurrentDateTime(),
	}).Where(squirrel.And{
		squirrel.Eq{"id": id},
		squirrel.Expr("balance >= ?", amount),
	})

	result := dbtx(r.db, ctx)
	query, args, err := stmt.ToSql()
	if err != nil {
		return nil, translateSqlError(err)
	}

	res, err := result.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, translateSqlError(err)
	}
	if rows, _ := res.RowsAffected(); rows == 0 {
		return nil, fmt.Errorf("insufficient balance")
	}

	return r.GetById(ctx, id)
}
