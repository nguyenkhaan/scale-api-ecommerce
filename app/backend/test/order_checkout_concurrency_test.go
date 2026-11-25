package backend_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/cart"
	"cloudian/cloudian-scale/internal/module/order"
)

func TestCheckoutLastUnitsContention(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := newCheckoutAPIRouter(t, database)
	const buyers = 12
	skuID := cartSeedSKU(t, database, cartSKUOptions{OnHand: 5})
	customers := make([]cartCustomer, buyers)
	requests := make([]order.PlaceOrderRequest, buyers)
	for i := range buyers {
		customers[i] = cartNewCustomer(t, router, database, fmt.Sprintf("checkout-last-units-%d", i))
		line := cartAdd(t, customers[i].testClient, skuID, 1)
		requests[i] = checkoutRequestFromPreview(checkoutPreviewRequest(t, customers[i].testClient, line))
	}
	responses := make([]*httptest.ResponseRecorder, buyers)
	cartRunParallel(buyers, func(i int) {
		responses[i] = checkoutPost(t, customers[i].testClient, fmt.Sprintf("last-unit-key-%02d", i), requests[i])
	})
	created, rejected := 0, 0
	for i, response := range responses {
		switch response.Code {
		case http.StatusCreated:
			created++
		case http.StatusConflict:
			requireProblem(t, response, http.StatusConflict, problem.CodeInsufficientStock)
			rejected++
		default:
			t.Errorf("buyer %d response = %d %s; want 201 or 409", i, response.Code, response.Body.String())
		}
	}
	if created != 5 || rejected != 7 {
		t.Errorf("created/rejected = %d/%d, want 5/7", created, rejected)
	}
	if count := cartCount(t, database, `SELECT count(*) FROM orders`); count != 5 {
		t.Errorf("orders = %d, want 5", count)
	}
	var reserved int64
	if err := database.Raw(`SELECT reserved FROM inventories WHERE sku_id = ?`, skuID).Scan(&reserved).Error; err != nil {
		t.Fatalf("read reserved inventory: %v", err)
	}
	if reserved != 5 {
		t.Errorf("reserved = %d, want 5", reserved)
	}
	lifecycleRequireInvariants(t, database)
}

func TestCheckoutOverlappingCartsInReverseOrderNeverPartial(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := newCheckoutAPIRouter(t, database)
	for round := 0; round < 20; round++ {
		skuA := cartSeedSKU(t, database, cartSKUOptions{OnHand: 1})
		skuB := cartSeedSKU(t, database, cartSKUOptions{OnHand: 1})
		first := cartNewCustomer(t, router, database, fmt.Sprintf("checkout-reverse-a-%d", round))
		second := cartNewCustomer(t, router, database, fmt.Sprintf("checkout-reverse-b-%d", round))
		firstA, firstB := cartAdd(t, first.testClient, skuA, 1), cartAdd(t, first.testClient, skuB, 1)
		secondB, secondA := cartAdd(t, second.testClient, skuB, 1), cartAdd(t, second.testClient, skuA, 1)
		firstRequest := checkoutRequestFromPreview(checkoutPreviewRequest(t, first.testClient, firstA, firstB))
		secondRequest := checkoutRequestFromPreview(checkoutPreviewRequest(t, second.testClient, secondB, secondA))
		responses := make([]*httptest.ResponseRecorder, 2)
		cartRunParallel(2, func(i int) {
			if i == 0 {
				responses[i] = checkoutPost(t, first.testClient, fmt.Sprintf("reverse-key-a-%02d", round), firstRequest)
				return
			}
			responses[i] = checkoutPost(t, second.testClient, fmt.Sprintf("reverse-key-b-%02d", round), secondRequest)
		})
		created, rejected := 0, 0
		for i, response := range responses {
			switch response.Code {
			case http.StatusCreated:
				created++
				createdOrder := decodeJSON[order.OrderResponse](t, response)
				if len(createdOrder.Items) != 2 {
					t.Errorf("round %d winner %d has %d items, want full 2-SKU order", round, i, len(createdOrder.Items))
				}
			case http.StatusConflict:
				body := decodeJSON[problem.Problem](t, response)
				if body.Code != problem.CodeInsufficientStock {
					t.Errorf("round %d loser %d problem = %q, want INSUFFICIENT_STOCK", round, i, body.Code)
				}
				rejected++
			default:
				t.Errorf("round %d client %d response = %d %s; want 201 or 409", round, i, response.Code, response.Body.String())
			}
		}
		if created != 1 || rejected != 1 {
			t.Errorf("round %d created/rejected = %d/%d, want 1/1", round, created, rejected)
		}
		if count := cartCount(t, database, `SELECT count(*) FROM orders WHERE user_id IN ?`, []string{first.ID, second.ID}); count != 1 {
			t.Errorf("round %d orders = %d, want 1 complete order", round, count)
		}
	}
	lifecycleRequireInvariants(t, database)
}

func TestCheckoutPurchaseLimitAcrossKeys(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := newCheckoutAPIRouter(t, database)
	customer := cartNewCustomer(t, router, database, "checkout-purchase-limit")
	skuID := cartSeedSKU(t, database, cartSKUOptions{OnHand: 10})
	saleItem := cartSeedSale(t, database, skuID, cartSaleOptions{QuotaTotal: 10, PurchaseLimit: 2})
	lineResponse := cartPut(t, customer.testClient, skuID, map[string]any{
		"quantity": 1, "price_source": "flash_sale", "flash_sale_item_id": saleItem,
	})
	if lineResponse.Code != http.StatusOK {
		t.Fatalf("add flash sale line = %d %s", lineResponse.Code, lineResponse.Body.String())
	}
	line := cartFindItem(t, decodeJSON[cart.CartResponse](t, lineResponse), skuID)
	request := checkoutRequestFromPreview(checkoutPreviewRequest(t, customer.testClient, line))
	const attempts = 4
	responses := make([]*httptest.ResponseRecorder, attempts)
	cartRunParallel(attempts, func(i int) {
		responses[i] = checkoutPost(t, customer.parallel(router), fmt.Sprintf("purchase-limit-key-%02d", i), request)
	})
	created, rejected := 0, 0
	for i, response := range responses {
		switch response.Code {
		case http.StatusCreated:
			created++
		case http.StatusConflict:
			requireProblem(t, response, http.StatusConflict, problem.CodePurchaseLimitExceeded)
			rejected++
		default:
			t.Errorf("attempt %d response = %d %s; want 201 or 409", i, response.Code, response.Body.String())
		}
	}
	if created != 2 || rejected != 2 {
		t.Errorf("created/rejected = %d/%d, want 2/2", created, rejected)
	}
	var usage int64
	if err := database.Raw(`SELECT reserved_quantity FROM flash_sale_customer_usage WHERE user_id = ? AND flash_sale_item_id = ?`, customer.ID, saleItem).Scan(&usage).Error; err != nil {
		t.Fatalf("read customer sale usage: %v", err)
	}
	if usage != 2 {
		t.Errorf("reserved customer usage = %d, want 2", usage)
	}
	lifecycleRequireInvariants(t, database)
}
