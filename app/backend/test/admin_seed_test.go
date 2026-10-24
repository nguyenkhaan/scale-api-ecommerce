package backend_test

import (
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/bootstrap"
)

func TestAdminSeedRequiresExplicitNonProductionGuard(t *testing.T) {
	testCases := []struct {
		name          string
		environment   string
		allowSeed     string
		wantErrorText string
	}{
		{name: "production", environment: "production", allowSeed: "true", wantErrorText: "development or test"},
		{name: "missing confirmation", environment: "development", allowSeed: "", wantErrorText: "ALLOW_ADMIN_SEED"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			t.Setenv("ENV", testCase.environment)
			t.Setenv("ALLOW_ADMIN_SEED", testCase.allowSeed)
			t.Setenv("ADMIN_EMAIL", "admin@example.com")
			t.Setenv("ADMIN_PASSWORD", "a-secure-development-password")

			_, err := bootstrap.LoadAdminSeedConfig()
			if err == nil || !strings.Contains(err.Error(), testCase.wantErrorText) {
				t.Fatalf("LoadAdminSeedConfig() error = %v, want text %q", err, testCase.wantErrorText)
			}
		})
	}
}

func TestAdminSeedNormalizesEmail(t *testing.T) {
	t.Setenv("ENV", "development")
	t.Setenv("ALLOW_ADMIN_SEED", "true")
	t.Setenv("ADMIN_EMAIL", "  Admin@Example.COM ")
	t.Setenv("ADMIN_PASSWORD", "a-secure-development-password")

	adminConfig, err := bootstrap.LoadAdminSeedConfig()
	if err != nil {
		t.Fatalf("LoadAdminSeedConfig() returned an unexpected error: %v", err)
	}
	if adminConfig.Email != "Admin@Example.COM" {
		t.Errorf("Email = %q, want display email preserved and trimmed", adminConfig.Email)
	}
	if adminConfig.NormalizedEmail != "admin@example.com" {
		t.Errorf("NormalizedEmail = %q, want %q", adminConfig.NormalizedEmail, "admin@example.com")
	}
}

func TestAdminSeedRejectsShortPassword(t *testing.T) {
	t.Setenv("ENV", "test")
	t.Setenv("ALLOW_ADMIN_SEED", "true")
	t.Setenv("ADMIN_EMAIL", "admin@example.com")
	t.Setenv("ADMIN_PASSWORD", "too-short")

	if _, err := bootstrap.LoadAdminSeedConfig(); err == nil {
		t.Fatal("LoadAdminSeedConfig() error = nil, want short-password error")
	}
}
