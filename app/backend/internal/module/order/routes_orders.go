package order

import "github.com/gin-gonic/gin"

// RegisterOrderRoutes mounts the customer order endpoints under the /api group. requireSession also
// enforces the CSRF token for the cancel mutation.
func RegisterOrderRoutes(api *gin.RouterGroup, controller *OrdersController, requireSession gin.HandlerFunc) {
	orders := api.Group("/orders", requireSession)
	orders.GET("", controller.List)
	orders.GET("/:orderId", controller.Get)
	orders.POST("/:orderId/cancel", controller.Cancel)
}
