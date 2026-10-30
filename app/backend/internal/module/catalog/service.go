package catalog

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"cloudian/cloudian-scale/internal/model"

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

var (
	ErrProductNotFound  = errors.New("product not found")
	ErrSKUNotFound      = errors.New("SKU not found")
	ErrCategoryNotFound = errors.New("category not found")
	ErrSKUCodeTaken     = errors.New("SKU code already exists")
)

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

	if len(skus) == 0 {
		return ProductResponse{}, ErrProductNotFound
	}

	return newProductResponse(product, skus), nil
}

func (service *Service) ListPublicProducts(
	ctx context.Context,
	query ListPublicProductsQuery,
) (ProductPageResponse, error) {
	if query.Limit == 0 {
		query.Limit = 20
	}

	filter := publicFilterFingerprint(query.Query, query.CategoryID)
	var cursor productCursor
	if query.Cursor != "" {
		decoded, err := decodeProductCursor(query.Cursor, filter)
		if err != nil {
			return ProductPageResponse{}, err
		}
		cursor = decoded
	}

	database := service.database.WithContext(ctx).
		Model(&model.Product{}).
		Select("products.*").
		Joins("LEFT JOIN categories ON categories.id = products.category_id").
		Where(
			`products.status = ?
				AND (products.category_id IS NULL OR categories.is_visible = TRUE)
				AND EXISTS (
					SELECT 1 FROM skus
					WHERE skus.product_id = products.id AND skus.status = ?
				)`,
			model.ProductStatusActive,
			model.SKUStatusActive,
		)

	if query.CategoryID != nil {
		database = database.Where("products.category_id = ?", *query.CategoryID)
	}
	if query.Query != "" {
		database = database.Where(
			"products.name ILIKE ? ESCAPE E'\\\\'",
			"%"+escapeLikeLiteral(query.Query)+"%",
		)
	}
	if query.Cursor != "" {
		database = database.Where(
			"(products.created_at, products.id) < (?, ?)",
			cursor.CreatedAt,
			cursor.ID,
		)
	}

	var products []model.Product
	if err := database.
		Order("products.created_at DESC, products.id DESC").
		Limit(query.Limit + 1).
		Find(&products).Error; err != nil {
		return ProductPageResponse{}, fmt.Errorf("list public products: %w", err)
	}

	page := ProductPageResponse{Items: make([]ProductResponse, 0, min(len(products), query.Limit))}
	if len(products) > query.Limit {
		products = products[:query.Limit]
		last := products[len(products)-1]
		nextCursor, err := encodeProductCursor(last.CreatedAt, last.ID, filter)
		if err != nil {
			return ProductPageResponse{}, fmt.Errorf("encode public product cursor: %w", err)
		}
		page.NextCursor = &nextCursor
	}
	if len(products) == 0 {
		return page, nil
	}

	productIDs := make([]string, 0, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}
	var skus []model.SKU
	if err := service.database.WithContext(ctx).
		Where("product_id IN ? AND status = ?", productIDs, model.SKUStatusActive).
		Order("product_id ASC, id ASC").
		Find(&skus).Error; err != nil {
		return ProductPageResponse{}, fmt.Errorf("list public product SKUs: %w", err)
	}
	skusByProduct := make(map[string][]model.SKU, len(products))
	for _, sku := range skus {
		skusByProduct[sku.ProductID] = append(skusByProduct[sku.ProductID], sku)
	}
	for _, product := range products {
		page.Items = append(page.Items, newProductResponse(product, skusByProduct[product.ID]))
	}

	return page, nil
}

func (service *Service) GetAdminProduct(ctx context.Context, productID string) (ProductResponse, error) {
	var product model.Product
	if err := service.database.WithContext(ctx).Where("id = ?", productID).Take(&product).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ProductResponse{}, ErrProductNotFound
		}
		return ProductResponse{}, fmt.Errorf("find admin product: %w", err)
	}

	var skus []model.SKU
	if err := service.database.WithContext(ctx).
		Where("product_id = ?", product.ID).
		Order("id ASC").
		Find(&skus).Error; err != nil {
		return ProductResponse{}, fmt.Errorf("find admin product SKUs: %w", err)
	}
	return newProductResponse(product, skus), nil
}

