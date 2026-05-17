package repository

import (
	"context"
	"fmt"

	"auction-service/infrastructure"
	"auction-service/model"

	"github.com/Masterminds/squirrel"
)

// AuctionBidRepository defines persistence operations for auction bids.
type AuctionBidRepository interface {
	// create
	Insert(ctx context.Context, bid *model.AuctionBid) error

	// read
	GetById(ctx context.Context, id string) (*model.AuctionBid, error)
	Fetch(ctx context.Context, options ...model.AuctionBidQueryOption) ([]model.AuctionBid, error)
	Count(ctx context.Context, options ...model.AuctionBidQueryOption) (int64, error)
}

type auctionBidRepository struct {
	db infrastructure.DBTX
}

func NewAuctionBidRepository(db infrastructure.DBTX) AuctionBidRepository {
	return &auctionBidRepository{db: db}
}

// ------------------------------------------------------------------ helpers

func (r *auctionBidRepository) tableName() string { return model.AuctionBidTableName }
func (r *auctionBidRepository) alias() string     { return "ab" }
func (r *auctionBidRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *auctionBidRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

func (r *auctionBidRepository) buildBaseStmt(option model.AuctionBidQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.Select().From(r.fromTable())

	if option.UserId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("user_id"): *option.UserId})
	}
	if option.AuctionId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("auction_id"): *option.AuctionId})
	}

	return stmt
}

func (r *auctionBidRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.AuctionBid, error) {
	b := model.AuctionBid{}
	if err := get(r.db, ctx, &b, stmt); err != nil {
		return nil, err
	}
	return &b, nil
}

func (r *auctionBidRepository) fetchInternal(ctx context.Context, stmt squirrel.SelectBuilder) ([]model.AuctionBid, error) {
	bids := []model.AuctionBid{}
	if err := fetch(r.db, ctx, &bids, stmt); err != nil {
		return nil, err
	}
	return bids, nil
}

// ------------------------------------------------------------------ create

func (r *auctionBidRepository) Insert(ctx context.Context, bid *model.AuctionBid) error {
	return defaultInsert(r.db, ctx, bid)
}

// ------------------------------------------------------------------ read

func (r *auctionBidRepository) GetById(ctx context.Context, id string) (*model.AuctionBid, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *auctionBidRepository) Fetch(ctx context.Context, options ...model.AuctionBidQueryOption) ([]model.AuctionBid, error) {
	option := model.AuctionBidQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	stmt := r.buildBaseStmt(option).Column(r.f("*"))
	stmt = model.Prepare(stmt, &option)

	return r.fetchInternal(ctx, stmt)
}

func (r *auctionBidRepository) Count(ctx context.Context, options ...model.AuctionBidQueryOption) (int64, error) {
	option := model.AuctionBidQueryOption{}
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
