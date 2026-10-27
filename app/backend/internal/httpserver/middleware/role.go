package middleware

import (
	"net/http"

	"cloudian/cloudian-scale/internal/auth"
	"cloudian/cloudian-scale/internal/httpserver/problem"
	"github.com/gin-gonic/gin"
)

// RequireRole must run after RequireSession; it checks the role stored server-side, never one sent by the client.
func RequireRole(role string) gin.HandlerFunc {
	return func(context *gin.Context) {
		identity, ok := auth.CurrentIdentity(context)
		if !ok {
			problem.Abort(context, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
			return
		}
		if identity.Role != role {
			problem.Abort(context, http.StatusForbidden, problem.CodeForbidden, "This account cannot access this resource.")
			return
		}
		context.Next()
	}
}