func (service *Service) ListAdminProducts(
	ctx context.Context,
	query AdminProductListQuery,
) (ProductPageResponse, error) {
	if query.Limit == 0 {
		query.Limit = 20
	}
	filter := adminProductFilterFingerprint(query.Status)
	var cursor productCursor
	if query.Cursor != "" {
		decoded, err := decodeProductCursor(query.Cursor, filter)
		if err != nil {
			return ProductPageResponse{}, err
		}
		cursor = decoded
	}

	database := service.database.WithContext(ctx).Model(&model.Product{})
	if query.Status != nil {
		database = database.Where("status = ?", *query.Status)
	}
	if query.Cursor != "" {
		database = database.Where("(created_at, id) < (?, ?)", cursor.CreatedAt, cursor.ID)
	}

	var products []model.Product
	if err := database.
		Order("created_at DESC, id DESC").
		Limit(query.Limit + 1).
		Find(&products).Error; err != nil {
		return ProductPageResponse{}, fmt.Errorf("list admin products: %w", err)
	}

	page := ProductPageResponse{Items: make([]ProductResponse, 0, min(len(products), query.Limit))}
	if len(products) > query.Limit {
		products = products[:query.Limit]
		last := products[len(products)-1]
		nextCursor, err := encodeProductCursor(last.CreatedAt, last.ID, filter)
		if err != nil {
			return ProductPageResponse{}, fmt.Errorf("encode admin product cursor: %w", err)
		}
		page.NextCursor = &nextCursor
	}
	if len(products) == 0 {
		return page, nil
	}

	productIDs := make([]string, 0, len(products))
	for _, product := range products {
		productIDs = append(productIDs, product.ID)
	}
	var skus []model.SKU
	if err := service.database.WithContext(ctx).
		Where("product_id IN ?", productIDs).
		Order("product_id ASC, id ASC").
		Find(&skus).Error; err != nil {
		return ProductPageResponse{}, fmt.Errorf("list admin product SKUs: %w", err)
	}
	skusByProduct := make(map[string][]model.SKU, len(products))
	for _, sku := range skus {
		skusByProduct[sku.ProductID] = append(skusByProduct[sku.ProductID], sku)
	}
	for _, product := range products {
		page.Items = append(page.Items, newProductResponse(product, skusByProduct[product.ID]))
	}
	return page, nil
}

func (service *Service) CreateProduct(
	ctx context.Context,
	actorUserID string,
	request CreateProductRequest,
) (ProductResponse, error) {
	product := model.Product{
		CategoryID:  request.CategoryID,
		Name:        request.Name,
		Description: request.Description,
		ImageURL:    request.ImageURL,
		Status:      model.ProductStatusActive,
	}
	skus := make([]model.SKU, 0, len(request.SKUs))
	err := service.database.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if err := requireCategory(transaction, request.CategoryID); err != nil {
			return err
		}
		if err := transaction.Create(&product).Error; err != nil {
			return fmt.Errorf("insert product: %w", err)
		}
		for _, input := range request.SKUs {
			sku, err := createSKUWithInventory(transaction, actorUserID, product.ID, input)
			if err != nil {
				return err
			}
			skus = append(skus, sku)
		}
		return nil
	})
	if err != nil {
		return ProductResponse{}, err
	}
	return newProductResponse(product, skus), nil
}

func (service *Service) CreateSKU(
	ctx context.Context,
	actorUserID, productID string,
	request CreateSKUInput,
) (SKUResponse, error) {
	var sku model.SKU
	err := service.database.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		var product model.Product
		if err := transaction.Where("id = ?", productID).Take(&product).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return ErrProductNotFound
			}
			return fmt.Errorf("find SKU parent product: %w", err)
		}
		created, err := createSKUWithInventory(transaction, actorUserID, productID, request)
		if err != nil {
			return err
		}
		sku = created
		return nil
	})
	if err != nil {
		return SKUResponse{}, err
	}
	return newSKUResponse(sku), nil
}

