package seeder

import (
	"context"
	"fmt"
	"log"
	"time"

	"auction-service/global"
	"auction-service/model"
	"auction-service/util"

	"github.com/jmoiron/sqlx"
)

func SeedSuperAdmin(db *sqlx.DB) error {
	config := global.GetSuperAdminConfig()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// check if superadmin already exists
	var count int
	err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM users WHERE email = $1", config.Email).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check existing superadmin: %w", err)
	}

	if count > 0 {
		log.Println("Superadmin already exists, skipping seed")
		return nil
	}

	hashedPassword, err := util.HashPassword(config.Password)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	now := time.Now().UTC()
	user := &model.User{
		Id:        util.NewUuid(),
		Name:      "Super Admin",
		Email:     config.Email,
		Password:  hashedPassword,
		Role:      model.UserRoleSuperAdmin,
		IsActive:  true,
		CreatedAt: now,
		UpdatedAt: now,
	}

	_, err = db.ExecContext(ctx,
		`INSERT INTO users (id, name, email, password, role, is_active, created_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, $8)`,
		user.Id, user.Name, user.Email, user.Password, user.Role, user.IsActive, user.CreatedAt, user.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("failed to insert superadmin: %w", err)
	}

	log.Printf("Superadmin seeded successfully: %s", user.Email)
	return nil
}
