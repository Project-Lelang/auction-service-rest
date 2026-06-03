package seeder

import (
	"context"
	"errors"
	"log"
	"time"

	"auction-service/constant"
	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"
)

// SeedPaymentMethods inserts the default Midtrans payment method if it does not
// already exist.  Idempotent: safe to call on every deployment.
func SeedPaymentMethods(db infrastructure.DBTX) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	repo := repository.NewPaymentMethodRepository(db)

	_, err := repo.GetByCode(ctx, constant.PaymentMethodCodeMidtrans)
	if err != nil && !errors.Is(err, constant.ErrNoData) {
		return err
	}
	if err == nil {
		log.Println("Midtrans payment method already exists, skipping seed")
		return nil
	}

	pm := &model.PaymentMethod{
		Id:       util.NewUuid(),
		Code:     constant.PaymentMethodCodeMidtrans,
		Type:     constant.PaymentMethodTypeMidtrans,
		Name:     "Midtrans",
		IsActive: true,
	}

	if err := repo.Insert(ctx, pm); err != nil {
		return err
	}

	log.Printf("Midtrans payment method seeded with id=%s", pm.Id)
	return nil
}
