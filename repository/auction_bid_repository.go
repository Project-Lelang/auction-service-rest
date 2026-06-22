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
	GetById(ctx context.Context, id int64) (*model.AuctionBid, error)
	GetHighestByAuctionId(ctx context.Context, auctionId int64) (*model.AuctionBid, error)
	GetNextHighestByAuctionId(ctx context.Context, auctionId int64, excludeBidId int64) (*model.AuctionBid, error)
	// GetNextHighestByAuctionIdExcludingUsers returns the highest bid for an auction
	// whose user_id is NOT in excludeUserIds. Used by OwnSecondChance to skip all
	// bids (including lower bids) from users who have already been cancelled winners.
	GetNextHighestByAuctionIdExcludingUsers(ctx context.Context, auctionId int64, excludeUserIds []int64) (*model.AuctionBid, error)
	Fetch(ctx context.Context, options ...model.AuctionBidQueryOption) ([]model.AuctionBid, error)
	FetchByIds(ctx context.Context, ids []int64) ([]model.AuctionBid, error)
	FetchByAuctionIds(ctx context.Context, auctionIds []int64) ([]model.AuctionBid, error)
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

func (r *auctionBidRepository) GetById(ctx context.Context, id int64) (*model.AuctionBid, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

// GetHighestByAuctionId returns the single highest bid for an auction.
func (r *auctionBidRepository) GetHighestByAuctionId(ctx context.Context, auctionId int64) (*model.AuctionBid, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_id"): auctionId}).
		OrderBy(r.f("amount") + " DESC").
		Limit(1)
	return r.getInternal(ctx, stmt)
}

// GetNextHighestByAuctionId returns the highest bid that is NOT the excluded bid.
// Used to find the fallback winner when the current winner doesn't pay.
func (r *auctionBidRepository) GetNextHighestByAuctionId(ctx context.Context, auctionId int64, excludeBidId int64) (*model.AuctionBid, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_id"): auctionId}).
		Where(squirrel.NotEq{r.f("id"): excludeBidId}).
		OrderBy(r.f("amount") + " DESC").
		Limit(1)
	return r.getInternal(ctx, stmt)
}

// GetNextHighestByAuctionIdExcludingUsers returns the highest bid whose user_id
// is not in excludeUserIds. squirrel translates a slice value in NotEq to a
// NOT IN clause. Used by OwnSecondChance so that ALL bids from a deadbeat winner
// are skipped, not just the single bid that was selected as winning bid.
func (r *auctionBidRepository) GetNextHighestByAuctionIdExcludingUsers(ctx context.Context, auctionId int64, excludeUserIds []int64) (*model.AuctionBid, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_id"): auctionId}).
		Where(squirrel.NotEq{r.f("user_id"): excludeUserIds}).
		OrderBy(r.f("amount") + " DESC").
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

func (r *auctionBidRepository) FetchByIds(ctx context.Context, ids []int64) ([]model.AuctionBid, error) {
	if len(ids) == 0 {
		return []model.AuctionBid{}, nil
	}
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): ids})
	return r.fetchInternal(ctx, stmt)
}

func (r *auctionBidRepository) FetchByAuctionIds(ctx context.Context, auctionIds []int64) ([]model.AuctionBid, error) {
	if len(auctionIds) == 0 {
		return []model.AuctionBid{}, nil
	}
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_id"): auctionIds}).
		OrderBy(r.f("amount") + " DESC")
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
