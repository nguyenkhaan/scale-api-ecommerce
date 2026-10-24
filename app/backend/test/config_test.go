package backend_test

import (
	"testing"

	"cloudian/cloudian-scale/internal/config"
)

func TestLoadConfigUsesDefaultPort(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgresql://user:password@example.test/database")
	t.Setenv("PORT", "")
	t.Setenv("ENV", "test")

	applicationConfig, err := config.Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error: %v", err)
	}
	if applicationConfig.Port != "4000" {
		t.Errorf("Port = %q, want %q", applicationConfig.Port, "4000")
	}
}

func TestLoadConfigTrimsEnvironmentValues(t *testing.T) {
	t.Setenv("DATABASE_URL", "  postgresql://user:password@example.test/database  ")
	t.Setenv("PORT", " 8080 ")
	t.Setenv("ENV", " development ")

	applicationConfig, err := config.Load()
	if err != nil {
		t.Fatalf("Load() returned an unexpected error: %v", err)
	}
	if applicationConfig.DatabaseUrl != "postgresql://user:password@example.test/database" {
		t.Error("DatabaseUrl was not trimmed")
	}
	if applicationConfig.Port != "8080" {
		t.Errorf("Port = %q, want %q", applicationConfig.Port, "8080")
	}
	if applicationConfig.Env != "development" {
		t.Errorf("Env = %q, want %q", applicationConfig.Env, "development")
	}
}

func TestLoadConfigRejectsMissingDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")
	t.Setenv("PORT", "4000")
	t.Setenv("ENV", "test")

	if _, err := config.Load(); err == nil {
		t.Fatal("Load() error = nil, want an error when DATABASE_URL is empty")
	}
}
