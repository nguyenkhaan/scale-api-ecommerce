package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/database"
	"cloudian/cloudian-scale/internal/migration"
)

const commandTimeout = 60 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	if len(os.Args) != 2 {
		return errors.New("usage: go run ./cmd/migrate <up|status|reset>")
	}
	applicationConfig, err := config.Load()
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), commandTimeout)
	defer cancel()

	postgresDatabase, err := database.Open(ctx, applicationConfig.DatabaseUrl)
	if err != nil {
		return err
	}
	defer postgresDatabase.Close()

	if os.Args[1] == "reset" {
		return migration.Reset(
			ctx,
			postgresDatabase.SQL(),
			applicationConfig.Env,
			os.Getenv("ALLOW_DATABASE_RESET"),
		)
	}
	return migration.Run(ctx, postgresDatabase.SQL(), os.Args[1])
}
