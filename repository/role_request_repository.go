package repository

import (
	"context"
	"fmt"

	"auction-service/infrastructure"
	"auction-service/model"

	"github.com/Masterminds/squirrel"
)

// RoleRequestRepository defines persistence operations for role requests.
type RoleRequestRepository interface {
	// create
	Insert(ctx context.Context, roleRequest *model.RoleRequest) error

	// read
	GetById(ctx context.Context, id string) (*model.RoleRequest, error)
	GetPendingByUserIdAndRole(ctx context.Context, userId string, role string) (*model.RoleRequest, error)
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

// ------------------------------------------------------------------ create

func (r *roleRequestRepository) Insert(ctx context.Context, roleRequest *model.RoleRequest) error {
	return defaultInsert(r.db, ctx, roleRequest)
}

// ------------------------------------------------------------------ read

func (r *roleRequestRepository) GetById(ctx context.Context, id string) (*model.RoleRequest, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}

func (r *roleRequestRepository) GetPendingByUserIdAndRole(ctx context.Context, userId string, role string) (*model.RoleRequest, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("user_id"): userId}).
		Where(squirrel.Eq{r.f("role"): role}).
		Where(squirrel.Eq{r.f("status"): "REQUESTED"}).
		Limit(1)
	return r.getInternal(ctx, stmt)
}
