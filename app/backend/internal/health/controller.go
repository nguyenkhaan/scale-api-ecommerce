package health

import (
	"context"
	"net/http"
	"github.com/gin-gonic/gin"
)

type DatabasePinger interface {
	Ping(ctx context.Context) error
}
//interface nhan vao 1 gia tri, ma gia tri do da thuc hien receive cac ham duoc khai bao ben trong interface
func HealthRouter(database DatabasePinger) *gin.Engine {
	router := gin.New() 
	router.Use(gin.Logger() , gin.Recovery()) 
	router.GET("/health/liveness" , func(c *gin.Context) {
		c.JSON(http.StatusOK , gin.H{ //map 
			"message": "Build with Cloudian Love Cloud",
		})
	}) 
	router.GET("/health/ready" , func(c *gin.Context) {
		requestContext := c.Request.Context()
		if err := database.Ping(requestContext); err != nil {
			c.JSON(http.StatusServiceUnavailable , gin.H{
				"status": "service down",
			})
		}
		c.JSON(http.StatusOK , gin.H{
			"status": "service up",
		})
	})
	return router
}