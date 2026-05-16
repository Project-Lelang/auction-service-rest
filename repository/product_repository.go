package repository

import (
	"context"
	"fmt"
	"strings"

	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

// ProductRepository defines all persistence operations for the Product entity.
type ProductRepository interface {
	// create
	Insert(ctx context.Context, product *model.Product) error

	// read
	GetById(ctx context.Context, id string) (*model.Product, error)
	Fetch(ctx context.Context, options ...model.ProductQueryOption) ([]model.Product, error)
	Count(ctx context.Context, options ...model.ProductQueryOption) (int64, error)

	// update
	Update(ctx context.Context, id string, name string, description *string, condition string) (*model.Product, error)
	UpdateStatus(ctx context.Context, id string, status string) (*model.Product, error)
}

type productRepository struct {
	db infrastructure.DBTX
}

func NewProductRepository(db infrastructure.DBTX) ProductRepository {
	return &productRepository{db: db}
}

// ------------------------------------------------------------------ helpers

func (r *productRepository) tableName() string { return model.ProductTableName }
func (r *productRepository) alias() string     { return "p" }
func (r *productRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *productRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

// buildBaseStmt returns a SelectBuilder with FROM and WHERE filters applied.
// Callers should add columns and call model.Prepare() before executing.
func (r *productRepository) buildBaseStmt(option model.ProductQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.Select().From(r.fromTable())

	if option.UserId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("user_id"): *option.UserId})
	}
	if option.Status != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("status"): *option.Status})
	}
	if option.Condition != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("condition"): *option.Condition})
	}
	if option.Search != nil && strings.TrimSpace(*option.Search) != "" {
		stmt = stmt.Where(squirrel.Like{r.f("name"): "%" + *option.Search + "%"})
	}

	return stmt
}

func (r *productRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.Product, error) {
	p := model.Product{}
	if err := get(r.db, ctx, &p, stmt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *productRepository) fetchInternal(ctx context.Context, stmt squirrel.SelectBuilder) ([]model.Product, error) {
	products := []model.Product{}
	if err := fetch(r.db, ctx, &products, stmt); err != nil {
		return nil, err
	}
	return products, nil
}

// ------------------------------------------------------------------ create

func (r *productRepository) Insert(ctx context.Context, product *model.Product) error {
	return defaultInsert(r.db, ctx, product)
}

// ------------------------------------------------------------------ read

func (r *productRepository) GetById(ctx context.Context, id string) (*model.Product, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *productRepository) Fetch(ctx context.Context, options ...model.ProductQueryOption) ([]model.Product, error) {
	option := model.ProductQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	stmt := r.buildBaseStmt(option).Column(r.f("*"))
	stmt = model.Prepare(stmt, &option)

	return r.fetchInternal(ctx, stmt)
}

func (r *productRepository) Count(ctx context.Context, options ...model.ProductQueryOption) (int64, error) {
	option := model.ProductQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}
	option.IsCount = true

	stmt := r.buildBaseStmt(option)
	stmt = model.Prepare(stmt, &option)

	var count int64
	if err := get(r.db, ctx, &count, stmt); err != nil {
		return 0, err
	}
	return count, nil
}

// ------------------------------------------------------------------ update

func (r *productRepository) Update(ctx context.Context, id string, name string, description *string, condition string) (*model.Product, error) {
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"name":        name,
			"description": description,
			"condition":   condition,
			"updated_at":  util.CurrentDateTime(),
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *productRepository) UpdateStatus(ctx context.Context, id string, status string) (*model.Product, error) {
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"status":     status,
			"updated_at": util.CurrentDateTime(),
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}
