package repository

import (
	"context"
	"fmt"

	"auction-service/infrastructure"
	"auction-service/model"

	"github.com/Masterminds/squirrel"
)

// ProductStatusHistoryRepository defines persistence operations for product status histories.
type ProductStatusHistoryRepository interface {
	// create
	Insert(ctx context.Context, h *model.ProductStatusHistory) error

	// read
	FetchByProductId(ctx context.Context, productId string) ([]model.ProductStatusHistory, error)
	FetchByProductIds(ctx context.Context, productIds []string) ([]model.ProductStatusHistory, error)
}

type productStatusHistoryRepository struct {
	db infrastructure.DBTX
}

func NewProductStatusHistoryRepository(db infrastructure.DBTX) ProductStatusHistoryRepository {
	return &productStatusHistoryRepository{db: db}
}

// ------------------------------------------------------------------ helpers

func (r *productStatusHistoryRepository) tableName() string {
	return model.ProductStatusHistoryTableName
}
func (r *productStatusHistoryRepository) alias() string { return "psh" }
func (r *productStatusHistoryRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *productStatusHistoryRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

// ------------------------------------------------------------------ create

func (r *productStatusHistoryRepository) Insert(ctx context.Context, h *model.ProductStatusHistory) error {
	return defaultInsert(r.db, ctx, h)
}

// ------------------------------------------------------------------ read

func (r *productStatusHistoryRepository) FetchByProductId(ctx context.Context, productId string) ([]model.ProductStatusHistory, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("product_id"): productId}).
		OrderBy(r.f("created_at") + " DESC")

	histories := []model.ProductStatusHistory{}
	if err := fetch(r.db, ctx, &histories, stmt); err != nil {
		return nil, err
	}
	return histories, nil
}

func (r *productStatusHistoryRepository) FetchByProductIds(ctx context.Context, productIds []string) ([]model.ProductStatusHistory, error) {
	if len(productIds) == 0 {
		return []model.ProductStatusHistory{}, nil
	}
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("product_id"): productIds}).
		OrderBy(r.f("created_at") + " DESC")

	histories := []model.ProductStatusHistory{}
	if err := fetch(r.db, ctx, &histories, stmt); err != nil {
		return nil, err
	}
	return histories, nil
}
