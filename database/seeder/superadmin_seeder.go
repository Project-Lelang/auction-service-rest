package seeder

import (
	"context"
	"errors"
	"log"
	"time"

	"auction-service/constant"
	"auction-service/data_type"
	"auction-service/global"
	"auction-service/infrastructure"
	"auction-service/model"
	"auction-service/repository"
	"auction-service/util"
)

func SeedSuperAdmin(db infrastructure.DBTX) error {
	config := global.GetSuperAdminConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	userRepo := repository.NewUserRepository(db)
	userRoleRepo := repository.NewUserRoleRepository(db)

	_, err := userRepo.GetByPhone(ctx, config.Phone)
	if err != nil && !errors.Is(err, constant.ErrNoData) {
		return err
	}
	if err == nil {
		log.Println("Superadmin already exists, skipping seed")
		return nil
	}

	hashedPassword, err := util.HashPassword(config.Password)
	if err != nil {
		return err
	}

	birth := data_type.NewDateTime(time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC))
	user := &model.User{
		Id:         util.NewUuid(),
		Fullname:   "Super Admin",
		Phone:      config.Phone,
		Birth:      birth,
		IsVerified: true,
		IsDeleted:  false,
		Password:   hashedPassword,
	}

	if err := userRepo.Insert(ctx, user); err != nil {
		return err
	}

	if err := userRoleRepo.Insert(ctx, &model.UserRole{
		Id:     util.NewUuid(),
		UserId: user.Id,
		Role:   constant.RoleSuperAdmin,
	}); err != nil {
		return err
	}

	log.Printf("Superadmin seeded successfully: phone=%s", config.Phone)
	return nil
}
