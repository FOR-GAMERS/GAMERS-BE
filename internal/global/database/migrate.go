package database

import (
	"database/sql"
	"fmt"
	"log"
	"path/filepath"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/mysql"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

// RunMigrations executes database migrations
func RunMigrations(db *sql.DB, migrationsPath string) error {
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil && err != migrate.ErrNilVersion {
		return fmt.Errorf("could not get migration version: %w", err)
	}

	log.Printf("Current migration version: %d (dirty: %v)", version, dirty)

	// Migration 실행
	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Println("✅ No new migrations to apply")
			return nil
		}
		return fmt.Errorf("could not run migrations: %w", err)
	}

	newVersion, _, _ := m.Version()
	log.Printf("✅ Migrations completed successfully (version: %d)", newVersion)
	return nil
}

// RollbackMigration rolls back the last migration
func RollbackMigration(db *sql.DB, migrationsPath string, steps int) error {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Steps(-steps); err != nil {
		return fmt.Errorf("could not rollback migrations: %w", err)
	}

	log.Printf("✅ Rolled back %d migration(s)", steps)
	return nil
}

// GetMigrationVersion returns current migration version
func GetMigrationVersion(db *sql.DB, migrationsPath string) (uint, bool, error) {
	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"mysql",
		driver,
	)
	if err != nil {
		return 0, false, fmt.Errorf("could not create migrate instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		if err == migrate.ErrNilVersion {
			return 0, false, nil
		}
		return 0, false, err
	}

	return version, dirty, nil
}

// ForceMigrationVersion forces database to specific version (useful for dirty state)
func ForceMigrationVersion(db *sql.DB, migrationsPath string, version int) error {
	absPath, err := filepath.Abs(migrationsPath)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	driver, err := mysql.WithInstance(db, &mysql.Config{})
	if err != nil {
		return fmt.Errorf("could not create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", absPath),
		"mysql",
		driver,
	)
	if err != nil {
		return fmt.Errorf("could not create migrate instance: %w", err)
	}

	if err := m.Force(version); err != nil {
		return fmt.Errorf("could not force version: %w", err)
	}

	log.Printf("✅ Forced migration to version %d", version)
	return nil
}
