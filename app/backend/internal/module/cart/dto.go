package cart

// UpsertCartItemRequest sets the absolute quantity of one SKU line.
type UpsertCartItemRequest struct {
	Quantity        int     `json:"quantity"`
	PriceSource     string  `json:"price_source"`
	FlashSaleItemID *string `json:"flash_sale_item_id"`
	ExpectedVersion *int64  `json:"expected_version"`
}

// CartItemResponse is one cart line; the gorm tags let the cart view query scan straight into it.
type CartItemResponse struct {
	ID                 string  `json:"id" gorm:"column:id"`
	SKUID              string  `json:"sku_id" gorm:"column:sku_id"`
	Quantity           int     `json:"quantity" gorm:"column:quantity"`
	Version            int64   `json:"version" gorm:"column:version"`
	PriceSource        string  `json:"price_source" gorm:"column:price_source"`
	FlashSaleItemID    *string `json:"flash_sale_item_id" gorm:"column:flash_sale_item_id"`
	ProductName        string  `json:"product_name" gorm:"column:product_name"`
	SKUCode            string  `json:"sku_code" gorm:"column:sku_code"`
	ReferenceUnitPrice int64   `json:"reference_unit_price" gorm:"column:reference_unit_price"`
}

type CartResponse struct {
	ID    string             `json:"id"`
	Items []CartItemResponse `json:"items"`
}

type SelectedCartItem struct {
	CartItemID string `json:"cart_item_id"`
	Version    int64  `json:"version"`
}

type CheckoutSelection struct {
	Items []SelectedCartItem `json:"items"`
}

type CheckoutLineResponse struct {
	CartItemID      string  `json:"cart_item_id"`
	CartItemVersion int64   `json:"cart_item_version"`
	SKUID           string  `json:"sku_id"`
	Quantity        int     `json:"quantity"`
	UnitPrice       int64   `json:"unit_price"`
	Subtotal        int64   `json:"subtotal"`
	PriceSource     string  `json:"price_source"`
	FlashSaleItemID *string `json:"flash_sale_item_id"`
}

type CheckoutPreviewResponse struct {
	Currency    string                 `json:"currency"`
	TotalAmount int64                  `json:"total_amount"`
	Lines       []CheckoutLineResponse `json:"lines"`
	PreviewedAt string                 `json:"previewed_at"`
}
