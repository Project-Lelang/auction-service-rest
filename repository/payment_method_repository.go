package repository

import (
	"context"
	"fmt"

	"auction-service/infrastructure"
	"auction-service/model"

	"github.com/Masterminds/squirrel"
)

// PaymentMethodRepository defines persistence operations for payment methods.
type PaymentMethodRepository interface {
	// create
	Insert(ctx context.Context, pm *model.PaymentMethod) error

	// read
	GetById(ctx context.Context, id string) (*model.PaymentMethod, error)
	GetByCode(ctx context.Context, code string) (*model.PaymentMethod, error)
	Fetch(ctx context.Context, options ...model.PaymentMethodQueryOption) ([]model.PaymentMethod, error)
}

type paymentMethodRepository struct {
	db infrastructure.DBTX
}

func NewPaymentMethodRepository(db infrastructure.DBTX) PaymentMethodRepository {
	return &paymentMethodRepository{db: db}
}

func (r *paymentMethodRepository) tableName() string { return model.PaymentMethodTableName }
func (r *paymentMethodRepository) alias() string     { return "pm" }
func (r *paymentMethodRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *paymentMethodRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

func (r *paymentMethodRepository) buildBaseStmt(option model.PaymentMethodQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.Select().From(r.fromTable())

	if option.IsActive != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("is_active"): *option.IsActive})
	}
	if option.Type != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("type"): *option.Type})
	}

	return stmt
}

func (r *paymentMethodRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.PaymentMethod, error) {
	pm := model.PaymentMethod{}
	if err := get(r.db, ctx, &pm, stmt); err != nil {
		return nil, err
	}
	return &pm, nil
}

func (r *paymentMethodRepository) fetchInternal(ctx context.Context, stmt squirrel.SelectBuilder) ([]model.PaymentMethod, error) {
	methods := []model.PaymentMethod{}
	if err := fetch(r.db, ctx, &methods, stmt); err != nil {
		return nil, err
	}
	return methods, nil
}

func (r *paymentMethodRepository) GetById(ctx context.Context, id string) (*model.PaymentMethod, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *paymentMethodRepository) Fetch(ctx context.Context, options ...model.PaymentMethodQueryOption) ([]model.PaymentMethod, error) {
	option := model.PaymentMethodQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}
	stmt := r.buildBaseStmt(option).Column(r.f("*"))
	stmt = model.Prepare(stmt, &option)
	return r.fetchInternal(ctx, stmt)
}

// ------------------------------------------------------------------ create

func (r *paymentMethodRepository) Insert(ctx context.Context, pm *model.PaymentMethod) error {
	return defaultInsert(r.db, ctx, pm)
}

// ------------------------------------------------------------------ extra reads

func (r *paymentMethodRepository) GetByCode(ctx context.Context, code string) (*model.PaymentMethod, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("code"): code}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}
