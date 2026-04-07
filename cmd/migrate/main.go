package main

import (
	"fmt"
	"log"

	"auction-service/global"
	"auction-service/infrastructure"
)

func main() {
	infraManager := infrastructure.NewInfrastructureManager(global.GetConfig())
	defer infraManager.CloseDB()

	if err := infraManager.MigrateDB(false, 0, nil); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	fmt.Println("migration completed successfully")
}
