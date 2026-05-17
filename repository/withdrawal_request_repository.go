package repository

import (
	"context"
	"fmt"

	"auction-service/infrastructure"
	"auction-service/model"

	"github.com/Masterminds/squirrel"
)

// WithdrawalRequestRepository defines persistence operations for withdrawal requests.
type WithdrawalRequestRepository interface {
	// create
	Insert(ctx context.Context, withdrawalRequest *model.WithdrawalRequest) error

	// read
	GetById(ctx context.Context, id string) (*model.WithdrawalRequest, error)
}

type withdrawalRequestRepository struct {
	db infrastructure.DBTX
}

func NewWithdrawalRequestRepository(db infrastructure.DBTX) WithdrawalRequestRepository {
	return &withdrawalRequestRepository{db: db}
}

// ------------------------------------------------------------------ helpers

func (r *withdrawalRequestRepository) tableName() string { return model.WithdrawalRequestTableName }
func (r *withdrawalRequestRepository) alias() string     { return "wr" }
func (r *withdrawalRequestRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *withdrawalRequestRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

func (r *withdrawalRequestRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.WithdrawalRequest, error) {
	wr := model.WithdrawalRequest{}
	if err := get(r.db, ctx, &wr, stmt); err != nil {
		return nil, err
	}
	return &wr, nil
}

// ------------------------------------------------------------------ create

func (r *withdrawalRequestRepository) Insert(ctx context.Context, withdrawalRequest *model.WithdrawalRequest) error {
	return defaultInsert(r.db, ctx, withdrawalRequest)
}

// ------------------------------------------------------------------ read

func (r *withdrawalRequestRepository) GetById(ctx context.Context, id string) (*model.WithdrawalRequest, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}
