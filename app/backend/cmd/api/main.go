package main

import (
	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/database"
	"cloudian/cloudian-scale/internal/health"
	"context"
	"fmt"
	"log"
	"time"
)
func main() {
	if err := run(); err != nil {
		log.Fatal("Error: %v" , err);
	} 
}
func run() error {
	//Loading environment 
	applicationConfig, err := config.Load();
	if err != nil {
		return err
	}
	
	startupContext, cancel := context.WithTimeout(
		context.Background(), 15 * time.Second,
	)
	defer cancel()  //Hoan viec cancel cho den khi ham run() ket thuc
	postgresDatabase, err := database.Open(startupContext , applicationConfig.DatabaseUrl)
	if err != nil {
		return err
	}
	defer postgresDatabase.Close() //Hoan viec thuc thi cho den khi ham run() ket thuc 
	router := health.HealthRouter(postgresDatabase)
	return router.Run(fmt.Sprintf(":%v" , applicationConfig.Port))
}