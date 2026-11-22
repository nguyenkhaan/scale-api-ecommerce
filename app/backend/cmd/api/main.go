package main

import (
	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/database"
	"cloudian/cloudian-scale/internal/httpserver"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/cart"
	"cloudian/cloudian-scale/internal/module/catalog"
	"cloudian/cloudian-scale/internal/module/flashsale"
	"cloudian/cloudian-scale/internal/module/inventory"
	"cloudian/cloudian-scale/internal/module/order"
	"cloudian/cloudian-scale/internal/module/payment"
	"context"
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
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
	//Khoi tao service moi o day, sau do tien hanh truyen vao ben trong depedency
	authService, err := auth.NewService(postgresDatabase.ORM(), auth.DefaultPasswordCost)
	if err != nil {
		return err
	}
	catalogService, err := catalog.NewService(postgresDatabase.ORM())
	if err != nil {
		return err
	}
	inventoryService, err := inventory.NewService(postgresDatabase.ORM())
	if err != nil {
		return err
	}
	flashSaleService, err := flashsale.NewService(postgresDatabase.ORM())
	if err != nil {
		return err
	}
	cartService, err := cart.NewService(postgresDatabase.ORM())
	if err != nil {
		return err
	}
	orderQueries, err := order.NewOrderQueries(postgresDatabase.ORM())
	if err != nil {
		return err
	}
	paymentService, err := newPaymentService(postgresDatabase.ORM())
	if err != nil {
		return err
	}
	//Khoi tao dependencies
	router := httpserver.NewRouter(httpserver.Dependencies{
		Database:         postgresDatabase,
		AuthService:      authService,
		CatalogService:   catalogService,
		InventoryService: inventoryService,
		FlashSaleService: flashSaleService,
		CartService:      cartService,
		OrderQueries:     orderQueries,
		PaymentService:   paymentService,
		SecureCookies:    !config.IsLocalEnvironment(applicationConfig.Env),
	})
	// Thuc hien dang ky api docs.
	if applicationConfig.Env == "dev" || applicationConfig.Env == "development" {
		httpserver.RegisterOpenAPIDocs(router, "../../docs/technical/api/openapi.yaml")
	}
	return router.Run(fmt.Sprintf(":%v", applicationConfig.Port))
}

// newPaymentService wires the simulator to deliver over HTTP to the real callback endpoint, so signature
// verification and event persistence are exercised exactly as an external gateway would.
func newPaymentService(orm *gorm.DB) (*payment.Service, error) {
	paymentConfig, err := payment.LoadConfig()
	if err != nil {
		return nil, err
	}
	var sender payment.Sender
	if paymentConfig.SimulatorEnabled {
		sender = payment.NewHTTPSender(paymentConfig.CallbackURL, nil)
	}
	return payment.NewService(orm, paymentConfig, sender)
}
