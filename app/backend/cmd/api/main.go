package main

import (
	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/database"
	"cloudian/cloudian-scale/internal/health"
	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func registerRoutes(database *database.Database) *gin.Engine {
	router := gin.New()
	router.Use(
		middleware.RequestIDMiddleware(),
		gin.LoggerWithFormatter(middleware.RequestLogFormatter),
		gin.Recovery(),
	)
	api := router.Group("/api")
	health.RegisterRoutes(api, database)
	return router
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

	return registerRoutes(postgresDatabase).Run(fmt.Sprintf(":%v", applicationConfig.Port))
}
