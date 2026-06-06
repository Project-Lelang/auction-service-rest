package main

import (
	"fmt"
	"log"

	"auction-service/database/seeder"
	"auction-service/global"
	"auction-service/infrastructure"
)

func main() {
	infraManager := infrastructure.NewInfrastructureManager(global.GetConfig())
	defer infraManager.CloseDB()

	if err := seeder.SeedSuperAdmin(infraManager.GetDB()); err != nil {
		log.Fatalf("superadmin seeder failed: %v", err)
	}

	if err := seeder.SeedPaymentMethods(infraManager.GetDB()); err != nil {
		log.Fatalf("payment methods seeder failed: %v", err)
	}

	fmt.Println("seeder completed successfully")
}
