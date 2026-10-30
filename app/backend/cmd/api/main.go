package main

import (
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/database"
	"cloudian/cloudian-scale/internal/httpserver"
	"context"
	"fmt"
	"log"
	"time"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	applicationConfig, err := config.Load()
	if err != nil {
		return err
	}
	startupContext, cancel := context.WithTimeout(
		context.Background(), 15*time.Second,
	)
	defer cancel()
	postgresDatabase, err := database.Open(startupContext, applicationConfig.DatabaseUrl)
	if err != nil {
		return err
	}
	defer postgresDatabase.Close()

	authService, err := auth.NewService(postgresDatabase.ORM(), auth.DefaultPasswordCost)
	if err != nil {
		return err
	}
	router := httpserver.NewRouter(httpserver.Dependencies{
		Database:      postgresDatabase,
		AuthService:   authService,
		SecureCookies: !config.IsLocalEnvironment(applicationConfig.Env),
	})
	return router.Run(fmt.Sprintf(":%v", applicationConfig.Port))
}
