package backend_test

import (
	"bytes"
	"log"
	"net/http"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/module/order"
)

func TestCheckoutCountersAndLogsOmitSecrets(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "checkout-log-secret")
	skuID := cartSeedSKU(t, database, cartSKUOptions{Price: 100_000, OnHand: 10})
	line := cartAdd(t, customer.testClient, skuID, 1)
	request := checkoutRequestFromPreview(checkoutPreviewRequest(t, customer.testClient, line))
	counters := order.NewCheckoutCounters()
	service, err := order.NewCheckoutService(database, counters)
	if err != nil {
		t.Fatalf("create checkout service: %v", err)
	}
	var logs bytes.Buffer
	oldWriter := log.Writer()
	log.SetOutput(&logs)
	t.Cleanup(func() { log.SetOutput(oldWriter) })
	const secretKey = "secret-idempotency-key-01"
	if _, err := checkoutPlaceOrder(t, service, customer.ID, secretKey, request); err != nil {
		t.Fatalf("create order: %v", err)
	}
	if _, err := checkoutPlaceOrder(t, service, customer.ID, secretKey, request); err != nil {
		t.Fatalf("replay order: %v", err)
	}
	cartExec(t, database, `UPDATE skus SET regular_price = 110000 WHERE id = ?`, skuID)
	if result, err := checkoutPlaceOrder(t, service, customer.ID, "secret-rejected-key-01", request); err != nil || result.Status != http.StatusConflict {
		t.Fatalf("reject price change = %+v, %v", result, err)
	}
	snapshot := counters.Snapshot()
	for outcome, want := range map[order.CheckoutOutcome]int64{
		order.OutcomeCreated: 1, order.OutcomeReplayed: 1, order.OutcomeRejected: 1,
	} {
		if snapshot[outcome] != want {
			t.Errorf("counter %s = %d, want %d (snapshot %+v)", outcome, snapshot[outcome], want, snapshot)
		}
	}
	for _, secret := range []string{secretKey, request.Recipient.Phone, request.Recipient.Address, customer.ID} {
		if strings.Contains(logs.String(), secret) {
			t.Errorf("checkout log contains secret %q: %s", secret, logs.String())
		}
	}
	lifecycleRequireInvariants(t, database)
}
