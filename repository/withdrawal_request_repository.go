package repository

import (
	"context"
	"fmt"

	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

// WithdrawalRequestRepository defines persistence operations for withdrawal requests.
type WithdrawalRequestRepository interface {
	// Create
	Insert(ctx context.Context, wr *model.WithdrawalRequest) error

	// Read
	GetById(ctx context.Context, id int64) (*model.WithdrawalRequest, error)
	Fetch(ctx context.Context, options ...model.WithdrawalRequestQueryOption) ([]model.WithdrawalRequest, error)
	Count(ctx context.Context, options ...model.WithdrawalRequestQueryOption) (int64, error)

	// Update
	Update(ctx context.Context, id int64, payload map[string]interface{}) error
}

type withdrawalRequestRepository struct {
	db infrastructure.DBTX
}

func NewWithdrawalRequestRepository(db infrastructure.DBTX) WithdrawalRequestRepository {
	return &withdrawalRequestRepository{
		db: db,
	}
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

func (r *withdrawalRequestRepository) buildFetchQuery(option model.WithdrawalRequestQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.
		Select(
			r.f("*"),
			"u.id AS user_id_join",
			"u.fullname AS user_fullname",
			"u.phone AS user_phone",
			"u.bank_name AS user_bank_name",
			"u.bank_account_name AS user_bank_account_name",
			"u.bank_account_number AS user_bank_account_number",
		).
		From(r.fromTable()).
		LeftJoin("users u ON u.id = " + r.f("user_id"))

	if option.UserId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("user_id"): *option.UserId})
	}

	if option.ValidatorUserId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("validator_user_id"): *option.ValidatorUserId})
	}

	if option.Status != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("status"): *option.Status})
	}

	return stmt
}

// ------------------------------------------------------------------ create

func (r *withdrawalRequestRepository) Insert(ctx context.Context, wr *model.WithdrawalRequest) error {
	return defaultInsert(r.db, ctx, wr)
}

// ------------------------------------------------------------------ read

func (r *withdrawalRequestRepository) GetById(ctx context.Context, id int64) (*model.WithdrawalRequest, error) {
	stmt := stmtBuilder.
		Select("*").
		From(r.tableName()).
		Where(squirrel.Eq{"id": id}).
		Limit(1)

	return r.getInternal(ctx, stmt)
}

func (r *withdrawalRequestRepository) Fetch(ctx context.Context, options ...model.WithdrawalRequestQueryOption) ([]model.WithdrawalRequest, error) {
	option := model.WithdrawalRequestQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	stmt := r.buildFetchQuery(option)
	stmt = model.Prepare(stmt, &option)

	type withdrawalRequestRow struct {
		model.WithdrawalRequest
		UserIdJoin            string  `db:"user_id_join"`
		UserFullname          string  `db:"user_fullname"`
		UserPhone             string  `db:"user_phone"`
		UserBankName          *string `db:"user_bank_name"`
		UserBankAccountName   *string `db:"user_bank_account_name"`
		UserBankAccountNumber *string `db:"user_bank_account_number"`
	}

	rows := []withdrawalRequestRow{}
	if err := fetch(r.db, ctx, &rows, stmt); err != nil {
		return nil, err
	}

	results := []model.WithdrawalRequest{}
	for _, row := range rows {
		item := row.WithdrawalRequest
		item.User = &model.WithdrawalRequestUser{
			Id:                row.UserIdJoin,
			Fullname:          row.UserFullname,
			Phone:             row.UserPhone,
			BankName:          row.UserBankName,
			BankAccountName:   row.UserBankAccountName,
			BankAccountNumber: row.UserBankAccountNumber,
		}
		results = append(results, item)
	}

	return results, nil
}

func (r *withdrawalRequestRepository) Count(ctx context.Context, options ...model.WithdrawalRequestQueryOption) (int64, error) {
	option := model.WithdrawalRequestQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	stmt := stmtBuilder.
		Select("COUNT(*)").
		From(r.fromTable())

	if option.UserId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("user_id"): *option.UserId})
	}

	if option.ValidatorUserId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("validator_user_id"): *option.ValidatorUserId})
	}

	if option.Status != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("status"): *option.Status})
	}

	var total int64
	if err := get(r.db, ctx, &total, stmt); err != nil {
		return 0, err
	}

	return total, nil
}

// ------------------------------------------------------------------ update

func (r *withdrawalRequestRepository) Update(ctx context.Context, id int64, payload map[string]interface{}) error {
	payload["updated_at"] = util.CurrentDateTime()

	return update(
		r.db,
		ctx,
		r.tableName(),
		payload,
		squirrel.Eq{"id": id},
	)
}
