package migration

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"strings"

	"github.com/pressly/goose/v3"
)

const (
	CommandStatus = "status"
	CommandUp     = "up"
	migrationDir  = "sql"
)

//go:embed sql/*.sql
var migrationFiles embed.FS

func Run(ctx context.Context, database *sql.DB, command string) error {
	if command != CommandUp && command != CommandStatus {
		return fmt.Errorf("unsupported migration command %q", command)
	}
	if database == nil {
		return errors.New("migration database is required")
	}
	if err := configureGoose(); err != nil {
		return err
	}
	if err := goose.RunContext(ctx, command, database, migrationDir); err != nil {
		return fmt.Errorf("run migration command %q: %w", command, err)
	}
	return nil
}

func Reset(ctx context.Context, database *sql.DB, environment, confirmation string) error {
	if !CanResetDatabase(environment, confirmation) {
		return errors.New("database reset requires ENV=test and ALLOW_DATABASE_RESET=true")
	}
	if database == nil {
		return errors.New("migration database is required")
	}
	if err := configureGoose(); err != nil {
		return err
	}
	if err := goose.RunContext(ctx, "reset", database, migrationDir); err != nil {
		return fmt.Errorf("reset test database: %w", err)
	}
	return nil
}

func CanResetDatabase(environment, confirmation string) bool {
	return strings.EqualFold(strings.TrimSpace(environment), "test") &&
		strings.EqualFold(strings.TrimSpace(confirmation), "true")
}

func configureGoose() error {
	goose.SetBaseFS(migrationFiles)
	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("configure PostgreSQL migrations: %w", err)
	}
	return nil
}
