package main

import (
	"github.com/FOR-GAMERS/GAMERS-BE/internal/global/config"
	"fmt"
	"log"
	"os"
	"strconv"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run scripts/force-migration.go <version>")
		fmt.Println("Example: go run scripts/force-migration.go 3")
		os.Exit(1)
	}

	version, err := strconv.Atoi(os.Args[1])
	if err != nil {
		log.Fatal("Invalid version number:", err)
	}

	// Initialize config
	dbConfig := config.NewConfigFromEnv()
	db, err := config.InitDB(dbConfig)
	if err != nil {
		log.Fatal("Failed to initialize config:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get SQL DB:", err)
	}

	// Get current migration status
	currentVersion, dirty, err := config.GetMigrationVersion(sqlDB, "./db/migrations")
	if err != nil {
		log.Printf("Warning: Could not get current version: %v", err)
	} else {
		log.Printf("Current migration version: %d (dirty: %v)", currentVersion, dirty)
	}

	// Force migration to specified version
	log.Printf("⚠️  Forcing migration to version %d...", version)
	if err := config.ForceMigrationVersion(sqlDB, "./db/migrations", version); err != nil {
		log.Fatal("Failed to force migration:", err)
	}

	// Verify new status
	newVersion, newDirty, err := config.GetMigrationVersion(sqlDB, "./db/migrations")
	if err != nil {
		log.Printf("Warning: Could not verify new version: %v", err)
	} else {
		log.Printf("✅ Migration status: version %d (dirty: %v)", newVersion, newDirty)
	}
}
