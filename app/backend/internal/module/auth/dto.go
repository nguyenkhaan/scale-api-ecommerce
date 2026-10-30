package auth

import (
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"

	"cloudian/cloudian-scale/internal/httpserver/problem"
	"cloudian/cloudian-scale/internal/model"
	"github.com/gin-gonic/gin"
)

const (
	requestBodyLimit       = 8 << 10
	emailMaxLength         = 254
	passwordMinLength      = 8
	passwordMaxBytes       = 72 // bcrypt ignores input past 72 bytes, so longer passwords are rejected
	loginPasswordMaxLength = 128
)

// CredentialsRequest is the body of both register and login; only the password rules differ.
type CredentialsRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Role  string `json:"role"`
}

type AuthResultResponse struct {
	User      UserResponse `json:"user"`
	ExpiresAt string       `json:"expires_at"`
	CSRFToken string       `json:"csrf_token"`
}

func decodeCredentials(context *gin.Context) (CredentialsRequest, []problem.Violation) {
	var request CredentialsRequest
	// Requiring JSON forces cross-site callers through a CORS preflight the API never grants, which blocks
	// login CSRF through text/plain HTML forms whose body happens to parse as JSON.
	mediaType, _, err := mime.ParseMediaType(context.GetHeader("Content-Type"))
	if err != nil || mediaType != "application/json" {
		return request, []problem.Violation{{Field: "Content-Type", Reason: "must be application/json"}}
	}
	decoder := json.NewDecoder(http.MaxBytesReader(context.Writer, context.Request.Body, requestBodyLimit))
	decoder.DisallowUnknownFields()
	invalidBody := []problem.Violation{{Field: "body", Reason: "must be a JSON object with only email and password"}}
	if err := decoder.Decode(&request); err != nil {
		return request, invalidBody
	}
	if _, err := decoder.Token(); !errors.Is(err, io.EOF) {
		return request, invalidBody
	}
	return request, nil
}

func (request CredentialsRequest) validateRegistration() []problem.Violation {
	violations := validateEmail(request.Email)
	switch {
	case utf8.RuneCountInString(request.Password) < passwordMinLength:
		violations = append(violations, problem.Violation{Field: "password", Reason: "must contain at least 8 characters"})
	case len(request.Password) > passwordMaxBytes:
		violations = append(violations, problem.Violation{Field: "password", Reason: "must be at most 72 bytes"})
	}
	return violations
}

func (request CredentialsRequest) validateLogin() []problem.Violation {
	violations := validateEmail(request.Email)
	switch {
	case request.Password == "":
		violations = append(violations, problem.Violation{Field: "password", Reason: "is required"})
	case utf8.RuneCountInString(request.Password) > loginPasswordMaxLength:
		violations = append(violations, problem.Violation{Field: "password", Reason: "must be at most 128 characters"})
	}
	return violations
}

func validateEmail(email string) []problem.Violation {
	trimmed := strings.TrimSpace(email)
	switch {
	case trimmed == "":
		return []problem.Violation{{Field: "email", Reason: "is required"}}
	case len(trimmed) > emailMaxLength:
		return []problem.Violation{{Field: "email", Reason: "must be at most 254 characters"}}
	case !isPlainASCIIAddress(trimmed):
		return []problem.Violation{{Field: "email", Reason: "must be a valid email address"}}
	}
	return nil
}

// isPlainASCIIAddress keeps Go's and PostgreSQL's lower-casing in agreement for the normalized email.
func isPlainASCIIAddress(email string) bool {
	for index := 0; index < len(email); index++ {
		if email[index] >= utf8.RuneSelf {
			return false
		}
	}
	address, err := mail.ParseAddress(email)
	return err == nil && address.Name == "" && address.Address == email
}

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func newUserResponse(user model.User) UserResponse {
	return UserResponse{ID: user.ID, Email: user.Email, Role: user.Role}
}

func newAuthResultResponse(session Session) AuthResultResponse {
	return AuthResultResponse{
		User:      newUserResponse(session.User),
		ExpiresAt: session.ExpiresAt.UTC().Format(time.RFC3339),
		CSRFToken: session.CSRFToken,
	}
}
