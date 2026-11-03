package bootstrap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"

	"cloudian/cloudian-scale/internal/model"
	"cloudian/cloudian-scale/internal/module/auth"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"uuid"
)

type SeedCategory struct {
	ID        string
	Name      string
	Slug      string
	IsVisible bool
}

type SeedSKU struct {
	Code                string
	Attributes          map[string]any
	RegularPrice        int64
	MaxQuantityPerOrder int
	Status              string
	OnHand              int64
	Reserved            int64
}

type SeedProduct struct {
	Name         string
	CategorySlug *string
	Description  string
	ImageURL     *string
	Status       string
	SKUs         []SeedSKU
}

func boolPtr(b bool) *bool {
	return &b
}

func strPtr(s string) *string {
	return &s
}

func CanSeedCatalog() error {
	environment := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	if environment == "development" || environment == "dev" || environment == "test" {
		return nil
	}
	if strings.EqualFold(strings.TrimSpace(os.Getenv("ALLOW_CATALOG_SEED")), "true") {
		return nil
	}
	return errors.New("catalog seed is restricted to development or test, or set ALLOW_CATALOG_SEED=true")
}

func SeedCatalog(ctx context.Context, db *gorm.DB) error {
	if db == nil {
		return errors.New("database connection is required")
	}
	if err := CanSeedCatalog(); err != nil {
		return err
	}

	return db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Ensure actor user exists for inventory ledger
		actorUserID, err := ensureActorUser(tx)
		if err != nil {
			return fmt.Errorf("ensure actor user: %w", err)
		}

		// 2. Seed Categories
		categoryMap, err := seedCategories(tx)
		if err != nil {
			return fmt.Errorf("seed categories: %w", err)
		}

		// 3. Seed Products, SKUs, Inventories & Adjustments
		if err := seedProductsAndInventory(tx, categoryMap, actorUserID); err != nil {
			return fmt.Errorf("seed products and inventory: %w", err)
		}

		return nil
	})
}

func ensureActorUser(tx *gorm.DB) (string, error) {
	var user model.User
	// Try to find an existing admin or active user
	if err := tx.Where("role = ? AND status = ?", model.RoleAdmin, "active").Order("created_at ASC").First(&user).Error; err == nil {
		return user.ID, nil
	}
	if err := tx.Where("status = ?", "active").Order("created_at ASC").First(&user).Error; err == nil {
		return user.ID, nil
	}

	// Create a default system seed admin
	passwordHash, err := bcrypt.GenerateFromPassword([]byte("Admin@12345678"), auth.DefaultPasswordCost)
	if err != nil {
		return "", fmt.Errorf("hash admin password: %w", err)
	}

	seedAdmin := model.User{
		ID:              uuid.NewV4().String(),
		Email:           "admin@salescale.internal",
		EmailNormalized: "admin@salescale.internal",
		PasswordHash:    string(passwordHash),
		Role:            model.RoleAdmin,
		Status:          "active",
	}

	if err := tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "email_normalized"}},
		DoNothing: true,
	}).Create(&seedAdmin).Error; err != nil {
		return "", err
	}

	if err := tx.Where("email_normalized = ?", "admin@salescale.internal").First(&user).Error; err != nil {
		return "", err
	}
	return user.ID, nil
}

func seedCategories(tx *gorm.DB) (map[string]string, error) {
	categories := []SeedCategory{
		{
			ID:        "61000000-0000-4000-8000-000000000001",
			Name:      "Điện thoại & Phụ kiện",
			Slug:      "dien-thoai-phu-kien",
			IsVisible: true,
		},
		{
			ID:        "61000000-0000-4000-8000-000000000002",
			Name:      "Thời trang & May mặc",
			Slug:      "thoi-trang-may-mac",
			IsVisible: true,
		},
		{
			ID:        "61000000-0000-4000-8000-000000000003",
			Name:      "Đồ gia dụng & Đời sống",
			Slug:      "do-gia-dung-doi-song",
			IsVisible: true,
		},
		{
			ID:        "61000000-0000-4000-8000-000000000004",
			Name:      "Danh mục nội bộ thử nghiệm (Ẩn)",
			Slug:      "danh-muc-an-thu-nghiem",
			IsVisible: false,
		},
	}

	categoryMap := make(map[string]string)
	for _, c := range categories {
		category := model.Category{
			ID:        c.ID,
			Name:      c.Name,
			Slug:      c.Slug,
			IsVisible: boolPtr(c.IsVisible),
		}
		if err := tx.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "slug"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "is_visible"}),
		}).Create(&category).Error; err != nil {
			return nil, fmt.Errorf("upsert category %s: %w", c.Slug, err)
		}

		var saved model.Category
		if err := tx.Where("slug = ?", c.Slug).First(&saved).Error; err != nil {
			return nil, err
		}
		categoryMap[c.Slug] = saved.ID
	}
	return categoryMap, nil
}

