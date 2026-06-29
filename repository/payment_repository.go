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

// PaymentRepository defines persistence operations for payments.
type PaymentRepository interface {
	// create
	Insert(ctx context.Context, payment *model.Payment) error

	// read
	GetById(ctx context.Context, id int64) (*model.Payment, error)
	GetByCode(ctx context.Context, code string) (*model.Payment, error)
	GetActiveByAuctionId(ctx context.Context, auctionId int64) (*model.Payment, error)
	GetCompletedByAuctionId(ctx context.Context, auctionId int64) (*model.Payment, error)
	Fetch(ctx context.Context, options ...model.PaymentQueryOption) ([]model.Payment, error)
	FetchByAuctionIds(ctx context.Context, auctionIds []int64) ([]model.Payment, error)
	Count(ctx context.Context, options ...model.PaymentQueryOption) (int64, error)
	// FetchExpiredWaiting returns payments that are still WAITING_FOR_PAYMENT but
	// whose expired_at has already passed. Used for startup recovery.
	FetchExpiredWaiting(ctx context.Context) ([]model.Payment, error)

	// update
	UpdateStatus(ctx context.Context, id int64, status string) (*model.Payment, error)
	UpdateSnapInfo(ctx context.Context, id int64, snapUrl string, snapToken string) (*model.Payment, error)
}

type paymentRepository struct {
	db infrastructure.DBTX
}

func NewPaymentRepository(db infrastructure.DBTX) PaymentRepository {
	return &paymentRepository{db: db}
}

func (r *paymentRepository) tableName() string { return model.PaymentTableName }
func (r *paymentRepository) alias() string     { return "p" }
func (r *paymentRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *paymentRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

func (r *paymentRepository) buildBaseStmt(option model.PaymentQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.Select().From(r.fromTable())

	if option.AuctionId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("auction_id"): *option.AuctionId})
	}
	if option.UserId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("user_id"): *option.UserId})
	}
	if option.Status != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("status"): *option.Status})
	}

	return stmt
}

func (r *paymentRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.Payment, error) {
	p := model.Payment{}
	if err := get(r.db, ctx, &p, stmt); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRepository) fetchInternal(ctx context.Context, stmt squirrel.SelectBuilder) ([]model.Payment, error) {
	payments := []model.Payment{}
	if err := fetch(r.db, ctx, &payments, stmt); err != nil {
		return nil, err
	}
	return payments, nil
}

func (r *paymentRepository) Insert(ctx context.Context, payment *model.Payment) error {
	return defaultInsert(r.db, ctx, payment)
}

func (r *paymentRepository) GetById(ctx context.Context, id int64) (*model.Payment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *paymentRepository) GetByCode(ctx context.Context, code string) (*model.Payment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("code"): code}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *paymentRepository) GetActiveByAuctionId(ctx context.Context, auctionId int64) (*model.Payment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_id"): auctionId}).
		Where(squirrel.Eq{r.f("status"): constant.PaymentStatusWaitingForPayment}).
		OrderBy(r.f("created_at") + " DESC").
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *paymentRepository) GetCompletedByAuctionId(ctx context.Context, auctionId int64) (*model.Payment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_id"): auctionId}).
		Where(squirrel.Eq{r.f("status"): constant.PaymentStatusCompleted}).
		OrderBy(r.f("created_at") + " DESC").
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *paymentRepository) Fetch(ctx context.Context, options ...model.PaymentQueryOption) ([]model.Payment, error) {
	option := model.PaymentQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}
	stmt := r.buildBaseStmt(option).Column(r.f("*"))
	stmt = model.Prepare(stmt, &option)
	return r.fetchInternal(ctx, stmt)
}

func (r *paymentRepository) FetchByAuctionIds(ctx context.Context, auctionIds []int64) ([]model.Payment, error) {
	if len(auctionIds) == 0 {
		return []model.Payment{}, nil
	}
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("auction_id"): auctionIds})
	return r.fetchInternal(ctx, stmt)
}

func (r *paymentRepository) Count(ctx context.Context, options ...model.PaymentQueryOption) (int64, error) {
	option := model.PaymentQueryOption{}
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

func (r *paymentRepository) FetchExpiredWaiting(ctx context.Context) ([]model.Payment, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("status"): constant.PaymentStatusWaitingForPayment}).
		Where(squirrel.Lt{r.f("expired_at"): util.CurrentDateTime()})
	return r.fetchInternal(ctx, stmt)
}

func (r *paymentRepository) UpdateStatus(ctx context.Context, id int64, status string) (*model.Payment, error) {
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

func (r *paymentRepository) UpdateSnapInfo(ctx context.Context, id int64, snapUrl string, snapToken string) (*model.Payment, error) {
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"snap_url":   snapUrl,
			"snap_token": snapToken,
			"updated_at": util.CurrentDateTime(),
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}
