package repository

import (
	"context"
	"fmt"

	"auction-service/constant"
	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

type SecondChanceOfferRepository interface {
	Insert(ctx context.Context, offer *model.SecondChanceOffer) error
	GetById(ctx context.Context, id int64) (*model.SecondChanceOffer, error)
	GetByIdForUpdate(ctx context.Context, id int64) (*model.SecondChanceOffer, error)
	GetPendingByAuctionId(ctx context.Context, auctionId int64) (*model.SecondChanceOffer, error)
	Fetch(ctx context.Context, options ...model.SecondChanceOfferQueryOption) ([]model.SecondChanceOffer, error)
	Count(ctx context.Context, options ...model.SecondChanceOfferQueryOption) (int64, error)
	FetchFinalByAuctionId(ctx context.Context, auctionId int64) ([]model.SecondChanceOffer, error)
	FetchExpiredPending(ctx context.Context) ([]model.SecondChanceOffer, error)
	UpdateStatus(ctx context.Context, id int64, status string) (*model.SecondChanceOffer, error)
}

type secondChanceOfferRepository struct {
	db infrastructure.DBTX
}

func NewSecondChanceOfferRepository(db infrastructure.DBTX) SecondChanceOfferRepository {
	return &secondChanceOfferRepository{db: db}
}

func (r *secondChanceOfferRepository) tableName() string { return model.SecondChanceOfferTableName }
func (r *secondChanceOfferRepository) alias() string     { return "sco" }
func (r *secondChanceOfferRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *secondChanceOfferRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

func (r *secondChanceOfferRepository) buildBaseStmt(option model.SecondChanceOfferQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.Select().From(r.fromTable())
	if option.AuctionId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("auction_id"): *option.AuctionId})
	}
	if option.SellerId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("seller_id"): *option.SellerId})
	}
	if option.BuyerId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("buyer_id"): *option.BuyerId})
	}
	if option.Status != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("status"): *option.Status})
	}
	return stmt
}

func (r *secondChanceOfferRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.SecondChanceOffer, error) {
	offer := model.SecondChanceOffer{}
	if err := get(r.db, ctx, &offer, stmt); err != nil {
		return nil, err
	}
	return &offer, nil
}

func (r *secondChanceOfferRepository) fetchInternal(ctx context.Context, stmt squirrel.SelectBuilder) ([]model.SecondChanceOffer, error) {
	offers := []model.SecondChanceOffer{}
	if err := fetch(r.db, ctx, &offers, stmt); err != nil {
		return nil, err
	}
	return offers, nil
}

func (r *secondChanceOfferRepository) Insert(ctx context.Context, offer *model.SecondChanceOffer) error {
	return defaultInsert(r.db, ctx, offer)
}

func (r *secondChanceOfferRepository) GetById(ctx context.Context, id int64) (*model.SecondChanceOffer, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *secondChanceOfferRepository) GetByIdForUpdate(ctx context.Context, id int64) (*model.SecondChanceOffer, error) {
	query := fmt.Sprintf(
		"SELECT %s.* FROM %s WHERE %s.id = ? LIMIT 1 FOR UPDATE",
		r.alias(), r.fromTable(), r.alias(),
	)
	offer := model.SecondChanceOffer{}
	dt := dbtx(r.db, ctx)
	if err := dt.GetContext(ctx, &offer, query, id); err != nil {
		return nil, translateSqlError(err)
	}
	return &offer, nil
}

func (r *secondChanceOfferRepository) GetPendingByAuctionId(ctx context.Context, auctionId int64) (*model.SecondChanceOffer, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_id"): auctionId}).
		Where(squirrel.Eq{r.f("status"): constant.SecondChanceOfferStatusPending}).
		Where(squirrel.Or{
			squirrel.Eq{r.f("expired_at"): nil},
			squirrel.Gt{r.f("expired_at"): util.CurrentDateTime()},
		}).
		OrderBy(r.f("created_at") + " DESC").
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *secondChanceOfferRepository) Fetch(ctx context.Context, options ...model.SecondChanceOfferQueryOption) ([]model.SecondChanceOffer, error) {
	option := model.SecondChanceOfferQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}
	stmt := r.buildBaseStmt(option).Column(r.f("*"))
	stmt = model.Prepare(stmt, &option)
	return r.fetchInternal(ctx, stmt)
}

func (r *secondChanceOfferRepository) Count(ctx context.Context, options ...model.SecondChanceOfferQueryOption) (int64, error) {
	option := model.SecondChanceOfferQueryOption{}
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

func (r *secondChanceOfferRepository) FetchFinalByAuctionId(ctx context.Context, auctionId int64) ([]model.SecondChanceOffer, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_id"): auctionId}).
		Where(squirrel.Eq{r.f("status"): []string{
			constant.SecondChanceOfferStatusAccepted,
			constant.SecondChanceOfferStatusRejected,
			constant.SecondChanceOfferStatusExpired,
		}})
	return r.fetchInternal(ctx, stmt)
}

func (r *secondChanceOfferRepository) FetchExpiredPending(ctx context.Context) ([]model.SecondChanceOffer, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("status"): constant.SecondChanceOfferStatusPending}).
		Where(squirrel.LtOrEq{r.f("expired_at"): util.CurrentDateTime()})
	return r.fetchInternal(ctx, stmt)
}

func (r *secondChanceOfferRepository) UpdateStatus(ctx context.Context, id int64, status string) (*model.SecondChanceOffer, error) {
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
