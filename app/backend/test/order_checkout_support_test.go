package backend_test

import (
	"testing"

	"cloudian/cloudian-scale/internal/module/cart"
	"cloudian/cloudian-scale/internal/module/order"
	"gorm.io/gorm"
)

func checkoutRequestFromPreview(preview cart.CheckoutPreviewResponse) order.PlaceOrderRequest {
	lines := make([]order.PlaceOrderLine, len(preview.Lines))
	for i, line := range preview.Lines {
		lines[i] = order.PlaceOrderLine{
			CartItemID: line.CartItemID, CartItemVersion: line.CartItemVersion, SKUID: line.SKUID,
			Quantity: line.Quantity, UnitPrice: line.UnitPrice, Subtotal: line.Subtotal,
			PriceSource: line.PriceSource, FlashSaleItemID: line.FlashSaleItemID,
		}
	}
	return order.PlaceOrderRequest{
		Lines: lines, Currency: preview.Currency, ExpectedTotal: preview.TotalAmount,
		Recipient: order.RecipientRequest{
			Name: "Nguyễn Văn An", Phone: "0901234567", Address: "12 Nguyễn Trãi, Quận 1",
		},
	}
}

func checkoutNewService(t *testing.T, database *gorm.DB) *order.CheckoutService {
	t.Helper()
	service, err := order.NewCheckoutService(database, nil)
	if err != nil {
		t.Fatalf("create checkout service: %v", err)
	}
	return service
}

func checkoutPreviewRequest(t *testing.T, client *testClient, items ...cart.CartItemResponse) cart.CheckoutPreviewResponse {
	t.Helper()
	response := cartPreview(t, client, items...)
	if response.Code != 200 {
		t.Fatalf("checkout preview status = %d; body = %s", response.Code, response.Body.String())
	}
	return decodeJSON[cart.CheckoutPreviewResponse](t, response)
}
