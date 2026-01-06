package main

import (
	"GAMERS-BE/internal/global/database"
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

	// Initialize database
	dbConfig := database.NewConfigFromEnv()
	db, err := database.InitDB(dbConfig)
	if err != nil {
		log.Fatal("Failed to initialize database:", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal("Failed to get SQL DB:", err)
	}

	// Get current migration status
	currentVersion, dirty, err := database.GetMigrationVersion(sqlDB, "./db/migrations")
	if err != nil {
		log.Printf("Warning: Could not get current version: %v", err)
	} else {
		log.Printf("Current migration version: %d (dirty: %v)", currentVersion, dirty)
	}

	// Force migration to specified version
	log.Printf("⚠️  Forcing migration to version %d...", version)
	if err := database.ForceMigrationVersion(sqlDB, "./db/migrations", version); err != nil {
		log.Fatal("Failed to force migration:", err)
	}

	// Verify new status
	newVersion, newDirty, err := database.GetMigrationVersion(sqlDB, "./db/migrations")
	if err != nil {
		log.Printf("Warning: Could not verify new version: %v", err)
	} else {
		log.Printf("✅ Migration status: version %d (dirty: %v)", newVersion, newDirty)
	}
}
