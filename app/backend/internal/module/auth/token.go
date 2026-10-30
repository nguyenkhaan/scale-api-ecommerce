package auth

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

const (
	sessionTokenDomain = "salescale/session-token/v1"
	csrfTokenDomain    = "salescale/csrf-token/v1"
	secretBytes        = 32
	maxSecretLength    = 128
)

func newSecret() string {
	buffer := make([]byte, secretBytes)
	// crypto/rand.Read never returns an error since Go 1.24.
	rand.Read(buffer)
	return base64.RawURLEncoding.EncodeToString(buffer)
}

// hashSecret keys the hash with a purpose label so a session token can never verify as a CSRF token.
func hashSecret(domain, secret string) []byte {
	mac := hmac.New(sha256.New, []byte(domain))
	mac.Write([]byte(secret))
	return mac.Sum(nil)
}
