package repository

import (
	"context"

	"auction-service/data_type"
	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/util"

	"github.com/Masterminds/squirrel"
)

type OtpRepository interface {
	Upsert(ctx context.Context, email, otp string, expiresAt data_type.DateTime) error
	GetByEmail(ctx context.Context, email string) (*model.Otp, error)
	MarkVerified(ctx context.Context, email string) error
}

type otpRepository struct {
	db infrastructure.DBTX
}

func NewOtpRepository(db infrastructure.DBTX) OtpRepository {
	return &otpRepository{db: db}
}

func (r *otpRepository) Upsert(ctx context.Context, email, otpVal string, expiresAt data_type.DateTime) error {
	now := util.CurrentDateTime()
	stmt := stmtBuilder.Insert(model.OtpTableName).
		Columns("email", "otp", "expires_at", "verified", "created_at", "updated_at").
		Values(email, otpVal, expiresAt, false, now, now).
		Suffix("ON DUPLICATE KEY UPDATE otp = VALUES(otp), expires_at = VALUES(expires_at), verified = false, updated_at = VALUES(updated_at)")
	_, err := exec(r.db, ctx, stmt)
	return err
}

func (r *otpRepository) GetByEmail(ctx context.Context, email string) (*model.Otp, error) {
	stmt := stmtBuilder.Select("*").
		From(model.OtpTableName).
		Where(squirrel.Eq{"email": email, "verified": false}).
		Limit(1)

	o := model.Otp{}
	if err := get(r.db, ctx, &o, stmt); err != nil {
		return nil, err
	}
	return &o, nil
}

func (r *otpRepository) MarkVerified(ctx context.Context, email string) error {
	stmt := stmtBuilder.Update(model.OtpTableName).
		Set("verified", true).
		Set("updated_at", util.CurrentDateTime()).
		Where(squirrel.Eq{"email": email})
	_, err := exec(r.db, ctx, stmt)
	return err
}
