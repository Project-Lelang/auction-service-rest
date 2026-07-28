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

type UserStrikeRepository interface {
	Insert(ctx context.Context, strike *model.UserStrike) error
	HasActiveByBidderId(ctx context.Context, bidderId int64) (bool, error)
	Fetch(ctx context.Context, options ...model.UserStrikeQueryOption) ([]model.UserStrike, error)
	Count(ctx context.Context, options ...model.UserStrikeQueryOption) (int64, error)
}

type userStrikeRepository struct {
	db infrastructure.DBTX
}

func NewUserStrikeRepository(db infrastructure.DBTX) UserStrikeRepository {
	return &userStrikeRepository{db: db}
}

func (r *userStrikeRepository) tableName() string { return model.UserStrikeTableName }
func (r *userStrikeRepository) alias() string     { return "us" }
func (r *userStrikeRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *userStrikeRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

func (r *userStrikeRepository) buildBaseStmt(option model.UserStrikeQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.Select().From(r.fromTable())
	if option.BidderId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("bidder_id"): *option.BidderId})
	}
	if option.Status != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("status"): *option.Status})
	}
	return stmt
}

func (r *userStrikeRepository) Insert(ctx context.Context, strike *model.UserStrike) error {
	return defaultInsert(r.db, ctx, strike)
}

func (r *userStrikeRepository) HasActiveByBidderId(ctx context.Context, bidderId int64) (bool, error) {
	stmt := stmtBuilder.Select("COUNT(1)").
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("bidder_id"): bidderId}).
		Where(squirrel.Eq{r.f("status"): constant.UserStrikeStatusActive}).
		Where(squirrel.Or{
			squirrel.Eq{r.f("expired_at"): nil},
			squirrel.Gt{r.f("expired_at"): util.CurrentDateTime()},
		})

	var count int64
	if err := get(r.db, ctx, &count, stmt); err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *userStrikeRepository) Fetch(ctx context.Context, options ...model.UserStrikeQueryOption) ([]model.UserStrike, error) {
	option := model.UserStrikeQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}
	stmt := r.buildBaseStmt(option).Column(r.f("*"))
	stmt = model.Prepare(stmt, &option)

	strikes := []model.UserStrike{}
	if err := fetch(r.db, ctx, &strikes, stmt); err != nil {
		return nil, err
	}
	return strikes, nil
}

func (r *userStrikeRepository) Count(ctx context.Context, options ...model.UserStrikeQueryOption) (int64, error) {
	option := model.UserStrikeQueryOption{}
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
