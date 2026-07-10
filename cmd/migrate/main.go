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
	fresh := flag.Bool("fresh", false, "Drop all migrated tables and run all migrations again (destructive)")
	flag.Parse()

	infraManager := infrastructure.NewInfrastructureManager(global.GetConfig())
	defer infraManager.Close()

	if *fresh {
		if *rollback || *steps != 0 || *force >= 0 {
			log.Fatal("-fresh cannot be combined with -rollback, -steps, or -force")
		}
		if err := infraManager.RefreshDB(); err != nil {
			log.Fatalf("fresh migration failed: %v", err)
		}
		fmt.Println("fresh migration and Redis queue reset completed successfully")
		return
	}

	var forcePtr *int
	if *force >= 0 {
		forcePtr = force
	}

	if err := infraManager.MigrateDB(*rollback, *steps, forcePtr); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	fmt.Println("migration completed successfully")
}
