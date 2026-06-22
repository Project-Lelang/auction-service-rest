package repository

import (
	"context"
	"fmt"

	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

// RoleRequestRepository defines persistence operations for role requests.
type RoleRequestRepository interface {
	// Create
	Insert(ctx context.Context, rr *model.RoleRequest) error

	// Read
	GetById(ctx context.Context, id int64) (*model.RoleRequest, error)
	GetPendingByUserIdAndRole(ctx context.Context, userId int64, role string) (*model.RoleRequest, error)
	Fetch(ctx context.Context, options ...model.RoleRequestQueryOption) ([]model.RoleRequest, error)
	Count(ctx context.Context, options ...model.RoleRequestQueryOption) (int64, error)

	// Update
	UpdateStatus(ctx context.Context, id int64, status string, message *string) error
	UpdateStatusUser(ctx context.Context, rr *model.RoleRequest) error
}

type roleRequestRepository struct {
	db infrastructure.DBTX
}

func NewRoleRequestRepository(db infrastructure.DBTX) RoleRequestRepository {
	return &roleRequestRepository{db: db}
}

// ------------------------------------------------------------------ helpers

func (r *roleRequestRepository) tableName() string { return model.RoleRequestTableName }
func (r *roleRequestRepository) alias() string     { return "rr" }
func (r *roleRequestRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *roleRequestRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

func (r *roleRequestRepository) getInternal(ctx context.Context, stmt squirrel.SelectBuilder) (*model.RoleRequest, error) {
	rq := model.RoleRequest{}
	if err := get(r.db, ctx, &rq, stmt); err != nil {
		return nil, err
	}
	return &rq, nil
}

func (r *roleRequestRepository) buildFetchQuery(option model.RoleRequestQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.
		Select(
			r.f("*"),
			"u.id AS user_id_join",
			"u.fullname AS user_fullname",
			"u.phone AS user_phone",
			"u.nik AS user_nik",
			"u.bank_account_number AS user_bank_account_number",
			"u.bank_account_name AS user_bank_account_name",
			"u.bank_name AS user_bank_name",
			"u.identity_image_path AS user_identity_image_path",
			"u.selfie_identity_image_path AS user_selfie_identity_image_path",
		).
		From(r.fromTable()).
		LeftJoin("users u ON u.id = " + r.f("user_id"))

	if option.UserId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("user_id"): *option.UserId})
	}

	if option.Status != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("status"): *option.Status})
	}

	if option.Role != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("role"): *option.Role})
	}

	return stmt
}

// ------------------------------------------------------------------ create

func (r *roleRequestRepository) Insert(ctx context.Context, rr *model.RoleRequest) error {
	return defaultInsert(r.db, ctx, rr)
}

// ------------------------------------------------------------------ read

func (r *roleRequestRepository) GetById(ctx context.Context, id int64) (*model.RoleRequest, error) {
	stmt := stmtBuilder.
		Select("*").
		From(r.tableName()).
		Where(squirrel.Eq{"id": id}).
		Limit(1)

	return r.getInternal(ctx, stmt)
}

func (r *roleRequestRepository) GetPendingByUserIdAndRole(ctx context.Context, userId int64, role string) (*model.RoleRequest, error) {
	stmt := stmtBuilder.
		Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("user_id"): userId}).
		Where(squirrel.Eq{r.f("role"): role}).
		Where(squirrel.Eq{r.f("status"): "REQUESTED"}).
		Limit(1)

	return r.getInternal(ctx, stmt)
}

func (r *roleRequestRepository) Fetch(ctx context.Context, options ...model.RoleRequestQueryOption) ([]model.RoleRequest, error) {
	option := model.RoleRequestQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	stmt := r.buildFetchQuery(option)
	stmt = model.Prepare(stmt, &option)

	type roleRequestRow struct {
		model.RoleRequest
		UserIdJoin                  int64   `db:"user_id_join"`
		UserFullname                string  `db:"user_fullname"`
		UserPhone                   string  `db:"user_phone"`
		UserNik                     *string `db:"user_nik"`
		UserBankAccountNumber       *string `db:"user_bank_account_number"`
		UserBankAccountName         *string `db:"user_bank_account_name"`
		UserBankName                *string `db:"user_bank_name"`
		UserIdentityImagePath       *string `db:"user_identity_image_path"`
		UserSelfieIdentityImagePath *string `db:"user_selfie_identity_image_path"`
	}

	rows := []roleRequestRow{}
	if err := fetch(r.db, ctx, &rows, stmt); err != nil {
		return nil, err
	}

	results := []model.RoleRequest{}
	for _, row := range rows {
		item := row.RoleRequest
		item.User = &model.RoleRequestUser{
			Id:                      row.UserIdJoin,
			Fullname:                row.UserFullname,
			Phone:                   row.UserPhone,
			Nik:                     row.UserNik,
			BankAccountNumber:       row.UserBankAccountNumber,
			BankAccountName:         row.UserBankAccountName,
			BankName:                row.UserBankName,
			IdentityImagePath:       row.UserIdentityImagePath,
			SelfieIdentityImagePath: row.UserSelfieIdentityImagePath,
		}
		results = append(results, item)
	}

	return results, nil
}

func (r *roleRequestRepository) Count(ctx context.Context, options ...model.RoleRequestQueryOption) (int64, error) {
	option := model.RoleRequestQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	stmt := stmtBuilder.
		Select("COUNT(" + r.f("id") + ")").
		From(r.fromTable())

	if option.UserId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("user_id"): *option.UserId})
	}

	if option.Status != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("status"): *option.Status})
	}

	if option.Role != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("role"): *option.Role})
	}

	var total int64
	err := get(r.db, ctx, &total, stmt)
	return total, err
}

// ------------------------------------------------------------------ update

func (r *roleRequestRepository) UpdateStatus(ctx context.Context, id int64, status string, message *string) error {
	payload := map[string]interface{}{
		"status":     status,
		"updated_at": util.CurrentDateTime(),
	}

	if message != nil {
		payload["message"] = *message
	}

	return update(
		r.db,
		ctx,
		r.tableName(),
		payload,
		squirrel.Eq{"id": id},
	)
}

func (r *roleRequestRepository) UpdateStatusUser(ctx context.Context, rr *model.RoleRequest) error {
	payload := map[string]interface{}{
		"status":     rr.Status,
		"role":       rr.Role,
		"message":    rr.Message,
		"updated_at": util.CurrentDateTime(),
	}

	return update(
		r.db,
		ctx,
		r.tableName(),
		payload,
		squirrel.Eq{"id": rr.Id},
	)
}
