package httpserver

import (
	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/catalog"
	"cloudian/cloudian-scale/internal/module/health"
	"cloudian/cloudian-scale/internal/module/inventory"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Database      health.DatabasePinger
	AuthService   *auth.Service
	InventoryService *inventory.Service
	CatalogService *catalog.Service
	SecureCookies bool
}

func NewRouter(dependencies Dependencies) *gin.Engine {
	//Khai bao middleware function d su dung 
	requireSession := middleware.RequireSession(dependencies.AuthService)
	requireAdmin := middleware.RequireRole(model.RoleAdmin)
	//Bat dau khai bao router 
	router := gin.New()
	router.Use(
		middleware.RequestIDMiddleware(),
		gin.LoggerWithFormatter(middleware.RequestLogFormatter),
		gin.Recovery(),
	)
	api := router.Group("/api")
	health.RegisterRoutes(api, dependencies.Database)
	auth.RegisterRoutes(
		api,
		auth.NewController(dependencies.AuthService, dependencies.SecureCookies),
		middleware.RequireSession(dependencies.AuthService),
	)
	catalog.RegisterRoutes(
		api,
		catalog.NewController(dependencies.CatalogService),
	)
	inventory.RegisterRoutes(
		api, 
		inventory.NewController(dependencies.InventoryService),
		requireSession,
		requireAdmin,
	)
	//Thuc hien RegisterRoute o day 
	return router
}
