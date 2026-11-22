package order

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"gorm.io/gorm"
)

const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)

var (
	// ErrInvalidStatus means the status filter is not an order status.
	ErrInvalidStatus = errors.New("invalid order status")
	// ErrInvalidLimit means the page size is outside 1..MaxPageLimit.
	ErrInvalidLimit = errors.New("invalid order page limit")
)

var orderStatuses = []string{StatusPendingPayment, "confirmed", "payment_failed", "cancelled", "expired"}

// OrderQueries is the customer read model and customer-driven transitions over PostgreSQL.
type OrderQueries struct {
	db *gorm.DB
}

func NewOrderQueries(db *gorm.DB) (*OrderQueries, error) {
	if db == nil {
		return nil, errors.New("order queries require a database")
	}
	return &OrderQueries{db: db}, nil
}

// ListParams are already-parsed list arguments; callers apply DefaultPageLimit themselves.
type ListParams struct {
	Cursor string
	Limit  int
	Status string
}

// orderSelect computes payable and the payment projections in the database so the answer shares one
// clock and one snapshot with the order row.
const orderSelect = `SELECT o.id, o.order_number, o.status, o.currency, o.total_amount,
	o.recipient_name, o.recipient_phone, o.recipient_address, o.created_at, o.expires_at,
	(o.status = 'pending_payment' AND o.expires_at > now()) AS payable,
	(SELECT ps.status FROM payment_sessions ps WHERE ps.order_id = o.id) AS payment_status,
	EXISTS (SELECT 1 FROM payment_reconciliations pr WHERE pr.order_id = o.id AND pr.status = 'open')
		AS reconciliation_required
	FROM orders o`

// ListOrders returns the caller's orders newest first, one cursor page at a time, using two queries
// regardless of page size: the page itself and the lines of every order on it.
func (q *OrderQueries) ListOrders(ctx context.Context, userID string, params ListParams) (OrderPage, error) {
	limit := params.Limit
	if limit < 1 || limit > MaxPageLimit {
		return OrderPage{}, ErrInvalidLimit
	}
	if params.Status != "" && !slices.Contains(orderStatuses, params.Status) {
		return OrderPage{}, ErrInvalidStatus
	}
	query := orderSelect + ` WHERE o.user_id = ?`
	args := []any{userID}
	if params.Status != "" {
		query += ` AND o.status = ?`
		args = append(args, params.Status)
	}
	if params.Cursor != "" {
		cursor, err := decodeOrderCursor(params.Cursor, params.Status)
		if err != nil {
			return OrderPage{}, err
		}
		query += ` AND (o.created_at, o.id) < (?, ?)`
		args = append(args, cursor.CreatedAt, cursor.ID)
	}
	query += ` ORDER BY o.created_at DESC, o.id DESC LIMIT ?`
	args = append(args, limit+1)

	var rows []orderRow
	if err := q.db.WithContext(ctx).Raw(query, args...).Scan(&rows).Error; err != nil {
		return OrderPage{}, fmt.Errorf("list orders: %w", err)
	}
	var nextCursor *string
	if len(rows) > limit {
		rows = rows[:limit]
		last := rows[len(rows)-1]
		next, err := encodeOrderCursor(orderCursor{CreatedAt: last.CreatedAt, ID: last.ID, Status: params.Status})
		if err != nil {
			return OrderPage{}, fmt.Errorf("encode cursor: %w", err)
		}
		nextCursor = &next
	}
	items, err := attachItems(ctx, q.db, rows)
	if err != nil {
		return OrderPage{}, err
	}
	return OrderPage{Items: items, NextCursor: nextCursor}, nil
}

// GetOrder returns one order owned by userID; other users' orders look like missing ones.
func (q *OrderQueries) GetOrder(ctx context.Context, userID, orderID string) (OrderResponse, error) {
	return getOwnedOrder(ctx, q.db, userID, orderID)
}

func getOwnedOrder(ctx context.Context, db *gorm.DB, userID, orderID string) (OrderResponse, error) {
	var rows []orderRow
	if err := db.WithContext(ctx).Raw(orderSelect+` WHERE o.id = ? AND o.user_id = ?`, orderID, userID).
		Scan(&rows).Error; err != nil {
		return OrderResponse{}, fmt.Errorf("get order: %w", err)
	}
	if len(rows) == 0 {
		return OrderResponse{}, ErrOrderNotFound
	}
	orders, err := attachItems(ctx, db, rows)
	if err != nil {
		return OrderResponse{}, err
	}
	return orders[0], nil
}

// attachItems builds the responses and loads the lines of every order with a single IN query.
func attachItems(ctx context.Context, db *gorm.DB, rows []orderRow) ([]OrderResponse, error) {
	orders := make([]OrderResponse, len(rows))
	if len(rows) == 0 {
		return orders, nil
	}
	ids := make([]string, len(rows))
	position := make(map[string]int, len(rows))
	for i, row := range rows {
		ids[i] = row.ID
		position[row.ID] = i
		orders[i] = OrderResponse{
			ID: row.ID, OrderNumber: row.OrderNumber, Status: row.Status, Currency: row.Currency,
			TotalAmount: row.TotalAmount, CreatedAt: row.CreatedAt, ExpiresAt: row.ExpiresAt,
			Recipient: RecipientResponse{Name: row.RecipientName, Phone: row.RecipientPhone, Address: row.RecipientAddress},
			Items:     []OrderItemResponse{}, Payable: row.Payable, PaymentStatus: row.PaymentStatus,
			ReconciliationRequired: row.ReconciliationRequired,
		}
	}
	var items []itemRow
	if err := db.WithContext(ctx).Raw(
		`SELECT id, order_id, sku_id, sku_code, product_name, sku_attributes::text AS sku_attributes,
			quantity, unit_price, subtotal, price_source
		 FROM order_items WHERE order_id IN ? ORDER BY order_id, sku_code, id`, ids,
	).Scan(&items).Error; err != nil {
		return nil, fmt.Errorf("load order items: %w", err)
	}
	for _, item := range items {
		index := position[item.OrderID]
		orders[index].Items = append(orders[index].Items, OrderItemResponse{
			ID: item.ID, SKUID: item.SKUID, SKUCode: item.SKUCode, ProductName: item.ProductName,
			SKUAttributes: json.RawMessage(item.SKUAttributes), Quantity: item.Quantity,
			UnitPrice: item.UnitPrice, Subtotal: item.Subtotal, PriceSource: item.PriceSource,
		})
	}
	return orders, nil
}
