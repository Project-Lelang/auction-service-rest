package repository

import (
	"context"
	"fmt"

	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

// AuctionWinnerRepository defines persistence operations for auction winners.
type AuctionWinnerRepository interface {
	// create
	Insert(ctx context.Context, winner *model.AuctionWinner) error

	// read
	GetById(ctx context.Context, id string) (*model.AuctionWinner, error)
	GetActiveByAuctionIdForUpdate(ctx context.Context, auctionId string) (*model.AuctionWinner, error)
	GetLatestByAuctionId(ctx context.Context, auctionId string) (*model.AuctionWinner, error)
	Fetch(ctx context.Context, options ...model.AuctionWinnerQueryOption) ([]model.AuctionWinner, error)
	FetchByAuctionIds(ctx context.Context, auctionIds []string) ([]model.AuctionWinner, error)
	Count(ctx context.Context, options ...model.AuctionWinnerQueryOption) (int64, error)

	// update
	UpdateBidId(ctx context.Context, id string, auctionBidId string) (*model.AuctionWinner, error)
	UpdateStatus(ctx context.Context, id string, status string) (*model.AuctionWinner, error)
}

type auctionWinnerRepository struct {
	db infrastructure.DBTX
}

func NewAuctionWinnerRepository(db infrastructure.DBTX) AuctionWinnerRepository {
	return &auctionWinnerRepository{db: db}
}

func (r *auctionWinnerRepository) tableName() string { return model.AuctionWinnerTableName }
func (r *auctionWinnerRepository) alias() string     { return "aw" }
func (r *auctionWinnerRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *auctionWinnerRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

func (r *auctionWinnerRepository) buildBaseStmt(option model.AuctionWinnerQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.Select().From(r.fromTable())

	if option.AuctionId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("auction_id"): *option.AuctionId})
	}
	if option.Status != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("status"): *option.Status})
	}

	return stmt
}

func (r *auctionWinnerRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.AuctionWinner, error) {
	w := model.AuctionWinner{}
	if err := get(r.db, ctx, &w, stmt); err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *auctionWinnerRepository) fetchInternal(ctx context.Context, stmt squirrel.SelectBuilder) ([]model.AuctionWinner, error) {
	winners := []model.AuctionWinner{}
	if err := fetch(r.db, ctx, &winners, stmt); err != nil {
		return nil, err
	}
	return winners, nil
}

func (r *auctionWinnerRepository) Insert(ctx context.Context, winner *model.AuctionWinner) error {
	return defaultInsert(r.db, ctx, winner)
}

func (r *auctionWinnerRepository) GetById(ctx context.Context, id string) (*model.AuctionWinner, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

// GetActiveByAuctionIdForUpdate returns the active (non-CANCELLED) winner for an
// auction with a row-level FOR UPDATE lock. "Active" covers ON_GOING,
// WAITING_FOR_PAYMENT and COMPLETED — i.e. any status that means the winner slot
// is occupied. Used both to prevent duplicate winner inserts and to guard
// payment/shipment state transitions.
func (r *auctionWinnerRepository) GetActiveByAuctionIdForUpdate(ctx context.Context, auctionId string) (*model.AuctionWinner, error) {
	query := fmt.Sprintf(
		"SELECT %s.* FROM %s WHERE %s.auction_id = ? AND %s.status != 'CANCELLED' LIMIT 1 FOR UPDATE",
		r.alias(), r.fromTable(), r.alias(), r.alias(),
	)
	w := model.AuctionWinner{}
	dt := dbtx(r.db, ctx)
	if err := dt.GetContext(ctx, &w, query, auctionId); err != nil {
		return nil, translateSqlError(err)
	}
	return &w, nil
}

func (r *auctionWinnerRepository) GetLatestByAuctionId(ctx context.Context, auctionId string) (*model.AuctionWinner, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_id"): auctionId}).
		OrderBy(r.f("created_at") + " DESC").
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *auctionWinnerRepository) Fetch(ctx context.Context, options ...model.AuctionWinnerQueryOption) ([]model.AuctionWinner, error) {
	option := model.AuctionWinnerQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}
	stmt := r.buildBaseStmt(option).Column(r.f("*"))
	stmt = model.Prepare(stmt, &option)
	return r.fetchInternal(ctx, stmt)
}

func (r *auctionWinnerRepository) FetchByAuctionIds(ctx context.Context, auctionIds []string) ([]model.AuctionWinner, error) {
	if len(auctionIds) == 0 {
		return []model.AuctionWinner{}, nil
	}
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_id"): auctionIds}).
		Where(squirrel.NotEq{r.f("status"): "CANCELLED"})
	return r.fetchInternal(ctx, stmt)
}

func (r *auctionWinnerRepository) Count(ctx context.Context, options ...model.AuctionWinnerQueryOption) (int64, error) {
	option := model.AuctionWinnerQueryOption{}
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

func (r *auctionWinnerRepository) UpdateBidId(ctx context.Context, id string, auctionBidId string) (*model.AuctionWinner, error) {
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"auction_bid_id": auctionBidId,
			"updated_at":     util.CurrentDateTime(),
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}

func (r *auctionWinnerRepository) UpdateStatus(ctx context.Context, id string, status string) (*model.AuctionWinner, error) {
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
