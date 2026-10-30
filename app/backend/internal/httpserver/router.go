package httpserver

import (
	"cloudian/cloudian-scale/internal/httpserver/middleware"
	"cloudian/cloudian-scale/internal/module/auth"
	"cloudian/cloudian-scale/internal/module/health"
	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	Database      health.DatabasePinger
	AuthService   *auth.Service
	SecureCookies bool
}

func NewRouter(dependencies Dependencies) *gin.Engine {
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
	//Thuc hien RegisterRoute o day 
	return router
}
