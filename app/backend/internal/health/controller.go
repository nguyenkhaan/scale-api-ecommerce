package health

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

type DatabasePinger interface {
	Ping(ctx context.Context) error
}

func RegisterRoutes(router *gin.RouterGroup, database DatabasePinger) *gin.RouterGroup {
	router.GET("/health/liveness", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Build with Cloudian 💙 Cloud",
		})
	})
	router.GET("/health/ready", func(c *gin.Context) {
		requestContext := c.Request.Context()
		if err := database.Ping(requestContext); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status": "service down",
			})
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"status": "service up",
		})
	})
	return router
}
