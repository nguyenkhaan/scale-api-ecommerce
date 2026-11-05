package backend_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"cloudian/cloudian-scale/internal/module/cart"
)

func TestCartConcurrentFirstAccessCreatesOneCart(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)

	for round := 0; round < 10; round++ {
		customer := cartNewCustomer(t, router, database, "cart-first-access")
		responses := make([]*httptest.ResponseRecorder, 6)
		cartRunParallel(len(responses), func(i int) {
			responses[i] = customer.parallel(router).send(t, http.MethodGet, "/api/v1/cart", nil)
		})

		cartIDs := map[string]bool{}
		for i, response := range responses {
			if response.Code != http.StatusOK {
				t.Fatalf("round %d request %d status = %d; body = %s", round, i, response.Code, response.Body.String())
			}
			cartIDs[decodeJSON[cart.CartResponse](t, response).ID] = true
		}
		if len(cartIDs) != 1 {
			t.Errorf("round %d: responses reported %d distinct carts, want 1", round, len(cartIDs))
		}
		if n := cartCount(t, database, `SELECT count(*) FROM carts WHERE user_id = ? AND status = 'active'`, customer.ID); n != 1 {
			t.Errorf("round %d: active carts = %d, want 1", round, n)
		}
	}
}

func TestCartConcurrentFirstAddsCreateOneCart(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	skus := make([]string, 6)
	for i := range skus {
		skus[i] = cartSeedSKU(t, database, cartSKUOptions{})
	}

	for round := 0; round < 10; round++ {
		customer := cartNewCustomer(t, router, database, "cart-first-add")
		responses := make([]*httptest.ResponseRecorder, len(skus))
		cartRunParallel(len(skus), func(i int) {
			responses[i] = cartPut(t, customer.parallel(router), skus[i], map[string]any{"quantity": 1})
		})
		for i, response := range responses {
			if response.Code != http.StatusOK {
				t.Fatalf("round %d add %d status = %d; body = %s", round, i, response.Code, response.Body.String())
			}
		}
		if n := cartCount(t, database, `SELECT count(*) FROM carts WHERE user_id = ? AND status = 'active'`, customer.ID); n != 1 {
			t.Errorf("round %d: active carts = %d, want 1", round, n)
		}
		if view := cartGet(t, customer.testClient); len(view.Items) != len(skus) {
			t.Errorf("round %d: cart lines = %d, want %d", round, len(view.Items), len(skus))
		}
	}
}

func TestCartConcurrentAddsNeverExceedTwentyLines(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	existing := make([]string, cart.MaxCartItems-1)
	for i := range existing {
		existing[i] = cartSeedSKU(t, database, cartSKUOptions{})
	}
	newSKUs := make([]string, 8)
	for i := range newSKUs {
		newSKUs[i] = cartSeedSKU(t, database, cartSKUOptions{})
	}

	for round := 0; round < 10; round++ {
		customer := cartNewCustomer(t, router, database, "cart-limit-race")
		cartGet(t, customer.testClient)
		for _, skuID := range existing {
			cartExec(t, database,
				`INSERT INTO cart_items (cart_id, sku_id, quantity, intended_price_source)
				 SELECT id, ?, 1, 'regular' FROM carts WHERE user_id = ? AND status = 'active'`,
				skuID, customer.ID)
		}

		responses := make([]*httptest.ResponseRecorder, len(newSKUs))
		cartRunParallel(len(newSKUs), func(i int) {
			responses[i] = cartPut(t, customer.parallel(router), newSKUs[i], map[string]any{"quantity": 1})
		})

		succeeded := 0
		for i, response := range responses {
			switch response.Code {
			case http.StatusOK:
				succeeded++
			case http.StatusUnprocessableEntity:
			default:
				t.Errorf("round %d add %d status = %d; body = %s", round, i, response.Code, response.Body.String())
			}
		}
		if succeeded != 1 {
			t.Errorf("round %d: successful adds = %d, want exactly 1 free slot", round, succeeded)
		}
		lines := cartCount(t, database,
			`SELECT count(*) FROM cart_items ci JOIN carts c ON c.id = ci.cart_id WHERE c.user_id = ?`, customer.ID)
		if lines != cart.MaxCartItems {
			t.Errorf("round %d: cart lines = %d, want %d", round, lines, cart.MaxCartItems)
		}
	}
}

func TestCartConcurrentWritersWithSameExpectedVersion(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cart-lost-update")
	skuID := cartSeedSKU(t, database, cartSKUOptions{})
	item := cartAdd(t, customer.testClient, skuID, 1)

	responses := make([]*httptest.ResponseRecorder, 5)
	cartRunParallel(len(responses), func(i int) {
		responses[i] = cartPut(t, customer.parallel(router), skuID, map[string]any{"quantity": i + 1, "expected_version": item.Version})
	})

	succeeded, conflicted := 0, 0
	for i, response := range responses {
		switch response.Code {
		case http.StatusOK:
			succeeded++
		case http.StatusConflict:
			requireProblem(t, response, http.StatusConflict, "CART_ITEM_CHANGED")
			conflicted++
		default:
			t.Errorf("writer %d status = %d; body = %s", i, response.Code, response.Body.String())
		}
	}
	if succeeded != 1 || conflicted != len(responses)-1 {
		t.Errorf("200s = %d, 409s = %d; want 1 and %d", succeeded, conflicted, len(responses)-1)
	}
	if final := cartGet(t, customer.testClient).Items[0]; final.Version != 2 {
		t.Errorf("final version = %d, want 2 (one applied write)", final.Version)
	}
}

func TestCartConcurrentUnversionedWritesBumpVersionOncePerWrite(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	customer := cartNewCustomer(t, router, database, "cart-last-write")
	skuID := cartSeedSKU(t, database, cartSKUOptions{})
	cartAdd(t, customer.testClient, skuID, 1)

	const writers = 5
	responses := make([]*httptest.ResponseRecorder, writers)
	cartRunParallel(writers, func(i int) {
		responses[i] = cartPut(t, customer.parallel(router), skuID, map[string]any{"quantity": i + 1})
	})
	for i, response := range responses {
		if response.Code != http.StatusOK {
			t.Fatalf("writer %d status = %d; body = %s", i, response.Code, response.Body.String())
		}
	}
	if final := cartGet(t, customer.testClient).Items[0]; final.Version != 1+writers {
		t.Errorf("final version = %d, want %d (every write counted)", final.Version, 1+writers)
	}
}

func TestCartConcurrentAddsOfSameNewSKUKeepOneLine(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := cartNewRouter(t, database)
	skuID := cartSeedSKU(t, database, cartSKUOptions{})

	for round := 0; round < 10; round++ {
		customer := cartNewCustomer(t, router, database, "cart-same-sku")
		responses := make([]*httptest.ResponseRecorder, 2)
		cartRunParallel(len(responses), func(i int) {
			responses[i] = cartPut(t, customer.parallel(router), skuID, map[string]any{"quantity": i + 1})
		})
		for i, response := range responses {
			if response.Code != http.StatusOK {
				t.Fatalf("round %d add %d status = %d; body = %s", round, i, response.Code, response.Body.String())
			}
		}
		if view := cartGet(t, customer.testClient); len(view.Items) != 1 {
			t.Errorf("round %d: lines = %d, want 1", round, len(view.Items))
		}
	}
}
