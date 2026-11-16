package backend_test

import "testing"

func TestOrderMigrationEnforcesIdempotencyAndReservationConstraints(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()

	userID := "61000000-0000-4000-8000-000000000001"
	if err := database.Exec(`INSERT INTO users (id, email, email_normalized, password_hash, role, status)
		VALUES (?, 'order-user@example.com', 'order-user@example.com', 'hash', 'customer', 'active')`, userID).Error; err != nil {
		t.Fatalf("insert user: %v", err)
	}

	if err := database.Exec(`INSERT INTO idempotency_records (user_id, operation, idempotency_key, request_hash)
		VALUES (?, 'place_order', 'checkout-key-0001', decode(repeat('aa', 32), 'hex'))`, userID).Error; err != nil {
		t.Fatalf("insert idempotency record: %v", err)
	}
	if err := database.Exec(`INSERT INTO idempotency_records (user_id, operation, idempotency_key, request_hash)
		VALUES (?, 'place_order', 'checkout-key-0001', decode(repeat('bb', 32), 'hex'))`, userID).Error; err == nil {
		t.Fatal("expected unique idempotency identity violation")
	}

	if err := database.Exec(`INSERT INTO orders (user_id, order_number, status, currency, total_amount, recipient_name, recipient_phone, recipient_address, expires_at)
		VALUES (?, 'ORD-0001', 'pending_payment', 'VND', 100000, 'Nguyen Van A', '0900000000', 'Ha Noi', now() + interval '5 minutes')`, userID).Error; err != nil {
		t.Fatalf("insert pending order: %v", err)
	}
	if err := database.Exec(`INSERT INTO orders (user_id, order_number, status, currency, total_amount, recipient_name, recipient_phone, recipient_address, expires_at)
		VALUES (?, 'ORD-0002', 'unsupported', 'VND', 100000, 'Nguyen Van A', '0900000000', 'Ha Noi', now() + interval '5 minutes')`, userID).Error; err == nil {
		t.Fatal("expected order-status check violation")
	}
}
