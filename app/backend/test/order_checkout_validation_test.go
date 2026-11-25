package backend_test

import (
	"math"
	"net/http/httptest"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/order"
	"github.com/gin-gonic/gin"
)

func validCheckoutRequest() order.PlaceOrderRequest {
	return order.PlaceOrderRequest{
		Lines: []order.PlaceOrderLine{{
			CartItemID:      "10000000-0000-4000-8000-000000000001",
			CartItemVersion: 1,
			SKUID:           "20000000-0000-4000-8000-000000000001",
			Quantity:        2,
			UnitPrice:       100,
			Subtotal:        200,
			PriceSource:     "regular",
		}},
		Currency:      "VND",
		ExpectedTotal: 200,
		Recipient: order.RecipientRequest{
			Name: "An", Phone: "09012345", Address: "12 Nguyen Trai",
		},
	}
}

func checkoutViolationFields(violations []problem.Violation) map[string]bool {
	fields := make(map[string]bool, len(violations))
	for _, violation := range violations {
		fields[violation.Field] = true
	}
	return fields
}

func TestValidatePlaceOrderRejectsMalformedRequests(t *testing.T) {
	secondLine := func() order.PlaceOrderLine {
		line := validCheckoutRequest().Lines[0]
		line.CartItemID = "10000000-0000-4000-8000-000000000002"
		line.SKUID = "20000000-0000-4000-8000-000000000002"
		return line
	}
	flashID := "30000000-0000-4000-8000-000000000001"
	cases := []struct {
		name   string
		mutate func(*order.PlaceOrderRequest)
		field  string
	}{
		{"empty lines", func(r *order.PlaceOrderRequest) { r.Lines = nil }, "lines"},
		{"too many lines", func(r *order.PlaceOrderRequest) { r.Lines = make([]order.PlaceOrderLine, 21) }, "lines"},
		{"invalid cart item id", func(r *order.PlaceOrderRequest) { r.Lines[0].CartItemID = "bad" }, "lines[0].cart_item_id"},
		{"duplicate cart item id", func(r *order.PlaceOrderRequest) {
			r.Lines = append(r.Lines, secondLine())
			r.Lines[1].CartItemID = r.Lines[0].CartItemID
		}, "lines[1].cart_item_id"},
		{"duplicate sku id", func(r *order.PlaceOrderRequest) {
			r.Lines = append(r.Lines, secondLine())
			r.Lines[1].SKUID = r.Lines[0].SKUID
		}, "lines[1].sku_id"},
		{"zero quantity", func(r *order.PlaceOrderRequest) { r.Lines[0].Quantity = 0 }, "lines[0].quantity"},
		{"zero cart version", func(r *order.PlaceOrderRequest) { r.Lines[0].CartItemVersion = 0 }, "lines[0].cart_item_version"},
		{"zero price", func(r *order.PlaceOrderRequest) { r.Lines[0].UnitPrice = 0 }, "lines[0].unit_price"},
		{"wrong subtotal", func(r *order.PlaceOrderRequest) { r.Lines[0].Subtotal = 201 }, "lines[0].subtotal"},
		{"subtotal overflow", func(r *order.PlaceOrderRequest) {
			r.Lines[0].Quantity = 2
			r.Lines[0].UnitPrice = math.MaxInt64
			r.Lines[0].Subtotal = 1
		}, "lines[0].subtotal"},
		{"flash source missing sale item", func(r *order.PlaceOrderRequest) { r.Lines[0].PriceSource = "flash_sale" }, "lines[0].flash_sale_item_id"},
		{"regular source with sale item", func(r *order.PlaceOrderRequest) { r.Lines[0].FlashSaleItemID = &flashID }, "lines[0].flash_sale_item_id"},
		{"unknown price source", func(r *order.PlaceOrderRequest) { r.Lines[0].PriceSource = "vip" }, "lines[0].price_source"},
		{"invalid sale item id", func(r *order.PlaceOrderRequest) {
			r.Lines[0].PriceSource = "flash_sale"
			invalid := "bad"
			r.Lines[0].FlashSaleItemID = &invalid
		}, "lines[0].flash_sale_item_id"},
		{"wrong currency", func(r *order.PlaceOrderRequest) { r.Currency = "USD" }, "currency"},
		{"wrong expected total", func(r *order.PlaceOrderRequest) { r.ExpectedTotal = 201 }, "expected_total"},
		{"blank recipient name", func(r *order.PlaceOrderRequest) { r.Recipient.Name = "  " }, "recipient.name"},
		{"recipient name too long", func(r *order.PlaceOrderRequest) { r.Recipient.Name = strings.Repeat("名", 121) }, "recipient.name"},
		{"recipient phone too short", func(r *order.PlaceOrderRequest) { r.Recipient.Phone = "1234567" }, "recipient.phone"},
		{"recipient address too long", func(r *order.PlaceOrderRequest) { r.Recipient.Address = strings.Repeat("道", 501) }, "recipient.address"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := validCheckoutRequest()
			tc.mutate(&request)
			fields := checkoutViolationFields(order.ValidatePlaceOrder(&request))
			if !fields[tc.field] {
				t.Fatalf("violation fields = %v, want %q", fields, tc.field)
			}
		})
	}
}

