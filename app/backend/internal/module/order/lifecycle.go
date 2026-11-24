package order

import (
	"errors"
	"fmt"

	"gorm.io/gorm"
)

// Outcome is a terminal state a pending_payment order can move to.
type Outcome string

const (
	OutcomeConfirmed     Outcome = "confirmed"
	OutcomePaymentFailed Outcome = "payment_failed"
	OutcomeCancelled     Outcome = "cancelled"
	OutcomeExpired       Outcome = "expired"

	StatusPendingPayment = "pending_payment"
)

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrOrderNotDue   = errors.New("order has not reached its payment deadline")
)

// Transition reports what Finish did. When another path won the race, Applied is false and Status is
// the order's current status, so callers can answer with the real state instead of an error.
type Transition struct {
	OrderID        string
	UserID         string
	Status         string
	Applied        bool
	DeadlinePassed bool
}

// Finish is the single guarded path that ends a pending order and its reservations exactly once.
// It must run inside the caller's transaction so a payment event is committed together with its effect.
//
// Lock order follows data-model.md "Kết thúc order": order row, the order's cart lines by id (only when
// confirming, because only a confirmation deletes them), reservations by sku_id, inventories by sku_id,
// flash sale items by id, then customer usage by item id. Checkout locks cart lines before inventory,
// sale items and usage in the same relative order, so the two paths cannot deadlock.
//
// Deadlines use database time: an order whose deadline has passed can no longer be confirmed or
// cancelled (BR-42), so any request on an overdue order expires it instead. The deadline is checked again
// with clock_timestamp() once every lock is held, so waiting on hot rows cannot let a confirmation slip
// past expires_at; only commit latency remains between that check and the commit.
func Finish(tx *gorm.DB, orderID string, requested Outcome) (Transition, error) {
	var locked struct {
		ID             string `gorm:"column:id"`
		UserID         string `gorm:"column:user_id"`
		Status         string `gorm:"column:status"`
		DeadlinePassed bool   `gorm:"column:deadline_passed"`
	}
	result := tx.Raw(
		`SELECT id, user_id, status, expires_at <= now() AS deadline_passed FROM orders WHERE id = ? FOR UPDATE`,
		orderID,
	).Scan(&locked)
	if result.Error != nil {
		return Transition{}, fmt.Errorf("lock order: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return Transition{}, ErrOrderNotFound
	}
	transition := Transition{
		OrderID: locked.ID, UserID: locked.UserID, Status: locked.Status, DeadlinePassed: locked.DeadlinePassed,
	}
	if locked.Status != StatusPendingPayment {
		return transition, nil
	}
	target := requested
	if locked.DeadlinePassed {
		target = OutcomeExpired
	} else if requested == OutcomeExpired {
		return transition, ErrOrderNotDue
	}

	if target == OutcomeConfirmed {
		if err := tx.Exec(`SELECT ci.id FROM cart_items ci JOIN order_items oi ON oi.cart_item_id = ci.id
			WHERE oi.order_id = ? ORDER BY ci.id FOR UPDATE OF ci`, locked.ID).Error; err != nil {
			return Transition{}, fmt.Errorf("lock ordered cart lines: %w", err)
		}
	}
	if err := lockOrderResources(tx, locked.ID, locked.UserID); err != nil {
		return Transition{}, err
	}
	if target != OutcomeExpired {
		var overdue bool
		if err := tx.Raw(`SELECT expires_at <= clock_timestamp() FROM orders WHERE id = ?`, locked.ID).
			Scan(&overdue).Error; err != nil {
			return Transition{}, fmt.Errorf("recheck deadline: %w", err)
		}
		if overdue {
			target, transition.DeadlinePassed = OutcomeExpired, true
		}
	}
	if err := settleReservations(tx, locked.ID, locked.UserID, target == OutcomeConfirmed); err != nil {
		return Transition{}, err
	}
	update := tx.Exec(
		`UPDATE orders SET status = ?, terminal_at = clock_timestamp(), updated_at = clock_timestamp()
		WHERE id = ? AND status = 'pending_payment'`,
		string(target), locked.ID,
	)
	if update.Error != nil {
		return Transition{}, fmt.Errorf("update order status: %w", update.Error)
	}
	if update.RowsAffected != 1 {
		return Transition{}, fmt.Errorf("order %s left pending_payment while locked", locked.ID)
	}
	transition.Status = string(target)
	transition.Applied = true
	return transition, nil
}

// activeReservationsOf selects the order's still-held reservations; every statement below reuses it.
const activeReservationsOf = `
	FROM inventory_reservations r
	JOIN order_items oi ON oi.id = r.order_item_id
	WHERE oi.order_id = ? AND r.status = 'active'`

func lockOrderResources(tx *gorm.DB, orderID, userID string) error {
	var held []string
	if err := tx.Raw(`SELECT r.id`+activeReservationsOf+` ORDER BY r.sku_id, r.id FOR UPDATE OF r`, orderID).
		Scan(&held).Error; err != nil {
		return fmt.Errorf("lock reservations: %w", err)
	}
	if len(held) == 0 {
		return fmt.Errorf("pending order %s has no active reservations", orderID)
	}
	var skuIDs, itemIDs, usageIDs []string
	if err := tx.Raw(`SELECT sku_id FROM inventories WHERE sku_id IN (SELECT r.sku_id`+activeReservationsOf+`)
		ORDER BY sku_id FOR UPDATE`, orderID).Scan(&skuIDs).Error; err != nil {
		return fmt.Errorf("lock inventories: %w", err)
	}
	if err := tx.Raw(`SELECT id FROM flash_sale_items WHERE id IN (SELECT oi.flash_sale_item_id`+activeReservationsOf+`)
		ORDER BY id FOR UPDATE`, orderID).Scan(&itemIDs).Error; err != nil {
		return fmt.Errorf("lock flash sale items: %w", err)
	}
	if err := tx.Raw(`SELECT flash_sale_item_id FROM flash_sale_customer_usage
		WHERE user_id = ? AND flash_sale_item_id IN (SELECT oi.flash_sale_item_id`+activeReservationsOf+`)
		ORDER BY flash_sale_item_id FOR UPDATE`, userID, orderID).Scan(&usageIDs).Error; err != nil {
		return fmt.Errorf("lock customer usage: %w", err)
	}
	if len(usageIDs) != len(itemIDs) {
		return fmt.Errorf("order %s: %d flash sale items but %d usage rows", orderID, len(itemIDs), len(usageIDs))
	}
	return nil
}

// settleReservations moves every held reservation to sold (confirmed) or released, and adjusts the
// counters by the same quantities: confirming spends on_hand and quota; releasing only returns the hold.
func settleReservations(tx *gorm.DB, orderID, userID string, sold bool) error {
	steps := []struct {
		name string
		sql  string
		args []any
	}{
		{"inventories", `UPDATE inventories i
			SET reserved = i.reserved - d.quantity,
				on_hand = i.on_hand - CASE WHEN ? THEN d.quantity ELSE 0 END,
				version = i.version + 1, updated_at = now()
			FROM (SELECT r.sku_id, SUM(r.quantity) AS quantity` + activeReservationsOf + ` GROUP BY r.sku_id) d
			WHERE i.sku_id = d.sku_id`, []any{sold, orderID}},
		{"flash sale quota", `UPDATE flash_sale_items f
			SET quota_reserved = f.quota_reserved - d.quantity,
				quota_sold = f.quota_sold + CASE WHEN ? THEN d.quantity ELSE 0 END,
				updated_at = now()
			FROM (SELECT oi.flash_sale_item_id AS id, SUM(r.quantity) AS quantity` + activeReservationsOf + `
				AND oi.flash_sale_item_id IS NOT NULL GROUP BY oi.flash_sale_item_id) d
			WHERE f.id = d.id`, []any{sold, orderID}},
		{"customer usage", `UPDATE flash_sale_customer_usage u
			SET reserved_quantity = u.reserved_quantity - d.quantity,
				sold_quantity = u.sold_quantity + CASE WHEN ? THEN d.quantity ELSE 0 END,
				updated_at = now()
			FROM (SELECT oi.flash_sale_item_id AS id, SUM(r.quantity) AS quantity` + activeReservationsOf + `
				AND oi.flash_sale_item_id IS NOT NULL GROUP BY oi.flash_sale_item_id) d
			WHERE u.flash_sale_item_id = d.id AND u.user_id = ?`, []any{sold, orderID, userID}},
	}
	for _, step := range steps {
		if err := tx.Exec(step.sql, step.args...).Error; err != nil {
			return fmt.Errorf("settle %s: %w", step.name, err)
		}
	}

	reservationStatus := "released"
	if sold {
		reservationStatus = "finalized"
	}
	if err := tx.Exec(`UPDATE inventory_reservations r SET status = ?, released_at = clock_timestamp()
		FROM order_items oi WHERE oi.id = r.order_item_id AND oi.order_id = ? AND r.status = 'active'`,
		reservationStatus, orderID).Error; err != nil {
		return fmt.Errorf("settle reservations: %w", err)
	}
	if !sold {
		return nil
	}
	// BR-13: only cart lines still at the version that was ordered are removed; edited lines stay.
	if err := tx.Exec(`DELETE FROM cart_items ci USING order_items oi
		WHERE oi.order_id = ? AND oi.cart_item_id = ci.id AND ci.version = oi.cart_item_version`,
		orderID).Error; err != nil {
		return fmt.Errorf("remove purchased cart lines: %w", err)
	}
	return nil
}
