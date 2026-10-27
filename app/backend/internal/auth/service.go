package auth

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"
	"uuid"

	"cloudian/cloudian-scale/internal/model"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const (
	// DefaultPasswordCost is shared with the bootstrap admin seed so both kinds of account hash alike.
	DefaultPasswordCost = 12

	sessionLifetime         = 7 * 24 * time.Hour
	lastSeenRefreshInterval = 5 * time.Minute
)

var (
	ErrEmailTaken         = errors.New("email is already registered")
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUnauthenticated    = errors.New("session is missing, expired or revoked")
)

// Session carries the raw secrets of a newly issued session; only their hashes are stored.
type Session struct {
	User      model.User
	Token     string
	CSRFToken string
	ExpiresAt time.Time
}

type Service struct {
	database     *gorm.DB
	passwordCost int
	dummyHash    []byte
}

func NewService(database *gorm.DB, passwordCost int) (*Service, error) {
	if database == nil {
		return nil, errors.New("auth database is required")
	}
	dummyHash, err := bcrypt.GenerateFromPassword([]byte("salescale-timing-equalizer"), passwordCost)
	if err != nil {
		return nil, fmt.Errorf("prepare password hashing: %w", err)
	}
	return &Service{database: database, passwordCost: passwordCost, dummyHash: dummyHash}, nil
}

// Register always creates a customer; the role is never taken from the request.
func (service *Service) Register(ctx context.Context, email, password string) (Session, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(password), service.passwordCost)
	if err != nil {
		return Session{}, fmt.Errorf("hash password: %w", err)
	}
	user := model.User{
		ID:              uuid.NewV4().String(),
		Email:           strings.TrimSpace(email),
		EmailNormalized: normalizeEmail(email),
		PasswordHash:    string(passwordHash),
		Role:            model.RoleCustomer,
		Status:          model.UserStatusActive,
	}

	var session Session
	err = service.database.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if err := transaction.Create(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrDuplicatedKey) {
				return ErrEmailTaken
			}
			return fmt.Errorf("insert user: %w", err)
		}
		var err error
		session, err = createSession(transaction, user)
		return err
	})
	return session, err
}

func (service *Service) Login(ctx context.Context, email, password string) (Session, error) {
	database := service.database.WithContext(ctx)
	normalizedEmail := normalizeEmail(email)
	if err := reserveLoginAttempt(database, normalizedEmail); err != nil {
		return Session{}, err
	}

	var user model.User
	err := database.Where("email_normalized = ? AND status = ?", normalizedEmail, model.UserStatusActive).Take(&user).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		// Spend the same bcrypt time as a real check so response timing does not reveal registered emails.
		_ = bcrypt.CompareHashAndPassword(service.dummyHash, []byte(password))
		return Session{}, ErrInvalidCredentials
	}
	if err != nil {
		return Session{}, fmt.Errorf("find user: %w", err)
	}
	if bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)) != nil {
		return Session{}, ErrInvalidCredentials
	}

	var session Session
	err = database.Transaction(func(transaction *gorm.DB) error {
		if err := clearLoginAttempts(transaction, normalizedEmail); err != nil {
			return err
		}
		var err error
		session, err = createSession(transaction, user)
		return err
	})
	return session, err
}

// Authenticate resolves a session token; expiry is decided by database time, never the API clock.
func (service *Service) Authenticate(ctx context.Context, sessionToken string) (Identity, error) {
	if sessionToken == "" || len(sessionToken) > maxSecretLength {
		return Identity{}, ErrUnauthenticated
	}
	database := service.database.WithContext(ctx)
	var row struct {
		SessionID      string `gorm:"column:session_id"`
		CSRFSecretHash []byte `gorm:"column:csrf_secret_hash"`
		LastSeenStale  bool   `gorm:"column:last_seen_stale"`
		UserID         string `gorm:"column:user_id"`
		Email          string `gorm:"column:email"`
		Role           string `gorm:"column:role"`
	}
	result := database.Raw(
		`SELECT s.id AS session_id, s.csrf_secret_hash,
			s.last_seen_at < now() - make_interval(secs => ?) AS last_seen_stale,
			u.id AS user_id, u.email, u.role
		FROM sessions s
		JOIN users u ON u.id = s.user_id
		WHERE s.token_hash = ? AND s.revoked_at IS NULL AND s.expires_at > now() AND u.status = ?`,
		lastSeenRefreshInterval.Seconds(), hashSecret(sessionTokenDomain, sessionToken), model.UserStatusActive,
	).Scan(&row)
	if result.Error != nil {
		return Identity{}, fmt.Errorf("load session: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return Identity{}, ErrUnauthenticated
	}
	// Refreshing at most every few minutes avoids a database write on every authenticated request.
	if row.LastSeenStale {
		if err := database.Exec("UPDATE sessions SET last_seen_at = now() WHERE id = ?", row.SessionID).Error; err != nil {
			return Identity{}, fmt.Errorf("refresh session activity: %w", err)
		}
	}
	return Identity{
		UserID:         row.UserID,
		Email:          row.Email,
		Role:           row.Role,
		SessionID:      row.SessionID,
		csrfSecretHash: row.CSRFSecretHash,
	}, nil
}

func (service *Service) Logout(ctx context.Context, sessionID string) error {
	err := service.database.WithContext(ctx).
		Exec("UPDATE sessions SET revoked_at = now() WHERE id = ? AND revoked_at IS NULL", sessionID).Error
	if err != nil {
		return fmt.Errorf("revoke session: %w", err)
	}
	return nil
}

func createSession(database *gorm.DB, user model.User) (Session, error) {
	token, csrfToken := newSecret(), newSecret()
	var expiresAt time.Time
	err := database.Raw(
		`INSERT INTO sessions (id, user_id, token_hash, csrf_secret_hash, expires_at, last_seen_at, created_at)
		VALUES (?, ?, ?, ?, now() + make_interval(secs => ?), now(), now())
		RETURNING expires_at`,
		uuid.NewV4().String(), user.ID,
		hashSecret(sessionTokenDomain, token), hashSecret(csrfTokenDomain, csrfToken),
		sessionLifetime.Seconds(),
	).Scan(&expiresAt).Error
	if err != nil {
		return Session{}, fmt.Errorf("insert session: %w", err)
	}
	return Session{User: user, Token: token, CSRFToken: csrfToken, ExpiresAt: expiresAt}, nil
}
