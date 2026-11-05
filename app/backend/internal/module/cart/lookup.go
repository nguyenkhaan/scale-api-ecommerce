package cart

import (
	"time"

	"cloudian/cloudian-scale/internal/model"
	"gorm.io/gorm"
)

// skuInfo is the purchasable state of a SKU: its own and its product/category availability plus stock.
type skuInfo struct {
	ID                  string `gorm:"column:id"`
	Code                string `gorm:"column:code"`
	ProductName         string `gorm:"column:product_name"`
	RegularPrice        int64  `gorm:"column:regular_price"`
	MaxQuantityPerOrder int    `gorm:"column:max_quantity_per_order"`
	Available           bool   `gorm:"column:available"`
	OnHand              int64  `gorm:"column:on_hand"`
	Reserved            int64  `gorm:"column:reserved"`
}

// flashInfo is a flash sale item together with its campaign window.
type flashInfo struct {
	ID            string    `gorm:"column:id"`
	SKUID         string    `gorm:"column:sku_id"`
	SalePrice     int64     `gorm:"column:sale_price"`
	QuotaTotal    int64     `gorm:"column:quota_total"`
	QuotaReserved int64     `gorm:"column:quota_reserved"`
	QuotaSold     int64     `gorm:"column:quota_sold"`
	PurchaseLimit int64     `gorm:"column:purchase_limit"`
	Status        string    `gorm:"column:status"`
	StartsAt      time.Time `gorm:"column:starts_at"`
	EndsAt        time.Time `gorm:"column:ends_at"`
}

// acceptingOrders reports whether the campaign takes orders at now.
func (f flashInfo) acceptingOrders(now time.Time) bool {
	return f.Status == model.FlashSaleStatusScheduled && !now.Before(f.StartsAt) && now.Before(f.EndsAt)
}

// notEnded reports whether the campaign is scheduled and still running or upcoming at now.
func (f flashInfo) notEnded(now time.Time) bool {
	return f.Status == model.FlashSaleStatusScheduled && now.Before(f.EndsAt)
}

// dbNow returns the database clock so every time-based decision uses one authority.
func dbNow(db *gorm.DB) (time.Time, error) {
	var now time.Time
	err := db.Raw("SELECT now()").Row().Scan(&now)
	return now, err
}

func loadSKUInfos(db *gorm.DB, skuIDs []string) (map[string]skuInfo, error) {
	infos := make(map[string]skuInfo, len(skuIDs))
	if len(skuIDs) == 0 {
		return infos, nil
	}
	var rows []skuInfo
	err := db.Raw(`
		SELECT s.id, s.code, p.name AS product_name, s.regular_price, s.max_quantity_per_order,
		       (s.status = 'active' AND p.status = 'active' AND (c.id IS NULL OR c.is_visible)) AS available,
		       COALESCE(i.on_hand, 0) AS on_hand, COALESCE(i.reserved, 0) AS reserved
		FROM skus s
		JOIN products p ON p.id = s.product_id
		LEFT JOIN categories c ON c.id = p.category_id
		LEFT JOIN inventories i ON i.sku_id = s.id
		WHERE s.id IN ?`, skuIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		infos[row.ID] = row
	}
	return infos, nil
}

func loadFlashInfos(db *gorm.DB, itemIDs []string) (map[string]flashInfo, error) {
	infos := make(map[string]flashInfo, len(itemIDs))
	if len(itemIDs) == 0 {
		return infos, nil
	}
	var rows []flashInfo
	err := db.Raw(`
		SELECT fi.id, fi.sku_id, fi.sale_price, fi.quota_total, fi.quota_reserved, fi.quota_sold,
		       fi.purchase_limit, fc.status, fc.starts_at, fc.ends_at
		FROM flash_sale_items fi
		JOIN flash_sale_campaigns fc ON fc.id = fi.campaign_id
		WHERE fi.id IN ?`, itemIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		infos[row.ID] = row
	}
	return infos, nil
}

// loadUsedQuantities returns reserved+sold units per flash sale item for one customer.
func loadUsedQuantities(db *gorm.DB, userID string, itemIDs []string) (map[string]int64, error) {
	used := make(map[string]int64, len(itemIDs))
	if len(itemIDs) == 0 {
		return used, nil
	}
	var rows []struct {
		FlashSaleItemID string `gorm:"column:flash_sale_item_id"`
		Used            int64  `gorm:"column:used"`
	}
	err := db.Raw(`
		SELECT flash_sale_item_id, reserved_quantity + sold_quantity AS used
		FROM flash_sale_customer_usage
		WHERE user_id = ? AND flash_sale_item_id IN ?`, userID, itemIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		used[row.FlashSaleItemID] = row.Used
	}
	return used, nil
}
