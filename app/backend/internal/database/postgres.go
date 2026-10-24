package database

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Database struct {
	orm *gorm.DB
	pool *sql.DB
}
func Open(
	ctx context.Context, connectionString string,
) (*Database, error) {
	orm, err := gorm.Open(
		postgres.New(postgres.Config{
			DSN: connectionString,
			PreferSimpleProtocol: true,
		}), 
		&gorm.Config{
			DisableAutomaticPing: true, //Tat hanh vi nay de tu goi Ping(ctx Context)
			TranslateError: true,
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
		orm: orm, 
		pool: pool,
	}, nil
}
func configurePool(pool *sql.DB) {
	pool.SetMaxOpenConns(10)
	pool.SetMaxIdleConns(2)
	pool.SetConnMaxIdleTime(5 * time.Minute)
	pool.SetConnMaxLifetime(30 * time.Minute)
}
//receiver function, thuong su dung con tro
func (database *Database) ORM() *gorm.DB {
	return database.orm //Contro struct tu dong deference khi truy cap gia tri => Khong can dung (*database) van duoc
} 
func (database *Database) Ping(ctx context.Context) error {
	return database.pool.PingContext(ctx) //Trong Go, khong dung con tro tro den interface, tru khi co li do dac biet 

}
func (database *Database) Close() error {
	return database.pool.Close()
}