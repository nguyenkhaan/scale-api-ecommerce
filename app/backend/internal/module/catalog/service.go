package catalog

import (
	"cloudian/cloudian-scale/internal/model"
	"context"
	"errors"
	"fmt"

	"gorm.io/gorm"
)

type Service struct {
	database *gorm.DB
}
func NewService(database *gorm.DB) (*Service, error) {
	if database == nil {
		return nil, errors.New("catalog database is required")
	}
	return &Service{database: database}, nil
}

var ErrProductNotFound = errors.New("product not found")
//Cac buoc tien hanh: viet controller -> viet cac ham controller -> goi service -> viet ham service & khai bao dto 
//Thuc hien khai bao NewController ben trong router.go, file nay se thuc hien inject depedency 
//Thuc hien khai bao 

func (service *Service) GetPublicProduct(
	ctx context.Context,
	productID string,
) (ProductResponse, error) {
	database := service.database.WithContext(ctx)

	var product model.Product
	err := database.
		Model(&model.Product{}).
		Select("products.*").
		Joins("LEFT JOIN categories ON categories.id = products.category_id").
		Where(`
			products.id = ?
			AND products.status = ?
			AND (
				products.category_id IS NULL
				OR categories.is_visible = TRUE
			)
		`, productID, model.ProductStatusActive).
		Take(&product).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return ProductResponse{}, ErrProductNotFound
	}
	if err != nil {
		return ProductResponse{}, fmt.Errorf("find public product: %w", err)
	}

	var skus []model.SKU
	if err := database.
		Where("product_id = ? AND status = ?", product.ID, model.SKUStatusActive).
		Order("id ASC").
		Find(&skus).Error; err != nil {
		return ProductResponse{}, fmt.Errorf("find public product SKUs: %w", err)
	}

	// Product không có SKU active thì khách không thể mua.
	if len(skus) == 0 {
		return ProductResponse{}, ErrProductNotFound
	}

	return newProductResponse(product, skus), nil
}