func (service *Service) UpdateProduct(
	ctx context.Context,
	productID string,
	request UpdateProductRequest,
) (ProductResponse, error) {
	err := service.database.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		updates := make(map[string]any)
		if request.CategoryID.Present {
			if err := requireCategory(transaction, request.CategoryID.Value); err != nil {
				return err
			}
			if request.CategoryID.Value == nil {
				updates["category_id"] = nil
			} else {
				updates["category_id"] = *request.CategoryID.Value
			}
		}
		if request.Name.Present {
			updates["name"] = *request.Name.Value
		}
		if request.Description.Present {
			updates["description"] = *request.Description.Value
		}
		if request.ImageURL.Present {
			if request.ImageURL.Value == nil {
				updates["image_url"] = nil
			} else {
				updates["image_url"] = *request.ImageURL.Value
			}
		}
		if request.Active.Present {
			if request.Active.Value {
				updates["status"] = model.ProductStatusActive
			} else {
				updates["status"] = model.ProductStatusInactive
			}
		}

		result := transaction.Model(&model.Product{}).Where("id = ?", productID).Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("update product: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return ErrProductNotFound
		}
		return nil
	})
	if err != nil {
		return ProductResponse{}, err
	}
	return service.GetAdminProduct(ctx, productID)
}

func (service *Service) UpdateSKU(
	ctx context.Context,
	skuID string,
	request UpdateSKURequest,
) (SKUResponse, error) {
	updates := make(map[string]any)
	if request.RegularPrice.Present {
		updates["regular_price"] = request.RegularPrice.Value
	}
	if request.MaxQuantityPerOrder.Present {
		updates["max_quantity_per_order"] = request.MaxQuantityPerOrder.Value
	}
	if request.Active.Present {
		if request.Active.Value {
			updates["status"] = model.SKUStatusActive
		} else {
			updates["status"] = model.SKUStatusInactive
		}
	}

	var sku model.SKU
	err := service.database.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		result := transaction.Model(&model.SKU{}).Where("id = ?", skuID).Updates(updates)
		if result.Error != nil {
			return fmt.Errorf("update SKU: %w", result.Error)
		}
		if result.RowsAffected != 1 {
			return ErrSKUNotFound
		}
		if err := transaction.Where("id = ?", skuID).Take(&sku).Error; err != nil {
			return fmt.Errorf("reload updated SKU: %w", err)
		}
		return nil
	})
	if err != nil {
		return SKUResponse{}, err
	}
	return newSKUResponse(sku), nil
}

func requireCategory(transaction *gorm.DB, categoryID *string) error {
	if categoryID == nil {
		return nil
	}
	var category model.Category
	if err := transaction.Where("id = ?", *categoryID).Take(&category).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return ErrCategoryNotFound
		}
		return fmt.Errorf("find product category: %w", err)
	}
	return nil
}

func createSKUWithInventory(
	transaction *gorm.DB,
	actorUserID, productID string,
	request CreateSKUInput,
) (model.SKU, error) {
	sku := model.SKU{
		ProductID:           productID,
		Code:                request.Code,
		Attributes:          request.Attributes,
		RegularPrice:        request.RegularPrice,
		MaxQuantityPerOrder: request.MaxQuantityPerOrder,
		Status:              model.SKUStatusActive,
	}
	if err := transaction.Create(&sku).Error; err != nil {
		if errors.Is(err, gorm.ErrDuplicatedKey) {
			return model.SKU{}, ErrSKUCodeTaken
		}
		return model.SKU{}, fmt.Errorf("insert SKU: %w", err)
	}
	if err := transaction.Create(&model.Inventory{SKUID: sku.ID, OnHand: request.InitialOnHand}).Error; err != nil {
		return model.SKU{}, fmt.Errorf("insert initial inventory: %w", err)
	}
	if request.InitialOnHand == 0 {
		return sku, nil
	}
	if err := transaction.Create(&model.InventoryAdjustment{
		SKUID:        sku.ID,
		ActorUserID:  actorUserID,
		Delta:        request.InitialOnHand,
		OnHandBefore: 0,
		OnHandAfter:  request.InitialOnHand,
		Reason:       model.InventoryAdjustmentReasonInitialStock,
	}).Error; err != nil {
		return model.SKU{}, fmt.Errorf("insert initial inventory adjustment: %w", err)
	}
	return sku, nil
}

func escapeLikeLiteral(value string) string {
	return strings.NewReplacer(
		`\`, `\\`,
		`%`, `\%`,
		`_`, `\_`,
	).Replace(value)
}
