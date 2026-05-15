package main

import (
	"flag"
	"fmt"
	"log"

	"auction-service/global"
	"auction-service/infrastructure"
)

func main() {
	rollback := flag.Bool("rollback", false, "Roll back migrations")
	steps := flag.Int("steps", 0, "Number of migration steps (0 = all)")
	force := flag.Int("force", -1, "Force migration version (clears dirty state)")
	flag.Parse()

	infraManager := infrastructure.NewInfrastructureManager(global.GetConfig())
	defer infraManager.CloseDB()

	var forcePtr *int
	if *force >= 0 {
		forcePtr = force
	}

	if err := infraManager.MigrateDB(*rollback, *steps, forcePtr); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	fmt.Println("migration completed successfully")
}
