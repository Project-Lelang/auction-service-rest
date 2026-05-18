package repository

import (
	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"
	"context"

	"github.com/Masterminds/squirrel"
)

type PaymentMethodRepository interface {
	Insert(ctx context.Context, pm *model.PaymentMethod) error
	GetById(ctx context.Context, id int64) (*model.PaymentMethod, error)
	Fetch(ctx context.Context, options ...model.PaymentMethodQueryOption) ([]model.PaymentMethod, error)
	Update(ctx context.Context, id int64, payload map[string]interface{}) error
	Count(ctx context.Context, options ...model.PaymentMethodQueryOption) (int64, error)
}

type paymentMethodRepository struct {
	db infrastructure.DBTX
}

func NewPaymentMethodRepository(db infrastructure.DBTX) PaymentMethodRepository {
	return &paymentMethodRepository{db: db}
}

func (r *paymentMethodRepository) tableName() string {
	return model.PaymentMethodTableName
}

func (r *paymentMethodRepository) Insert(ctx context.Context, pm *model.PaymentMethod) error {
	return defaultInsert(r.db, ctx, pm)
}

func (r *paymentMethodRepository) GetById(ctx context.Context, id int64) (*model.PaymentMethod, error) {
	stmt := stmtBuilder.Select("*").From(r.tableName()).Where(squirrel.Eq{"id": id}).Limit(1)
	res := model.PaymentMethod{}
	if err := get(r.db, ctx, &res, stmt); err != nil {
		return nil, err
	}
	return &res, nil
}

func (r *paymentMethodRepository) Count(ctx context.Context, options ...model.PaymentMethodQueryOption) (int64, error) {
	option := model.PaymentMethodQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	// Menggunakan SELECT COUNT(*) dari tabel payment method
	stmt := stmtBuilder.
		Select("COUNT(*)").
		From(r.tableName())

	// Filter harus disamakan dengan fungsi Fetch agar totalnya akurat
	if option.Name != nil {
		stmt = stmt.Where(squirrel.Like{"name": "%" + *option.Name + "%"})
	}
	if option.Code != nil {
		stmt = stmt.Where(squirrel.Eq{"code": *option.Code})
	}
	if option.Type != nil {
		stmt = stmt.Where(squirrel.Eq{"type": *option.Type})
	}
	if option.IsActive != nil {
		stmt = stmt.Where(squirrel.Eq{"is_active": *option.IsActive})
	}

	var total int64
	if err := get(r.db, ctx, &total, stmt); err != nil {
		return 0, err
	}

	return total, nil
}

func (r *paymentMethodRepository) Fetch(ctx context.Context, options ...model.PaymentMethodQueryOption) ([]model.PaymentMethod, error) {
	option := model.PaymentMethodQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	stmt := stmtBuilder.Select("*").From(r.tableName())

	if option.Name != nil {
		stmt = stmt.Where(squirrel.Like{"name": "%" + *option.Name + "%"})
	}
	if option.Code != nil {
		stmt = stmt.Where(squirrel.Eq{"code": *option.Code})
	}
	if option.Type != nil {
		stmt = stmt.Where(squirrel.Eq{
			"type": *option.Type,
		})
	}
	if option.IsActive != nil {
		stmt = stmt.Where(squirrel.Eq{"is_active": *option.IsActive})
	}

	stmt = model.Prepare(stmt, &option)

	res := []model.PaymentMethod{}
	err := fetch(r.db, ctx, &res, stmt)
	return res, err
}

func (r *paymentMethodRepository) Update(ctx context.Context, id int64, payload map[string]interface{}) error {
	payload["updated_at"] = util.CurrentDateTime()
	return update(r.db, ctx, r.tableName(), payload, squirrel.Eq{"id": id})
}
