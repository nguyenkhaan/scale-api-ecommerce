package middleware

import (
	"context"
	"errors"
	"net/http"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/module/auth"
	"github.com/gin-gonic/gin"
)

type SessionAuthenticator interface {
	Authenticate(ctx context.Context, sessionToken string) (auth.Identity, error)
}

// RequireSession resolves the session cookie and, for state-changing methods, also requires a valid CSRF token,
// so every cookie-authenticated mutation is protected without a separate opt-in.
func RequireSession(authenticator SessionAuthenticator) gin.HandlerFunc {
	return func(context *gin.Context) {
		token, err := context.Cookie(auth.SessionCookieName)
		if err != nil || token == "" {
			problem.Abort(context, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
			return
		}
		identity, err := authenticator.Authenticate(context.Request.Context(), token)
		if errors.Is(err, auth.ErrUnauthenticated) {
			problem.Abort(context, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
			return
		}
		if err != nil {
			problem.AbortInternal(context, err)
			return
		}
		if !isSafeMethod(context.Request.Method) && !identity.VerifyCSRF(context.GetHeader(auth.CSRFHeaderName)) {
			problem.Abort(context, http.StatusForbidden, problem.CodeForbidden, "The CSRF token is missing or invalid.")
			return
		}
		auth.SetIdentity(context, identity)
		context.Next()
	}
}

func isSafeMethod(method string) bool {
	return method == http.MethodGet || method == http.MethodHead || method == http.MethodOptions
}
