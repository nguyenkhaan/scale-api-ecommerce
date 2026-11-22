package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"strings"
	"time"
)

const (
	// SignatureHeader carries "t=<unix seconds>,v1=<hex HMAC-SHA256(secret, "<t>." + raw body)>".
	SignatureHeader = "X-Simulator-Signature"
	// SignatureTolerance bounds how far a signed timestamp may drift from server time.
	SignatureTolerance = 300 * time.Second
)

// Sign builds the signature header value for a raw body at the given time.
func Sign(secret string, at time.Time, body []byte) string {
	timestamp := strconv.FormatInt(at.Unix(), 10)
	return "t=" + timestamp + ",v1=" + hex.EncodeToString(mac(secret, timestamp, body))
}

func mac(secret, timestamp string, body []byte) []byte {
	hasher := hmac.New(sha256.New, []byte(secret))
	hasher.Write([]byte(timestamp + "."))
	hasher.Write(body)
	return hasher.Sum(nil)
}

// verifySignature checks the header over the exact raw body bytes. It fails closed on an empty secret,
// a malformed header, a timestamp outside the tolerance, or any digest mismatch (constant-time compare).
func verifySignature(secret, header string, body []byte, now time.Time) bool {
	if secret == "" {
		return false
	}
	var timestamp string
	var digests []string
	for _, part := range strings.Split(header, ",") {
		key, value, ok := strings.Cut(strings.TrimSpace(part), "=")
		switch {
		case !ok:
		case key == "t" && timestamp == "":
			timestamp = value
		case key == "v1":
			digests = append(digests, value)
		}
	}
	seconds, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil || len(digests) == 0 {
		return false
	}
	drift := now.Sub(time.Unix(seconds, 0))
	if drift > SignatureTolerance || drift < -SignatureTolerance {
		return false
	}
	expected := mac(secret, timestamp, body)
	matched := false
	for _, candidate := range digests {
		decoded, err := hex.DecodeString(candidate)
		if err == nil && hmac.Equal(decoded, expected) {
			matched = true
		}
	}
	return matched
}
