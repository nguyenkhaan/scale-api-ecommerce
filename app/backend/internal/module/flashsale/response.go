package flashsale

import (
	"time"

	"gorm.io/gorm"
)

// itemRow is one campaign item joined with the SKU and product data shown in responses.
type itemRow struct {
	ID            string `gorm:"column:id"`
	CampaignID    string `gorm:"column:campaign_id"`
	SKUID         string `gorm:"column:sku_id"`
	SKUCode       string `gorm:"column:sku_code"`
	ProductName   string `gorm:"column:product_name"`
	RegularPrice  int64  `gorm:"column:regular_price"`
	SalePrice     int64  `gorm:"column:sale_price"`
	QuotaTotal    int64  `gorm:"column:quota_total"`
	PurchaseLimit int    `gorm:"column:purchase_limit"`
}

// loadItemsByCampaign loads the items of all given campaigns with one joined query.
func loadItemsByCampaign(db *gorm.DB, campaignIDs []string) (map[string][]itemRow, error) {
	grouped := make(map[string][]itemRow, len(campaignIDs))
	if len(campaignIDs) == 0 {
		return grouped, nil
	}
	var rows []itemRow
	err := db.Raw(`
		SELECT i.id, i.campaign_id, i.sku_id, s.code AS sku_code, p.name AS product_name,
		       s.regular_price, i.sale_price, i.quota_total, i.purchase_limit
		FROM flash_sale_items i
		JOIN skus s ON s.id = i.sku_id
		JOIN products p ON p.id = s.product_id
		WHERE i.campaign_id IN ?
		ORDER BY i.campaign_id, s.code, i.id`, campaignIDs).Scan(&rows).Error
	if err != nil {
		return nil, err
	}
	for _, row := range rows {
		grouped[row.CampaignID] = append(grouped[row.CampaignID], row)
	}
	return grouped, nil
}

// buildResponses renders campaigns with their items; status comes from each row's database clock reading.
func buildResponses(db *gorm.DB, campaigns []campaignRow) ([]FlashSaleResponse, error) {
	ids := make([]string, len(campaigns))
	for i, campaign := range campaigns {
		ids[i] = campaign.ID
	}
	itemsByCampaign, err := loadItemsByCampaign(db, ids)
	if err != nil {
		return nil, err
	}
	responses := make([]FlashSaleResponse, len(campaigns))
	for i, campaign := range campaigns {
		rows := itemsByCampaign[campaign.ID]
		items := make([]FlashSaleItemResponse, len(rows))
		for j, row := range rows {
			items[j] = FlashSaleItemResponse{
				ID:            row.ID,
				SKUID:         row.SKUID,
				SKUCode:       row.SKUCode,
				ProductName:   row.ProductName,
				RegularPrice:  row.RegularPrice,
				SalePrice:     row.SalePrice,
				QuotaTotal:    row.QuotaTotal,
				PurchaseLimit: row.PurchaseLimit,
			}
		}
		responses[i] = FlashSaleResponse{
			ID:       campaign.ID,
			Name:     campaign.Name,
			Status:   campaign.effectiveStatus(),
			StartsAt: campaign.StartsAt.UTC().Format(time.RFC3339),
			EndsAt:   campaign.EndsAt.UTC().Format(time.RFC3339),
			Items:    items,
		}
	}
	return responses, nil
}

func buildResponse(db *gorm.DB, campaign campaignRow) (FlashSaleResponse, error) {
	responses, err := buildResponses(db, []campaignRow{campaign})
	if err != nil {
		return FlashSaleResponse{}, err
	}
	return responses[0], nil
}
