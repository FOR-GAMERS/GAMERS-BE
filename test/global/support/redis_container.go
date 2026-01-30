package support

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

type RedisContainer struct {
	container testcontainers.Container
	client    *redis.Client
}

func SetupRedisContainer(ctx context.Context) (*RedisContainer, error) {
	req := testcontainers.ContainerRequest{
		Image:        "redis:7-alpine",
		ExposedPorts: []string{"6379/tcp"},
		WaitingFor:   wait.ForLog("Ready to accept connections"),
	}
	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to start Redis container: %w", err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get host: %w", err)
	}

	port, err := container.MappedPort(ctx, "6379")
	if err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to get port: %w", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: fmt.Sprintf("%s:%s", host, port.Port()),
	})

	if err := client.Ping(ctx).Err(); err != nil {
		container.Terminate(ctx)
		return nil, fmt.Errorf("failed to ping Redis: %w", err)
	}

	return &RedisContainer{
		container: container,
		client:    client,
	}, nil
}

func (r *RedisContainer) GetClient() *redis.Client {
	return r.client
}

func (r *RedisContainer) Teardown(ctx context.Context) error {
	r.client.Close()
	if r.container != nil {
		return r.container.Terminate(ctx)
	}
	return nil
}
