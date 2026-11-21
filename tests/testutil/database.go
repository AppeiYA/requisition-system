package testutil

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
)

const (
	postgresImage = "postgres:16-alpine"

	testDBName     = "integration_test"
	testDBUser     = "test_user"
	testDBPassword = "test_password"
)

type TestDatabase struct {
	Pool      *pgxpool.Pool
	Container *postgres.PostgresContainer
	DSN       string
}

func SetupTestDatabase(ctx context.Context) (*TestDatabase, error) {
	container, err := postgres.Run(
		ctx,
		postgresImage,
		postgres.WithDatabase(testDBName),
		postgres.WithUsername(testDBUser),
		postgres.WithPassword(testDBPassword),
		postgres.BasicWaitStrategies(),
	)

	if err != nil {
		return nil, fmt.Errorf("start postgres container: %w", err)
	}

	connString, err := container.ConnectionString(
		ctx,
		"sslmode=disable",
	)

	if err != nil {
		_ = terminateContainer(container)
		return nil, fmt.Errorf("get postgres connection string: %w", err)
	}

	poolConfig, err := pgxpool.ParseConfig(connString)
	if err != nil {
		_ = terminateContainer(container)
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	poolConfig.MaxConns = 5
	poolConfig.MinConns = 1

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		_ = terminateContainer(container)
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		_ = terminateContainer(container)

		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return &TestDatabase{
		Pool:      pool,
		Container: container,
		DSN:       connString,
	}, nil
}

func (db *TestDatabase) Close() error {
	if db == nil {
		return nil
	}

	if db.Pool != nil {
		db.Pool.Close()
	}

	if db.Container != nil {
		return terminateContainer(db.Container)
	}

	return nil
}

func terminateContainer(container *postgres.PostgresContainer) error {
	if container == nil {
		return nil
	}

	cleanupCtx, cancel := context.WithTimeout(
		context.Background(),
		30*time.Second,
	)
	defer cancel()

	if err := container.Terminate(cleanupCtx); err != nil {
		return fmt.Errorf("terminate postgres container: %w", err)
	}

	return nil
}
