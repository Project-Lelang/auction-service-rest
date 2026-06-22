package repository

import (
	"context"
	"fmt"

	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

type NotificationRepository interface {
	Insert(ctx context.Context, notification *model.Notification) error
	GetById(ctx context.Context, id int64) (*model.Notification, error)
	Fetch(ctx context.Context, options ...model.NotificationQueryOption) ([]model.Notification, error)
	Count(ctx context.Context, options ...model.NotificationQueryOption) (int64, error)
	MarkRead(ctx context.Context, id int64) (*model.Notification, error)
}

type notificationRepository struct {
	db infrastructure.DBTX
}

func NewNotificationRepository(db infrastructure.DBTX) NotificationRepository {
	return &notificationRepository{db: db}
}

func (r *notificationRepository) tableName() string { return model.NotificationTableName }
func (r *notificationRepository) alias() string     { return "n" }
func (r *notificationRepository) fromTable() string {
	return fmt.Sprintf("%s %s", r.tableName(), r.alias())
}
func (r *notificationRepository) f(col string) string {
	return fmt.Sprintf("%s.%s", r.alias(), col)
}

func (r *notificationRepository) buildBaseStmt(option model.NotificationQueryOption) squirrel.SelectBuilder {
	stmt := stmtBuilder.Select().From(r.fromTable())

	if option.UserId != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("user_id"): *option.UserId})
	}
	if option.IsRead != nil {
		stmt = stmt.Where(squirrel.Eq{r.f("is_read"): *option.IsRead})
	}

	return stmt
}

func (r *notificationRepository) Insert(ctx context.Context, notification *model.Notification) error {
	return defaultInsert(r.db, ctx, notification)
}

func (r *notificationRepository) GetById(ctx context.Context, id int64) (*model.Notification, error) {
	stmt := stmtBuilder.Select(r.f("*")).
		From(r.fromTable()).
		Where(squirrel.Eq{r.f("id"): id}).
		Limit(1)

	notification := model.Notification{}
	if err := get(r.db, ctx, &notification, stmt); err != nil {
		return nil, err
	}
	return &notification, nil
}

func (r *notificationRepository) Fetch(ctx context.Context, options ...model.NotificationQueryOption) ([]model.Notification, error) {
	option := model.NotificationQueryOption{}
	if len(options) > 0 {
		option = options[0]
	}

	stmt := r.buildBaseStmt(option).Column(r.f("*"))
	stmt = model.Prepare(stmt, &option)

	notifications := []model.Notification{}
	if err := fetch(r.db, ctx, &notifications, stmt); err != nil {
		return nil, err
	}
	return notifications, nil
}

func (r *notificationRepository) Count(ctx context.Context, options ...model.NotificationQueryOption) (int64, error) {
	option := model.NotificationQueryOption{}
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

func (r *notificationRepository) MarkRead(ctx context.Context, id int64) (*model.Notification, error) {
	if err := update(r.db, ctx, r.tableName(),
		map[string]interface{}{
			"is_read":    true,
			"updated_at": util.CurrentDateTime(),
		},
		squirrel.Eq{"id": id},
	); err != nil {
		return nil, err
	}
	return r.GetById(ctx, id)
}
