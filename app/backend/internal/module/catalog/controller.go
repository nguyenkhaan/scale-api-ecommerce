package catalog

import (
	"cloudian/cloudian-scale/internal/httpserver/problem"
	"errors"
	"net/http"
	"uuid"

	"github.com/gin-gonic/gin"
)

type Controller struct {
	service *Service
} 

func NewController(service *Service) *Controller {
	return &Controller{service : service}
}

func RegisterRoutes(router *gin.RouterGroup , controller *Controller) {
	group := router.Group("/v1/products")
	group.GET("/:productId" , controller.GetProduct)
}

//Ham nay phai nhan vao gin.Context 
func (controller *Controller) GetProduct(context *gin.Context) {
	//Thuc hien viet logic de lay ra duoc chi tiet san pham 
	productID := context.Param("productId")
	if _, err := uuid.Parse(productID); err != nil {
		problem.AbortValidation(context , []problem.Violation{
			{
				Field: "productId",
				Reason: "must be a valid uuid",
			},  //Cu phap khai bao mang ben trong, chung ta hoan toan co the viet problem.Violation{...} nhung 
			//vi day la ben trong mang, da ro duoc KDL roi nen Golang cho phep luoc bo 
		})
	}
	product, err := controller.service.GetPublicProduct(context.Request.Context() , productID)
	switch {
		case errors.Is(err, ErrProductNotFound): 
			problem.Abort(context , http.StatusNotFound , problem.CodeNotFound , "Product was not found")
		case err != nil:
			problem.AbortInternal(context, err)
		default: context.JSON(http.StatusOK , product)
	}

}