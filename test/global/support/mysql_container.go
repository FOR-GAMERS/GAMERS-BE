package support

import (
	"context"
	"fmt"
	"time"

	"github.com/testcontainers/testcontainers-go/modules/mysql"
	mysqldriver "gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type MySQLContainer struct {
	container *mysql.MySQLContainer
	db        *gorm.DB
}

func SetupMySQLContainer(ctx context.Context) (*MySQLContainer, error) {
	mysqlContainer, err := mysql.Run(ctx,
		"mysql:8.0",
		mysql.WithDatabase("testdb"),
		mysql.WithUsername("testuser"),
		mysql.WithPassword("testpass"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to start MySQL container: %w", err)
	}

	connectionString, err := mysqlContainer.ConnectionString(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get connection string: %w", err)
	}

	connectionString += "?parseTime=true"

	var db *gorm.DB
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		db, err = gorm.Open(mysqldriver.Open(connectionString), &gorm.Config{})
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		mysqlContainer.Terminate(ctx)
		return nil, fmt.Errorf("failed to connect to MySQL: %w", err)
	}

	return &MySQLContainer{
		container: mysqlContainer,
		db:        db,
	}, nil
}

func (m *MySQLContainer) GetDB() *gorm.DB {
	return m.db
}

func (m *MySQLContainer) Teardown(ctx context.Context) error {
	if m.container != nil {
		return m.container.Terminate(ctx)
	}
	return nil
}
