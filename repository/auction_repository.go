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
	GetByIdForUpdate(ctx context.Context, id string) (*model.Auction, error)
	Fetch(ctx context.Context, options ...model.AuctionQueryOption) ([]model.Auction, error)
	Count(ctx context.Context, options ...model.AuctionQueryOption) (int64, error)
	FetchStartable(ctx context.Context) ([]model.Auction, error)
	FetchCloseable(ctx context.Context) ([]model.Auction, error)

	// update
	Update(ctx context.Context, id string, startingPrice float64, startTime, endTime data_type.DateTime, fee float64) (*model.Auction, error)
	UpdateStatus(ctx context.Context, id string, status string) (*model.Auction, error)
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

// FetchCloseable returns all ON_GOING auctions whose end_time has passed.
func (r *auctionRepository) FetchCloseable(ctx context.Context) ([]model.Auction, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("status"): "ON_GOING"}).
		Where(squirrel.Expr(r.f("end_time") + " <= UTC_TIMESTAMP()"))
	return r.fetchInternal(ctx, stmt)
}

// ------------------------------------------------------------------ create

func (r *auctionRepository) Insert(ctx context.Context, auction *model.Auction) error {
	return defaultInsert(r.db, ctx, auction)
}

// ------------------------------------------------------------------ read

// FetchStartable returns all SCHEDULED auctions whose start_time has passed.
func (r *auctionRepository) FetchStartable(ctx context.Context) ([]model.Auction, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("status"): "SCHEDULED"}).
		Where(squirrel.Expr(r.f("start_time") + " <= UTC_TIMESTAMP()"))
	return r.fetchInternal(ctx, stmt)
}

func (r *auctionRepository) GetById(ctx context.Context, id string) (*model.Auction, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

// GetByIdForUpdate acquires a row-level lock on the auction row (SELECT … FOR UPDATE).
// Must be called inside a transaction.
func (r *auctionRepository) GetByIdForUpdate(ctx context.Context, id string) (*model.Auction, error) {
	query := fmt.Sprintf(
		"SELECT %s.* FROM %s WHERE %s.id = ? LIMIT 1 FOR UPDATE",
		r.alias(), r.fromTable(), r.alias(),
	)
	a := model.Auction{}
	dt := dbtx(r.db, ctx)
	if err := dt.GetContext(ctx, &a, query, id); err != nil {
		return nil, translateSqlError(err)
	}
	return &a, nil
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

func (r *auctionRepository) UpdateStatus(ctx context.Context, id string, status string) (*model.Auction, error) {
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
