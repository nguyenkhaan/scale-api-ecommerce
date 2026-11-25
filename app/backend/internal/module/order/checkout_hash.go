package order

import (
	"crypto/sha256"
	"encoding/json"
	"sort"
)

type canonicalPlaceOrderRequest struct {
	Lines         []PlaceOrderLine `json:"lines"`
	Currency      string           `json:"currency"`
	ExpectedTotal int64            `json:"expected_total"`
	Recipient     RecipientRequest `json:"recipient"`
}

func CanonicalRequestHash(req PlaceOrderRequest) [32]byte {
	lines := append([]PlaceOrderLine(nil), req.Lines...)
	sort.Slice(lines, func(i, j int) bool { return lines[i].CartItemID < lines[j].CartItemID })
	canonical := canonicalPlaceOrderRequest{
		Lines: lines, Currency: req.Currency, ExpectedTotal: req.ExpectedTotal, Recipient: req.Recipient,
	}
	body, _ := json.Marshal(canonical)
	return sha256.Sum256(body)
}
