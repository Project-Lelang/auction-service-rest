package repository

import (
	"context"

	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

type UserFcmTokenRepository interface {
	Insert(ctx context.Context, userFcmToken *model.UserFcmToken) error
	Upsert(ctx context.Context, userFcmToken *model.UserFcmToken) error
	Delete(ctx context.Context, userId int64, fcmToken string) error
	DeleteByToken(ctx context.Context, fcmToken string) error
	GetByUserIdAndFcmToken(ctx context.Context, userId int64, fcmToken string) (*model.UserFcmToken, error)
	FetchByUserIds(ctx context.Context, userIds []int64) ([]model.UserFcmToken, error)
}

type userFcmTokenRepository struct {
	db infrastructure.DBTX
}

func NewUserFcmTokenRepository(db infrastructure.DBTX) UserFcmTokenRepository {
	return &userFcmTokenRepository{db: db}
}

func (r *userFcmTokenRepository) Insert(ctx context.Context, userFcmToken *model.UserFcmToken) error {
	return defaultInsert(r.db, ctx, userFcmToken)
}

func (r *userFcmTokenRepository) Upsert(ctx context.Context, userFcmToken *model.UserFcmToken) error {
	now := userFcmToken.GetUpdatedAt()
	if now.IsZero() {
		now = userFcmToken.GetCreatedAt()
	}
	if now.IsZero() {
		now = util.CurrentDateTime()
	}
	if userFcmToken.GetCreatedAt().IsZero() {
		userFcmToken.SetCreatedAt(now)
	}
	userFcmToken.SetUpdatedAt(now)

	query := `
		INSERT INTO user_fcm_tokens (user_id, fcm_token, created_at, updated_at)
		VALUES (?, ?, ?, ?)
		ON DUPLICATE KEY UPDATE
			user_id = VALUES(user_id),
			updated_at = VALUES(updated_at)
	`
	result, err := dbtx(r.db, ctx).ExecContext(ctx, query, userFcmToken.UserId, userFcmToken.FcmToken, userFcmToken.CreatedAt, userFcmToken.UpdatedAt)
	if err != nil {
		return translateSqlError(err)
	}
	return setLastInsertId(userFcmToken, result)
}

func (r *userFcmTokenRepository) FetchByUserIds(ctx context.Context, userIds []int64) ([]model.UserFcmToken, error) {
	if len(userIds) == 0 {
		return []model.UserFcmToken{}, nil
	}
	stmt := stmtBuilder.Select("*").
		From(model.UserFcmTokenTableName).
		Where(squirrel.Eq{"user_id": userIds})

	tokens := []model.UserFcmToken{}
	if err := fetch(r.db, ctx, &tokens, stmt); err != nil {
		return nil, err
	}
	return tokens, nil
}

func (r *userFcmTokenRepository) Delete(ctx context.Context, userId int64, fcmToken string) error {
	stmt := stmtBuilder.Delete(model.UserFcmTokenTableName).
		Where(squirrel.Eq{"user_id": userId, "fcm_token": fcmToken})
	_, err := exec(r.db, ctx, stmt)
	return err
}

func (r *userFcmTokenRepository) DeleteByToken(ctx context.Context, fcmToken string) error {
	stmt := stmtBuilder.Delete(model.UserFcmTokenTableName).
		Where(squirrel.Eq{"fcm_token": fcmToken})
	_, err := exec(r.db, ctx, stmt)
	return err
}

func (r *userFcmTokenRepository) GetByUserIdAndFcmToken(ctx context.Context, userId int64, fcmToken string) (*model.UserFcmToken, error) {
	stmt := stmtBuilder.Select("*").
		From(model.UserFcmTokenTableName).
		Where(squirrel.Eq{"user_id": userId, "fcm_token": fcmToken}).
		Limit(1)

	token := model.UserFcmToken{}
	if err := get(r.db, ctx, &token, stmt); err != nil {
		return nil, err
	}
	return &token, nil
}
