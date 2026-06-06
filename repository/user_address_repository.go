package repository

import (
	"context"
	"fmt"

	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

// UserAddressRepository defines persistence operations for user_addresses.
type UserAddressRepository interface {
	// create
	Insert(ctx context.Context, address *model.UserAddress) error

	// read
	GetById(ctx context.Context, id string) (*model.UserAddress, error)
	GetDefaultByUserId(ctx context.Context, userId string) (*model.UserAddress, error)
	Fetch(ctx context.Context, options ...model.UserAddressQueryOption) ([]model.UserAddress, error)
	Count(ctx context.Context, options ...model.UserAddressQueryOption) (int, error)

	// update
	Update(ctx context.Context, address *model.UserAddress) (*model.UserAddress, error)
	UnsetDefaultByUserId(ctx context.Context, userId string) error

	// delete
	Delete(ctx context.Context, id string) error
}

type userAddressRepository struct {
	db infrastructure.DBTX
}

func NewUserAddressRepository(db infrastructure.DBTX) UserAddressRepository {
	return &userAddressRepository{db: db}
}

func (r *userAddressRepository) tableName() string { return model.UserAddressTableName }
func (r *userAddressRepository) alias() string     { return "ua" }
func (r *userAddressRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *userAddressRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

func (r *userAddressRepository) buildBaseStmt(option model.UserAddressQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.Select().From(r.fromTable())

	if option.UserId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("user_id"): *option.UserId})
	}

	return stmt
}

func (r *userAddressRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.UserAddress, error) {
	a := model.UserAddress{}
	if err := get(r.db, ctx, &a, stmt); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *userAddressRepository) fetchInternal(ctx context.Context, stmt squirrel.SelectBuilder) ([]model.UserAddress, error) {
	addresses := []model.UserAddress{}
	if err := fetch(r.db, ctx, &addresses, stmt); err != nil {
		return nil, err
	}
	return addresses, nil
}

func (r *userAddressRepository) Insert(ctx context.Context, address *model.UserAddress) error {
	return defaultInsert(r.db, ctx, address)
}

func (r *userAddressRepository) GetById(ctx context.Context, id string) (*model.UserAddress, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *userAddressRepository) GetDefaultByUserId(ctx context.Context, userId string) (*model.UserAddress, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("user_id"): userId, r.f("is_default"): true}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *userAddressRepository) Fetch(ctx context.Context, options ...model.UserAddressQueryOption) ([]model.UserAddress, error) {
	option := model.UserAddressQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}
	stmt := r.buildBaseStmt(option).Column(r.f("*"))
	stmt = model.Prepare(stmt, &option)
	return r.fetchInternal(ctx, stmt)
}

func (r *userAddressRepository) Count(ctx context.Context, options ...model.UserAddressQueryOption) (int, error) {
	option := model.UserAddressQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}
	option.IsCount = true
	stmt := r.buildBaseStmt(option).Column("COUNT(*)")
	var count int
	if err := get(r.db, ctx, &count, stmt); err != nil {
		return 0, err
	}
	return count, nil
}

func (r *userAddressRepository) Update(ctx context.Context, address *model.UserAddress) (*model.UserAddress, error) {
	now := util.CurrentDateTime()
	address.UpdatedAt = now
	if err := update(r.db, ctx, r.tableName(), address.ToMap(), squirrel.Eq{"id": address.Id}); err != nil {
		return nil, err
	}
	return r.GetById(ctx, address.Id)
}

func (r *userAddressRepository) UnsetDefaultByUserId(ctx context.Context, userId string) error {
	now := util.CurrentDateTime()
	return update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"is_default": false,
			"updated_at": now,
		},
		squirrel.Eq{"user_id": userId, "is_default": true},
	)
}

func (r *userAddressRepository) Delete(ctx context.Context, id string) error {
	return destroy(r.db, ctx, r.tableName(), squirrel.Eq{"id": id})
}
