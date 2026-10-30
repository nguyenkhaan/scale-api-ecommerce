package auth

import (
	"crypto/hmac"

	"github.com/gin-gonic/gin"
)

const (
	SessionCookieName = "salescale_session"
	CSRFCookieName    = "salescale_csrf"
	CSRFHeaderName    = "X-CSRF-Token"

	identityContextKey = "auth.identity"
)

// Identity is the authenticated caller resolved from a server-side session.
type Identity struct {
	UserID         string
	Email          string
	Role           string
	SessionID      string
	csrfSecretHash []byte
}

func (identity Identity) VerifyCSRF(token string) bool {
	if token == "" || len(token) > maxSecretLength || len(identity.csrfSecretHash) == 0 {
		return false
	}
	return hmac.Equal(hashSecret(csrfTokenDomain, token), identity.csrfSecretHash)
}

func SetIdentity(context *gin.Context, identity Identity) {
	context.Set(identityContextKey, identity)
}

func CurrentIdentity(context *gin.Context) (Identity, bool) {
	value, exists := context.Get(identityContextKey)
	if !exists {
		return Identity{}, false
	}
	identity, ok := value.(Identity)
	return identity, ok
}
