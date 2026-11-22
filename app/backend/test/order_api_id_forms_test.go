package backend_test

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"testing"
)

// uuid.Parse accepts urn:uuid: forms that PostgreSQL rejects; they must never turn into a 500.
func TestOrderIDInURNFormIsHandledNotRejectedByDatabase(t *testing.T) {
	database := openIntegrationDatabase(t).ORM()
	router := orderAPINewRouter(t, database)
	owner := cartNewCustomer(t, router, database, "urn-owner")
	other := cartNewCustomer(t, router, database, "urn-other")
	orderID, _ := orderAPISeedOrder(t, database, owner.ID, 300, 1)
	urnID := "urn:uuid:" + orderID

	requireProblem(t, other.send(t, http.MethodGet, "/api/orders/"+urnID, nil), http.StatusNotFound, "RESOURCE_NOT_FOUND")
	requireProblem(t, other.send(t, http.MethodPost, "/api/orders/"+urnID+"/cancel", nil, other.csrfHeader()...),
		http.StatusNotFound, "RESOURCE_NOT_FOUND")
	requireProblem(t, owner.send(t, http.MethodGet, "/api/orders/urn:uuid:11111111-2222-4333-8444-555555555555", nil),
		http.StatusNotFound, "RESOURCE_NOT_FOUND")

	// For the owner the alias resolves to the same order through its canonical form.
	if got := orderAPIGet(t, owner.testClient, "/api/orders/"+urnID); got.ID != orderID {
		t.Errorf("urn form resolved to %s, want %s", got.ID, orderID)
	}

	rawCursor := fmt.Sprintf(`{"c":"2099-01-01T00:00:00Z","i":%q,"s":""}`, urnID)
	cursor := base64.RawURLEncoding.EncodeToString([]byte(rawCursor))
	response := owner.send(t, http.MethodGet, "/api/orders?cursor="+cursor, nil)
	if response.Code != http.StatusOK {
		t.Errorf("urn-form cursor status = %d; body = %s, want 200 (canonicalized), never 500", response.Code, response.Body.String())
	}
	lifecycleRequireInvariants(t, database)
}
