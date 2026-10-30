package auth

import (
	"errors"
	"net/http"
	"strconv"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"github.com/gin-gonic/gin"
)

type Controller struct {
	service       *Service
	secureCookies bool
}

func NewController(service *Service, secureCookies bool) *Controller {
	return &Controller{service: service, secureCookies: secureCookies}
}

// RegisterRoutes mounts /v1/auth; requireSession must resolve the session cookie and enforce CSRF on mutations.
func RegisterRoutes(router *gin.RouterGroup, controller *Controller, requireSession gin.HandlerFunc) {
	group := router.Group("/v1/auth", disableCaching)
	group.POST("/register", controller.Register)
	group.POST("/login", controller.Login)
	group.POST("/logout", requireSession, controller.Logout)
	group.GET("/me", requireSession, controller.Me)
}

func (controller *Controller) Register(context *gin.Context) {
	request, violations := decodeCredentials(context)
	if violations == nil {
		violations = request.validateRegistration()
	}
	if len(violations) > 0 {
		problem.AbortValidation(context, violations)
		return
	}

	session, err := controller.service.Register(context.Request.Context(), request.Email, request.Password)
	switch {
	case errors.Is(err, ErrEmailTaken):
		problem.Abort(context, http.StatusConflict, problem.CodeEmailAlreadyRegistered, "This email is already registered.")
	case err != nil:
		problem.AbortInternal(context, err)
	default:
		controller.writeSession(context, http.StatusCreated, session)
	}
}

func (controller *Controller) Login(context *gin.Context) {
	request, violations := decodeCredentials(context)
	if violations == nil {
		violations = request.validateLogin()
	}
	if len(violations) > 0 {
		problem.AbortValidation(context, violations)
		return
	}

	session, err := controller.service.Login(context.Request.Context(), request.Email, request.Password)
	var throttled *TooManyAttemptsError
	switch {
	case errors.As(err, &throttled):
		context.Header("Retry-After", strconv.Itoa(int(throttled.RetryAfter.Seconds())))
		problem.Abort(context, http.StatusTooManyRequests, problem.CodeRateLimited, "Too many failed login attempts. Try again later.")
	case errors.Is(err, ErrInvalidCredentials):
		problem.Abort(context, http.StatusUnauthorized, problem.CodeUnauthenticated, "Email or password is incorrect.")
	case err != nil:
		problem.AbortInternal(context, err)
	default:
		controller.writeSession(context, http.StatusOK, session)
	}
}

func (controller *Controller) Logout(context *gin.Context) {
	identity, ok := CurrentIdentity(context)
	if !ok {
		problem.Abort(context, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
		return
	}
	if err := controller.service.Logout(context.Request.Context(), identity.SessionID); err != nil {
		problem.AbortInternal(context, err)
		return
	}
	controller.setCookies(context, "", "", -1)
	context.Status(http.StatusNoContent)
}

func (controller *Controller) Me(context *gin.Context) {
	identity, ok := CurrentIdentity(context)
	if !ok {
		problem.Abort(context, http.StatusUnauthorized, problem.CodeUnauthenticated, "A valid session is required.")
		return
	}
	context.JSON(http.StatusOK, UserResponse{ID: identity.UserID, Email: identity.Email, Role: identity.Role})
}

func (controller *Controller) writeSession(context *gin.Context, status int, session Session) {
	controller.setCookies(context, session.Token, session.CSRFToken, int(sessionLifetime.Seconds()))
	context.JSON(status, newAuthResultResponse(session))
}

// setCookies writes both session cookies; a negative maxAge deletes them.
func (controller *Controller) setCookies(context *gin.Context, sessionToken, csrfToken string, maxAge int) {
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     SessionCookieName,
		Value:    sessionToken,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		Secure:   controller.secureCookies,
		SameSite: http.SameSiteLaxMode,
	})
	// Readable by the frontend so it can echo the token in X-CSRF-Token after a page reload.
	http.SetCookie(context.Writer, &http.Cookie{
		Name:     CSRFCookieName,
		Value:    csrfToken,
		Path:     "/",
		MaxAge:   maxAge,
		Secure:   controller.secureCookies,
		SameSite: http.SameSiteStrictMode,
	})
}

// disableCaching keeps responses carrying identity or CSRF tokens out of shared and browser caches.
func disableCaching(context *gin.Context) {
	context.Header("Cache-Control", "no-store")
	context.Next()
}
