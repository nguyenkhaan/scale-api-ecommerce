package order

import (
	"encoding/json"
	"time"
)

// RecipientResponse is the delivery snapshot stored with the order.
type RecipientResponse struct {
	Name    string `json:"name"`
	Phone   string `json:"phone"`
	Address string `json:"address"`
}

// OrderItemResponse is one immutable line snapshot of an order.
type OrderItemResponse struct {
	ID            string          `json:"id"`
	SKUID         string          `json:"sku_id"`
	SKUCode       string          `json:"sku_code"`
	ProductName   string          `json:"product_name"`
	SKUAttributes json.RawMessage `json:"sku_attributes"`
	Quantity      int             `json:"quantity"`
	UnitPrice     int64           `json:"unit_price"`
	Subtotal      int64           `json:"subtotal"`
	PriceSource   string          `json:"price_source"`
}

// OrderResponse is the customer view of an order. Payable, PaymentStatus and ReconciliationRequired are
// computed by the database at read time and are never stored.
type OrderResponse struct {
	ID                     string              `json:"id"`
	OrderNumber            string              `json:"order_number"`
	Status                 string              `json:"status"`
	Currency               string              `json:"currency"`
	TotalAmount            int64               `json:"total_amount"`
	Recipient              RecipientResponse   `json:"recipient"`
	Items                  []OrderItemResponse `json:"items"`
	CreatedAt              time.Time           `json:"created_at"`
	ExpiresAt              time.Time           `json:"expires_at"`
	Payable                bool                `json:"payable"`
	PaymentStatus          *string             `json:"payment_status"`
	ReconciliationRequired bool                `json:"reconciliation_required"`
}

// OrderPage is one cursor page of a customer's orders, newest first.
type OrderPage struct {
	Items      []OrderResponse `json:"items"`
	NextCursor *string         `json:"next_cursor"`
}

// orderRow and itemRow are the raw read-model rows scanned from PostgreSQL.
type orderRow struct {
	ID                     string    `gorm:"column:id"`
	OrderNumber            string    `gorm:"column:order_number"`
	Status                 string    `gorm:"column:status"`
	Currency               string    `gorm:"column:currency"`
	TotalAmount            int64     `gorm:"column:total_amount"`
	RecipientName          string    `gorm:"column:recipient_name"`
	RecipientPhone         string    `gorm:"column:recipient_phone"`
	RecipientAddress       string    `gorm:"column:recipient_address"`
	CreatedAt              time.Time `gorm:"column:created_at"`
	ExpiresAt              time.Time `gorm:"column:expires_at"`
	Payable                bool      `gorm:"column:payable"`
	PaymentStatus          *string   `gorm:"column:payment_status"`
	ReconciliationRequired bool      `gorm:"column:reconciliation_required"`
}

type itemRow struct {
	ID            string `gorm:"column:id"`
	OrderID       string `gorm:"column:order_id"`
	SKUID         string `gorm:"column:sku_id"`
	SKUCode       string `gorm:"column:sku_code"`
	ProductName   string `gorm:"column:product_name"`
	SKUAttributes string `gorm:"column:sku_attributes"`
	Quantity      int    `gorm:"column:quantity"`
	UnitPrice     int64  `gorm:"column:unit_price"`
	Subtotal      int64  `gorm:"column:subtotal"`
	PriceSource   string `gorm:"column:price_source"`
}
