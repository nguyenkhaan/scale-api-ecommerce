package database

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type Database struct {
	orm  *gorm.DB
	pool *sql.DB
}

func Open(
	ctx context.Context, connectionString string,
) (*Database, error) {
	orm, err := gorm.Open(
		postgres.New(postgres.Config{
			DSN:                  connectionString,
			PreferSimpleProtocol: true,
		}),
		&gorm.Config{
			DisableAutomaticPing: true,
			TranslateError:       true,
			Logger:               newQueryLogger(),
		},
	)
	if err != nil {
		return nil, fmt.Errorf("open Neon PostgreSQL with GORM: %w", err)
	}
	pool, err := orm.DB()
	if err != nil {
		return nil, fmt.Errorf("get SQL connection pool: %w", err)
	}
	configurePool(pool)

	if err := pool.PingContext(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping Neon PostgreSQL: %w", err)
	}
	return &Database{
		orm:  orm,
		pool: pool,
	}, nil
}

// newQueryLogger logs failed and slow statements without their bound values, so emails, password hashes
// and token hashes never reach the logs. Missing rows are expected (unknown emails) and are not logged.
func newQueryLogger() logger.Interface {
	return logger.New(log.New(os.Stdout, "", log.LstdFlags), logger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  logger.Warn,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
	})
}

func configurePool(pool *sql.DB) {
	pool.SetMaxOpenConns(10)
	pool.SetMaxIdleConns(2)
	pool.SetConnMaxIdleTime(5 * time.Minute)
	pool.SetConnMaxLifetime(30 * time.Minute)
}

func (database *Database) ORM() *gorm.DB {
	return database.orm
}

func (database *Database) SQL() *sql.DB {
	return database.pool
}

func (database *Database) Ping(ctx context.Context) error {
	return database.pool.PingContext(ctx)
}

func (database *Database) Close() error {
	return database.pool.Close()
}