func seedProductsAndInventory(tx *gorm.DB, categoryMap map[string]string, actorUserID string) error {
	products := []SeedProduct{
		{
			Name:         "iPhone 15 Pro Max 256GB",
			CategorySlug: strPtr("dien-thoai-phu-kien"),
			Description:  "Flagship cao cấp với khung viền Titan siêu bền nhẹ, chip Apple A17 Pro và camera telephoto 5x đỉnh cao.",
			ImageURL:     strPtr("https://images.unsplash.com/photo-1695048133142-1a20484d2569"),
			Status:       model.ProductStatusActive,
			SKUs: []SeedSKU{
				{
					Code:                "IP15PM-256-NAT",
					Attributes:          map[string]any{"color": "Titan Tự Nhiên", "storage": "256GB"},
					RegularPrice:        29990000,
					MaxQuantityPerOrder: 2,
					Status:              model.SKUStatusActive,
					OnHand:              50,
					Reserved:            2,
				},
				{
					Code:                "IP15PM-256-BLU",
					Attributes:          map[string]any{"color": "Titan Xanh", "storage": "256GB"},
					RegularPrice:        29990000,
					MaxQuantityPerOrder: 2,
					Status:              model.SKUStatusActive,
					OnHand:              30,
					Reserved:            0,
				},
				{
					Code:                "IP15PM-512-NAT",
					Attributes:          map[string]any{"color": "Titan Tự Nhiên", "storage": "512GB"},
					RegularPrice:        34990000,
					MaxQuantityPerOrder: 1,
					Status:              model.SKUStatusActive,
					OnHand:              15,
					Reserved:            1,
				},
			},
		},
		{
			Name:         "Samsung Galaxy S24 Ultra",
			CategorySlug: strPtr("dien-thoai-phu-kien"),
			Description:  "Quyền năng sáng tạo với Galaxy AI, camera zoom 100x sắc nét, vi xử lý Snapdragon 8 Gen 3 for Galaxy và bút S-Pen tích hợp.",
			ImageURL:     strPtr("https://images.unsplash.com/photo-1610945265064-0e34e5519bbf"),
			Status:       model.ProductStatusActive,
			SKUs: []SeedSKU{
				{
					Code:                "S24U-256-GRY",
					Attributes:          map[string]any{"color": "Xám Titan", "storage": "256GB"},
					RegularPrice:        27990000,
					MaxQuantityPerOrder: 3,
					Status:              model.SKUStatusActive,
					OnHand:              40,
					Reserved:            0,
				},
				{
					Code:                "S24U-512-BLK",
					Attributes:          map[string]any{"color": "Đen Titan", "storage": "512GB"},
					RegularPrice:        31990000,
					MaxQuantityPerOrder: 2,
					Status:              model.SKUStatusActive,
					OnHand:              20,
					Reserved:            0,
				},
			},
		},
		{
			Name:         "Củ Sạc Nhanh Anker 65W GaNPrime",
			CategorySlug: strPtr("dien-thoai-phu-kien"),
			Description:  "Củ sạc công nghệ GaN siêu nhỏ gọn, hỗ trợ 3 cổng ra sạc cùng lúc laptop, máy tính bảng và điện thoại.",
			ImageURL:     strPtr("https://images.unsplash.com/photo-1583863788434-e58a36330cf0"),
			Status:       model.ProductStatusActive,
			SKUs: []SeedSKU{
				{
					Code:                "ANKER-65W-BLK",
					Attributes:          map[string]any{"color": "Đen"},
					RegularPrice:        750000,
					MaxQuantityPerOrder: 5,
					Status:              model.SKUStatusActive,
					OnHand:              100,
					Reserved:            10,
				},
				{
					Code:                "ANKER-65W-WHT",
					Attributes:          map[string]any{"color": "Trắng"},
					RegularPrice:        750000,
					MaxQuantityPerOrder: 5,
					Status:              model.SKUStatusActive,
					OnHand:              80,
					Reserved:            0,
				},
			},
		},
		{
			Name:         "Áo Thun Trơn Nam Cotton 100% Co Giãn",
			CategorySlug: strPtr("thoi-trang-may-mac"),
			Description:  "Chất liệu sợi bông tự nhiên thoáng mát, co giãn 4 chiều mềm mại, không xù lông sau khi giặt máy.",
			ImageURL:     strPtr("https://images.unsplash.com/photo-1521572267360-ee0c2909d518"),
			Status:       model.ProductStatusActive,
			SKUs: []SeedSKU{
				{
					Code:                "TSHIRT-BLK-M",
					Attributes:          map[string]any{"color": "Đen", "size": "M"},
					RegularPrice:        199000,
					MaxQuantityPerOrder: 10,
					Status:              model.SKUStatusActive,
					OnHand:              200,
					Reserved:            5,
				},
				{
					Code:                "TSHIRT-BLK-L",
					Attributes:          map[string]any{"color": "Đen", "size": "L"},
					RegularPrice:        199000,
					MaxQuantityPerOrder: 10,
					Status:              model.SKUStatusActive,
					OnHand:              150,
					Reserved:            0,
				},
				{
					Code:                "TSHIRT-WHT-M",
					Attributes:          map[string]any{"color": "Trắng", "size": "M"},
					RegularPrice:        199000,
					MaxQuantityPerOrder: 10,
					Status:              model.SKUStatusActive,
					OnHand:              120,
					Reserved:            0,
				},
				{
					Code:                "TSHIRT-WHT-L",
					Attributes:          map[string]any{"color": "Trắng", "size": "L"},
					RegularPrice:        199000,
					MaxQuantityPerOrder: 10,
					Status:              model.SKUStatusActive,
					OnHand:              100,
					Reserved:            0,
				},
			},
		},
		{
			Name:         "Quần Jeans Nam Slimfit Cao Cấp",
			CategorySlug: strPtr("thoi-trang-may-mac"),
			Description:  "Thiết kế ôm chân thanh lịch, chất vải denim dày dặn chống phai màu, phù hợp đi làm và đi chơi.",
			ImageURL:     strPtr("https://images.unsplash.com/photo-1542272604-780c96856592"),
			Status:       model.ProductStatusActive,
			SKUs: []SeedSKU{
				{
					Code:                "JEAN-SLIM-30",
					Attributes:          map[string]any{"size": "30", "wash": "Dark Indigo"},
					RegularPrice:        499000,
					MaxQuantityPerOrder: 5,
					Status:              model.SKUStatusActive,
					OnHand:              60,
					Reserved:            0,
				},
				{
					Code:                "JEAN-SLIM-31",
					Attributes:          map[string]any{"size": "31", "wash": "Dark Indigo"},
					RegularPrice:        499000,
					MaxQuantityPerOrder: 5,
					Status:              model.SKUStatusActive,
					OnHand:              70,
					Reserved:            0,
				},
				{
					Code:                "JEAN-SLIM-32",
					Attributes:          map[string]any{"size": "32", "wash": "Dark Indigo"},
					RegularPrice:        499000,
					MaxQuantityPerOrder: 5,
					Status:              model.SKUStatusActive,
					OnHand:              50,
					Reserved:            0,
				},
			},
		},
		{
			Name:         "Nồi Chiên Không Dầu Điện Tử 6L RapidAir",
			CategorySlug: strPtr("do-gia-dung-doi-song"),
			Description:  "Dung tích lớn 6L nướng nguyên con gà, công nghệ đối lưu nhiệt giảm 90% lượng dầu mỡ thừa.",
			ImageURL:     strPtr("https://images.unsplash.com/photo-1584269600464-37b1b58a9fe7"),
			Status:       model.ProductStatusActive,
			SKUs: []SeedSKU{
				{
					Code:                "AF-6L-DIGITAL",
					Attributes:          map[string]any{"capacity": "6L", "color": "Đen Matte"},
					RegularPrice:        1890000,
					MaxQuantityPerOrder: 2,
					Status:              model.SKUStatusActive,
					OnHand:              35,
					Reserved:            3,
				},
			},
		},
		{
			Name:         "Máy Hút Bụi Cầm Tay Không Dây 20000Pa",
			CategorySlug: strPtr("do-gia-dung-doi-song"),
			Description:  "Lực hút mạnh mẽ với động cơ không chổi than, màng lọc bụi mịn HEPA và thời lượng pin tới 45 phút.",
			ImageURL:     strPtr("https://images.unsplash.com/photo-1558317374-067fb5f30001"),
			Status:       model.ProductStatusActive,
			SKUs: []SeedSKU{
				{
					Code:                "VAC-WIRELESS-V1",
					Attributes:          map[string]any{"power": "150W", "battery": "2500mAh"},
					RegularPrice:        1450000,
					MaxQuantityPerOrder: 2,
					Status:              model.SKUStatusActive,
					OnHand:              25,
					Reserved:            0,
				},
			},
		},
		// --- Test case products:
		{
			Name:         "Sản phẩm Flash Sale Cháy Hàng (Test Out Of Stock)",
			CategorySlug: strPtr("do-gia-dung-doi-song"),
			Description:  "Dành riêng cho test case hàng có số lượng tồn bằng 0 (out of stock).",
			ImageURL:     nil,
			Status:       model.ProductStatusActive,
			SKUs: []SeedSKU{
				{
					Code:                "FLASH-OUT-OF-STOCK",
					Attributes:          map[string]any{"special": "Zero Stock"},
					RegularPrice:        99000,
					MaxQuantityPerOrder: 1,
					Status:              model.SKUStatusActive,
					OnHand:              0,
					Reserved:            0,
				},
			},
		},
		{
			Name:         "Sản phẩm Ngừng Bán (Test Inactive Product)",
			CategorySlug: strPtr("do-gia-dung-doi-song"),
			Description:  "Sản phẩm có status inactive, khách vãng lai không xem được trên /v1/products.",
			ImageURL:     nil,
			Status:       model.ProductStatusInactive,
			SKUs: []SeedSKU{
				{
					Code:                "HIDDEN-PROD-SKU",
					Attributes:          map[string]any{"status": "archived"},
					RegularPrice:        500000,
					MaxQuantityPerOrder: 1,
					Status:              model.SKUStatusActive,
					OnHand:              10,
					Reserved:            0,
				},
			},
		},
		{
			Name:         "Sản phẩm Thuộc Danh Mục Ẩn (Test Hidden Category)",
			CategorySlug: strPtr("danh-muc-an-thu-nghiem"),
			Description:  "Sản phẩm active nhưng thuộc category is_visible=false, public catalog sẽ không trả về.",
			ImageURL:     nil,
			Status:       model.ProductStatusActive,
			SKUs: []SeedSKU{
				{
					Code:                "PROD-HIDDEN-CAT-SKU",
					Attributes:          map[string]any{"category_hidden": true},
					RegularPrice:        100000,
					MaxQuantityPerOrder: 1,
					Status:              model.SKUStatusActive,
					OnHand:              20,
					Reserved:            0,
				},
			},
		},
		{
			Name:         "Voucher Mua Sắm SaleScale 50K (Test No Category)",
			CategorySlug: nil,
			Description:  "Sản phẩm không thuộc category nào (category_id = NULL).",
			ImageURL:     nil,
			Status:       model.ProductStatusActive,
			SKUs: []SeedSKU{
				{
					Code:                "VOUCHER-50K",
					Attributes:          map[string]any{"discount": "50000 VND"},
					RegularPrice:        50000,
					MaxQuantityPerOrder: 5,
					Status:              model.SKUStatusActive,
					OnHand:              500,
					Reserved:            0,
				},
			},
		},
	}

	for _, p := range products {
		var categoryID *string
		if p.CategorySlug != nil {
			id := categoryMap[*p.CategorySlug]
			categoryID = &id
		}

		var product model.Product
		err := tx.Where("name = ?", p.Name).First(&product).Error
		if errors.Is(err, gorm.ErrRecordNotFound) {
			product = model.Product{
				ID:          uuid.NewV4().String(),
				CategoryID:  categoryID,
				Name:        p.Name,
				Description: p.Description,
				ImageURL:    p.ImageURL,
				Status:      p.Status,
			}
			if err := tx.Create(&product).Error; err != nil {
				return fmt.Errorf("create product %s: %w", p.Name, err)
			}
		} else if err != nil {
			return err
		} else {
			// Update existing product fields
			product.CategoryID = categoryID
			product.Description = p.Description
			product.ImageURL = p.ImageURL
			product.Status = p.Status
			if err := tx.Save(&product).Error; err != nil {
				return fmt.Errorf("update product %s: %w", p.Name, err)
			}
		}

		// Seed SKUs for this product
		for _, s := range p.SKUs {
			attrJSON, err := json.Marshal(s.Attributes)
			if err != nil {
				return fmt.Errorf("marshal sku attributes for %s: %w", s.Code, err)
			}

			var sku model.SKU
			err = tx.Where("code = ?", s.Code).First(&sku).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				sku = model.SKU{
					ID:                  uuid.NewV4().String(),
					ProductID:           product.ID,
					Code:                s.Code,
					Attributes:          attrJSON,
					RegularPrice:        s.RegularPrice,
					MaxQuantityPerOrder: s.MaxQuantityPerOrder,
					Status:              s.Status,
				}
				if err := tx.Create(&sku).Error; err != nil {
					return fmt.Errorf("create sku %s: %w", s.Code, err)
				}
			} else if err != nil {
				return err
			} else {
				sku.ProductID = product.ID
				sku.Attributes = attrJSON
				sku.RegularPrice = s.RegularPrice
				sku.MaxQuantityPerOrder = s.MaxQuantityPerOrder
				sku.Status = s.Status
				if err := tx.Save(&sku).Error; err != nil {
					return fmt.Errorf("update sku %s: %w", s.Code, err)
				}
			}

			// Seed Inventory for this SKU
			var inventory model.Inventory
			err = tx.Where("sku_id = ?", sku.ID).First(&inventory).Error
			if errors.Is(err, gorm.ErrRecordNotFound) {
				inventory = model.Inventory{
					SKUID:    sku.ID,
					OnHand:   s.OnHand,
					Reserved: s.Reserved,
					Version:  0,
				}
				if err := tx.Create(&inventory).Error; err != nil {
					return fmt.Errorf("create inventory for sku %s: %w", s.Code, err)
				}

				// Create initial inventory adjustment record if on_hand > 0
				if s.OnHand > 0 {
					adj := model.InventoryAdjustment{
						ID:           uuid.NewV4().String(),
						SKUID:        sku.ID,
						ActorUserID:  actorUserID,
						Delta:        s.OnHand,
						OnHandBefore: 0,
						OnHandAfter:  s.OnHand,
						Reason:       model.InventoryAdjustmentReasonInitialStock,
					}
					if err := tx.Create(&adj).Error; err != nil {
						return fmt.Errorf("create initial adjustment for sku %s: %w", s.Code, err)
					}
				}
			} else if err != nil {
				return err
			} else {
				// Keep reserved valid (must not exceed on_hand)
				if s.Reserved > s.OnHand {
					s.Reserved = s.OnHand
				}
				inventory.OnHand = s.OnHand
				inventory.Reserved = s.Reserved
				if err := tx.Save(&inventory).Error; err != nil {
					return fmt.Errorf("update inventory for sku %s: %w", s.Code, err)
				}
			}
		}
	}

	return nil
}
