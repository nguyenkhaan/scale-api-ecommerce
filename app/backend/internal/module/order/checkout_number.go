package order

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"
)

const orderNumberAlphabet = "0123456789ABCDEFGHJKMNPQRSTVWXYZ"

func newOrderNumber(now time.Time) (string, error) {
	var suffix [8]byte
	limit := big.NewInt(int64(len(orderNumberAlphabet)))
	for i := range suffix {
		value, err := rand.Int(rand.Reader, limit)
		if err != nil {
			return "", fmt.Errorf("generate order number: %w", err)
		}
		suffix[i] = orderNumberAlphabet[value.Int64()]
	}
	return "SS-" + now.UTC().Format("20060102") + "-" + string(suffix[:]), nil
}
