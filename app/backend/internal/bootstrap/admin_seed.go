package bootstrap

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"os"
	"strings"
	"unicode/utf8"
	"uuid"

	"cloudian/cloudian-scale/internal/auth"
	"golang.org/x/crypto/bcrypt"
)

const adminPasswordMinLength = 12

type AdminSeedConfig struct {
	Email           string
	NormalizedEmail string
	password        string
}

func LoadAdminSeedConfig() (AdminSeedConfig, error) {
	environment := strings.ToLower(strings.TrimSpace(os.Getenv("ENV")))
	if environment != "development" && environment != "dev" && environment != "test" {
		return AdminSeedConfig{}, errors.New("admin seed is limited to development or test")
	}
	if !strings.EqualFold(strings.TrimSpace(os.Getenv("ALLOW_ADMIN_SEED")), "true") {
		return AdminSeedConfig{}, errors.New("admin seed requires ALLOW_ADMIN_SEED=true")
	}

	email := strings.TrimSpace(os.Getenv("ADMIN_EMAIL"))
	if err := validateEmail(email); err != nil {
		return AdminSeedConfig{}, err
	}
	password := os.Getenv("ADMIN_PASSWORD")
	if utf8.RuneCountInString(password) < adminPasswordMinLength {
		return AdminSeedConfig{}, fmt.Errorf("ADMIN_PASSWORD must contain at least %d characters", adminPasswordMinLength)
	}

	return AdminSeedConfig{
		Email:           email,
		NormalizedEmail: strings.ToLower(email),
		password:        password,
	}, nil
}

func SeedAdmin(ctx context.Context, database *sql.DB, adminConfig AdminSeedConfig) error {
	if database == nil {
		return errors.New("seed database is required")
	}
	passwordHash, err := bcrypt.GenerateFromPassword(
		[]byte(adminConfig.password),
		auth.DefaultPasswordCost,
	)
	if err != nil {
		return fmt.Errorf("hash admin password: %w", err)
	}

	result, err := database.ExecContext(
		ctx,
		`INSERT INTO users (
			id, email, email_normalized, password_hash, role, status
		) VALUES ($1, $2, $3, $4, 'admin', 'active')
		ON CONFLICT (email_normalized) DO NOTHING`,
		uuid.NewV4().String(),
		adminConfig.Email,
		adminConfig.NormalizedEmail,
		string(passwordHash),
	)
	if err != nil {
		return fmt.Errorf("insert bootstrap admin: %w", err)
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("read bootstrap admin result: %w", err)
	}
	if rowsAffected == 1 {
		return nil
	}

	var existingRole string
	err = database.QueryRowContext(
		ctx,
		"SELECT role FROM users WHERE email_normalized = $1",
		adminConfig.NormalizedEmail,
	).Scan(&existingRole)
	if err != nil {
		return fmt.Errorf("verify existing bootstrap admin: %w", err)
	}
	if existingRole != "admin" {
		return errors.New("seed email already belongs to a non-admin user")
	}
	return nil
}

func validateEmail(email string) error {
	address, err := mail.ParseAddress(email)
	if err != nil || address.Name != "" || address.Address != email {
		return errors.New("ADMIN_EMAIL must be a valid email address")
	}
	return nil
}
