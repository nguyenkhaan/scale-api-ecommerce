package backend_test

import (
	"net/http"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/module/order"
	"gorm.io/gorm"
)

// TEST-14: concurrent cancels of one order release stock and quota exactly once.
func TestOrderCancelConcurrentReleasesOnce(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := orderAPINewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cancel-race")

	saleSKU := cartSeedSKU(t, database, cartSKUOptions{OnHand: 20})
	regularSKU := cartSeedSKU(t, database, cartSKUOptions{OnHand: 20})
	sale := cartSeedSale(t, database, saleSKU, cartSaleOptions{QuotaTotal: 5, PurchaseLimit: 3})
	// A second pending order keeps the same SKU reserved so a double release would be visible as drift.
	bystander := lifecycleSeedPendingOrder(t, database, customer.ID, 300,
		lifecycleLine{SKUID: regularSKU, Quantity: 3, UnitPrice: 10_000})
	orderID := lifecycleSeedPendingOrder(t, database, customer.ID, 300,
		lifecycleLine{SKUID: saleSKU, Quantity: 2, UnitPrice: 40_000, FlashSaleItemID: sale},
		lifecycleLine{SKUID: regularSKU, Quantity: 1, UnitPrice: 100_000},
	)
	saleBefore := lifecycleReadCounters(t, database, saleSKU, sale, customer.ID)
	regularBefore := lifecycleReadCounters(t, database, regularSKU, "", customer.ID)

	const requests = 8
	statuses := make([]int, requests)
	bodies := make([]order.OrderResponse, requests)
	cartRunParallel(requests, func(i int) {
		client := customer.parallel(router)
		response := client.send(t, http.MethodPost, "/api/orders/"+orderID+"/cancel", nil, client.csrfHeader()...)
		statuses[i] = response.Code
		if response.Code == http.StatusOK {
			bodies[i] = decodeJSON[order.OrderResponse](t, response)
		}
	})
	for i := range requests {
		if statuses[i] != http.StatusOK || bodies[i].Status != "cancelled" || bodies[i].Payable {
			t.Errorf("request %d = %d %+v, want 200 cancelled and not payable", i, statuses[i], bodies[i])
		}
	}

	saleAfter := lifecycleReadCounters(t, database, saleSKU, sale, customer.ID)
	wantSale := saleBefore
	wantSale.Reserved -= 2
	wantSale.QuotaReserved -= 2
	wantSale.UsageReserved -= 2
	if saleAfter != wantSale {
		t.Errorf("flash sale counters = %+v, want %+v (released once)", saleAfter, wantSale)
	}
	regularAfter := lifecycleReadCounters(t, database, regularSKU, "", customer.ID)
	if regularAfter.Reserved != regularBefore.Reserved-1 || regularAfter.OnHand != regularBefore.OnHand {
		t.Errorf("regular counters = %+v, want reserved %d (3 held by the other order)", regularAfter, regularBefore.Reserved-1)
	}
	if got := orderAPIStatus(t, database, bystander); got != "pending_payment" {
		t.Errorf("unrelated order status = %s, want pending_payment", got)
	}
	lifecycleRequireInvariants(t, database)
}

func TestOrderCancelRequiresOwnerAndCSRF(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := orderAPINewRouter(t, database)
	owner := cartNewCustomer(t, router, database, "cancel-owner")
	intruder := cartNewCustomer(t, router, database, "cancel-intruder")
	orderID, skuID := orderAPISeedOrder(t, database, owner.ID, 300, 2)

	status, body := orderAPICancel(t, intruder.testClient, orderID)
	if status != http.StatusNotFound {
		t.Fatalf("foreign cancel status = %d; body = %s, want 404", status, body)
	}
	requireProblem(t, intruder.send(t, http.MethodPost, "/api/orders/"+orderID+"/cancel", nil, intruder.csrfHeader()...),
		http.StatusNotFound, "RESOURCE_NOT_FOUND")
	if got := orderAPIStatus(t, database, orderID); got != "pending_payment" || orderAPIReserved(t, database, skuID) != 2 {
		t.Errorf("foreign cancel changed the order: status %s reserved %d", got, orderAPIReserved(t, database, skuID))
	}

	requireProblem(t, owner.send(t, http.MethodPost, "/api/orders/"+orderID+"/cancel", nil),
		http.StatusForbidden, "FORBIDDEN")
	requireProblem(t, newTestClient(router).send(t, http.MethodPost, "/api/orders/"+orderID+"/cancel", nil),
		http.StatusUnauthorized, "UNAUTHENTICATED")
	if got := orderAPIStatus(t, database, orderID); got != "pending_payment" {
		t.Errorf("request without CSRF cancelled the order: %s", got)
	}
	lifecycleRequireInvariants(t, database)
}

func TestOrderCancelOverdueOrderExpiresInstead(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := orderAPINewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cancel-overdue")
	orderID, skuID := orderAPISeedOrder(t, database, customer.ID, -5, 2)

	response := customer.send(t, http.MethodPost, "/api/orders/"+orderID+"/cancel", nil, customer.csrfHeader()...)
	problemBody := requireProblem(t, response, http.StatusConflict, "STATE_TRANSITION_CONFLICT")
	if problemBody.Detail == "" || !strings.Contains(problemBody.Detail, "expired") {
		t.Errorf("detail = %q, want it to name the expired status", problemBody.Detail)
	}
	if got := orderAPIStatus(t, database, orderID); got != "expired" {
		t.Errorf("status = %s, want expired", got)
	}
	if reserved := orderAPIReserved(t, database, skuID); reserved != 0 {
		t.Errorf("reserved = %d, want stock released once", reserved)
	}
	// A repeat stays a conflict and must not release again.
	requireProblem(t, customer.send(t, http.MethodPost, "/api/orders/"+orderID+"/cancel", nil, customer.csrfHeader()...),
		http.StatusConflict, "STATE_TRANSITION_CONFLICT")
	if reserved := orderAPIReserved(t, database, skuID); reserved != 0 {
		t.Errorf("reserved = %d after repeat, want 0", reserved)
	}
	lifecycleRequireInvariants(t, database)
}

func TestOrderCancelRejectsOtherTerminalStates(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := orderAPINewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cancel-terminal")

	for _, outcome := range []order.Outcome{order.OutcomeConfirmed, order.OutcomePaymentFailed} {
		orderID, _ := orderAPISeedOrder(t, database, customer.ID, 300, 1)
		if err := database.Transaction(func(tx *gorm.DB) error {
			_, err := order.Finish(tx, orderID, outcome)
			return err
		}); err != nil {
			t.Fatalf("finish %s: %v", outcome, err)
		}
		response := customer.send(t, http.MethodPost, "/api/orders/"+orderID+"/cancel", nil, customer.csrfHeader()...)
		problemBody := requireProblem(t, response, http.StatusConflict, "STATE_TRANSITION_CONFLICT")
		if !strings.Contains(problemBody.Detail, string(outcome)) {
			t.Errorf("detail %q does not state the %s status", problemBody.Detail, outcome)
		}
		if got := orderAPIStatus(t, database, orderID); got != string(outcome) {
			t.Errorf("status = %s, want %s unchanged", got, outcome)
		}
	}
	lifecycleRequireInvariants(t, database)
}
