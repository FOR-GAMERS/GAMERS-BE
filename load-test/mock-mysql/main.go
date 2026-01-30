package main

import (
	"fmt"
	"log"

	sqle "github.com/dolthub/go-mysql-server"
	"github.com/dolthub/go-mysql-server/memory"
	"github.com/dolthub/go-mysql-server/server"
	"github.com/dolthub/go-mysql-server/sql"
	"github.com/dolthub/go-mysql-server/sql/information_schema"
)

func main() {
	dbProvider := memory.NewDBProvider(
		memory.NewDatabase("gamers"),
		information_schema.NewInformationSchemaDatabase(),
	)

	engine := sqle.NewDefault(dbProvider)

	ctx := sql.NewEmptyContext()
	ctx.SetCurrentDatabase("gamers")

	for _, stmt := range createTableStatements {
		_, _, _, err := engine.Query(ctx, stmt)
		if err != nil {
			log.Fatalf("Failed to create table: %v\nSQL: %s", err, stmt)
		}
	}

	log.Println("All 12 tables created successfully")

	// Insert seed data for load testing
	for _, stmt := range seedDataStatements {
		_, _, _, err := engine.Query(ctx, stmt)
		if err != nil {
			log.Printf("Warning: seed data insert failed: %v\nSQL: %s", err, stmt)
		}
	}
	log.Printf("Seed data inserted: %d statements executed", len(seedDataStatements))

	config := server.Config{
		Protocol: "tcp",
		Address:  "0.0.0.0:3306",
	}

	s, err := server.NewServer(config, engine, memory.NewSessionBuilder(dbProvider), nil)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	fmt.Println("Mock MySQL server listening on 0.0.0.0:3306")
	fmt.Println("Database: gamers | User: root (no password)")

	if err := s.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