func TestValidatePlaceOrderNormalizesValidRequest(t *testing.T) {
	request := validCheckoutRequest()
	request.Lines[0].CartItemID = strings.ToUpper(request.Lines[0].CartItemID)
	request.Lines[0].SKUID = strings.ToUpper(request.Lines[0].SKUID)
	request.Recipient.Name = "  An  "
	request.Recipient.Phone = " 09012345 "
	request.Recipient.Address = "  12 Nguyen Trai  "

	if violations := order.ValidatePlaceOrder(&request); len(violations) != 0 {
		t.Fatalf("violations = %+v, want none", violations)
	}
	if request.Lines[0].CartItemID != "10000000-0000-4000-8000-000000000001" || request.Lines[0].SKUID != "20000000-0000-4000-8000-000000000001" {
		t.Errorf("UUIDs were not canonicalized: %+v", request.Lines[0])
	}
	if request.Recipient.Name != "An" || request.Recipient.Phone != "09012345" || request.Recipient.Address != "12 Nguyen Trai" {
		t.Errorf("recipient was not trimmed: %+v", request.Recipient)
	}
}

func TestValidateIdempotencyKeyLength(t *testing.T) {
	for _, tc := range []struct {
		length int
		valid  bool
	}{{7, false}, {8, true}, {255, true}, {256, false}} {
		t.Run(strings.Repeat("x", tc.length), func(t *testing.T) {
			got := order.ValidateIdempotencyKey(strings.Repeat("x", tc.length))
			if (len(got) == 0) != tc.valid {
				t.Fatalf("violations = %+v, valid = %v", got, tc.valid)
			}
			if len(got) > 0 && got[0].Field != "Idempotency-Key" {
				t.Errorf("field = %q, want Idempotency-Key", got[0].Field)
			}
		})
	}
}

func TestCanonicalRequestHash(t *testing.T) {
	base := validCheckoutRequest()
	base.Lines = append(base.Lines, order.PlaceOrderLine{
		CartItemID:      "10000000-0000-4000-8000-000000000002",
		CartItemVersion: 3,
		SKUID:           "20000000-0000-4000-8000-000000000002",
		Quantity:        1,
		UnitPrice:       50,
		Subtotal:        50,
		PriceSource:     "regular",
	})
	base.ExpectedTotal = 250
	if violations := order.ValidatePlaceOrder(&base); len(violations) != 0 {
		t.Fatalf("valid baseline request violations = %+v", violations)
	}
	want := order.CanonicalRequestHash(base)

	reordered := validCheckoutRequest()
	reordered.Lines = []order.PlaceOrderLine{base.Lines[1], base.Lines[0]}
	reordered.ExpectedTotal = base.ExpectedTotal
	reordered.Recipient.Name = " An "
	if violations := order.ValidatePlaceOrder(&reordered); len(violations) != 0 {
		t.Fatalf("reordered request violations = %+v", violations)
	}
	if got := order.CanonicalRequestHash(reordered); got != want {
		t.Errorf("reordered or whitespace-normalized request hash = %x, want %x", got, want)
	}

	cases := []struct {
		name   string
		mutate func(*order.PlaceOrderRequest)
	}{
		{"phone", func(r *order.PlaceOrderRequest) { r.Recipient.Phone = "09012346" }},
		{"cart version", func(r *order.PlaceOrderRequest) { r.Lines[0].CartItemVersion++ }},
		{"unit price", func(r *order.PlaceOrderRequest) {
			r.Lines[0].UnitPrice++
			r.Lines[0].Subtotal += 2
			r.ExpectedTotal += 2
		}},
		{"sale item", func(r *order.PlaceOrderRequest) {
			id := "30000000-0000-4000-8000-000000000001"
			r.Lines[0].PriceSource = "flash_sale"
			r.Lines[0].FlashSaleItemID = &id
		}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			changed := base
			changed.Lines = append([]order.PlaceOrderLine(nil), base.Lines...)
			changed.Recipient = base.Recipient
			tc.mutate(&changed)
			if violations := order.ValidatePlaceOrder(&changed); len(violations) != 0 {
				t.Fatalf("changed request violations = %+v", violations)
			}
			if got := order.CanonicalRequestHash(changed); got == want {
				t.Errorf("hash did not change for %s mutation", tc.name)
			}
		})
	}
}

func TestCheckoutProblemBuildersAndRawWriter(t *testing.T) {
	gin.SetMode(gin.TestMode)
	problemBody := problem.LineErrorsProblem("checkout rejected", []problem.LineProblem{
		{SKUID: "sku-1", Code: problem.CodeInsufficientStock},
		{SKUID: "sku-2", Code: problem.CodeItemUnavailable},
	})
	if problemBody.Status != 409 || problemBody.Code != problem.CodeItemUnavailable || len(problemBody.LineErrors) != 2 {
		t.Fatalf("line errors problem = %+v", problemBody)
	}
	if got := problem.CodeProblem(409, problem.CodeIdempotencyConflict, "different payload"); got.Status != 409 || got.Code != problem.CodeIdempotencyConflict {
		t.Errorf("code problem = %+v", got)
	}

	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	problem.WriteRaw(context, 201, "application/json", []byte(`{"id":"order-1"}`), map[string]string{"Idempotency-Replayed": "false"})
	if recorder.Code != 201 || recorder.Body.String() != `{"id":"order-1"}` || recorder.Header().Get("Idempotency-Replayed") != "false" || recorder.Header().Get("Content-Type") != "application/json" {
		t.Fatalf("raw response status=%d headers=%v body=%q", recorder.Code, recorder.Header(), recorder.Body.String())
	}
}
