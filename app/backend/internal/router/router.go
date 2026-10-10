package router

import (
	"github.com/gin-gonic/gin"

	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/errors"
	"cloudian/cloudian-scale/internal/handler"
	"cloudian/cloudian-scale/internal/middleware"
	"cloudian/cloudian-scale/internal/response"
)

// APIPrefix là tiền tố chung của mọi endpoint; proxy phía frontend cũng định tuyến theo giá trị này
const APIPrefix = "/api"

// SetupRouter cấu hình toàn bộ middleware và routes cho backend SaleScale
func SetupRouter(cfg *config.Config) *gin.Engine {
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()
	r.HandleMethodNotAllowed = true

	// Global middlewares
	r.Use(middleware.RequestIDMiddleware())
	r.Use(middleware.LoggerMiddleware())
	r.Use(middleware.RecoveryMiddleware())
	r.Use(middleware.CORSMiddleware(cfg.CORSAllowedOrigins))

	// Xử lý 404 Not Found theo chuẩn Response
	r.NoRoute(func(c *gin.Context) {
		appErr := errors.ErrNotFound(c.Request.URL.Path)
		response.Error(c, appErr)
	})

	// Xử lý 405 Method Not Allowed theo chuẩn Response
	r.NoMethod(func(c *gin.Context) {
		appErr := errors.ErrMethodNotAllowed(c.Request.Method)
		response.Error(c, appErr)
	})

	// Khởi tạo handlers
	healthHandler := handler.NewHealthHandler(cfg)

	// Định nghĩa nhóm route với prefix /api
	api := r.Group(APIPrefix)
	{
		api.GET("/health", healthHandler.HealthCheck)
		api.GET("/ping", healthHandler.Ping)
	}

	return r
}
