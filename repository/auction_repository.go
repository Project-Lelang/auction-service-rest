package repository

import (
	"context"
	"fmt"

	"auction-service/data_type"
	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

// AuctionRepository defines persistence operations for auctions.
type AuctionRepository interface {
	// create
	Insert(ctx context.Context, auction *model.Auction) error

	// read
	GetById(ctx context.Context, id string) (*model.Auction, error)
	Fetch(ctx context.Context, options ...model.AuctionQueryOption) ([]model.Auction, error)
	Count(ctx context.Context, options ...model.AuctionQueryOption) (int64, error)

	// update
	Update(ctx context.Context, id string, startingPrice float64, startTime, endTime data_type.DateTime, fee float64) (*model.Auction, error)
}

type auctionRepository struct {
	db infrastructure.DBTX
}

func NewAuctionRepository(db infrastructure.DBTX) AuctionRepository {
	return &auctionRepository{db: db}
}

// ------------------------------------------------------------------ helpers

func (r *auctionRepository) tableName() string { return model.AuctionTableName }
func (r *auctionRepository) alias() string     { return "a" }
func (r *auctionRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *auctionRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

func (r *auctionRepository) buildBaseStmt(option model.AuctionQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.Select().From(r.fromTable())

	if option.UserId != nil {
		stmt = stmt.Join("products p ON p.id = a.product_id").
			Where(squirrel.Eq{"p.user_id": *option.UserId})
	}

	if option.Status != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("status"): *option.Status})
	}

	return stmt
}

func (r *auctionRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.Auction, error) {
	a := model.Auction{}
	if err := get(r.db, ctx, &a, stmt); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *auctionRepository) fetchInternal(ctx context.Context, stmt squirrel.SelectBuilder) ([]model.Auction, error) {
	auctions := []model.Auction{}
	if err := fetch(r.db, ctx, &auctions, stmt); err != nil {
		return nil, err
	}
	return auctions, nil
}

// ------------------------------------------------------------------ create

func (r *auctionRepository) Insert(ctx context.Context, auction *model.Auction) error {
	return defaultInsert(r.db, ctx, auction)
}

// ------------------------------------------------------------------ read

func (r *auctionRepository) GetById(ctx context.Context, id string) (*model.Auction, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *auctionRepository) Fetch(ctx context.Context, options ...model.AuctionQueryOption) ([]model.Auction, error) {
	option := model.AuctionQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	stmt := r.buildBaseStmt(option).Column(r.f("*"))
	stmt = model.Prepare(stmt, &option)

	return r.fetchInternal(ctx, stmt)
}

func (r *auctionRepository) Count(ctx context.Context, options ...model.AuctionQueryOption) (int64, error) {
	option := model.AuctionQueryOption{}
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

func (r *auctionRepository) Update(ctx context.Context, id string, startingPrice float64, startTime, endTime data_type.DateTime, fee float64) (*model.Auction, error) {
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"starting_price": startingPrice,
			"start_time":     startTime,
			"end_time":       endTime,
			"fee":            fee,
			"updated_at":     util.CurrentDateTime(),
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}
