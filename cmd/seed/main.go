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
		log.Fatalf("seeder failed: %v", err)
	}

	fmt.Println("seeder completed successfully")
}
