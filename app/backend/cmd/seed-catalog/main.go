package main

import (
	"context"
	"log"
	"time"

	"cloudian/cloudian-scale/internal/bootstrap"
	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/database"
)

const commandTimeout = 60 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
	log.Print("catalog and inventory seed completed successfully")
}

func run() error {
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

	return bootstrap.SeedCatalog(ctx, postgresDatabase.ORM())
}
