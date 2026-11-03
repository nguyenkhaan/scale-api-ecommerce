package backend_test

import (
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/bootstrap"
)

func TestCatalogSeedRequiresExplicitEnvironmentGuard(t *testing.T) {
	testCases := []struct {
		name          string
		environment   string
		allowSeed     string
		wantErrorText string
	}{
		{name: "production rejected", environment: "production", allowSeed: "", wantErrorText: "restricted to development or test"},
		{name: "staging rejected", environment: "staging", allowSeed: "", wantErrorText: "restricted to development or test"},
		{name: "production with explicit allow", environment: "production", allowSeed: "true", wantErrorText: ""},
		{name: "development allowed", environment: "development", allowSeed: "", wantErrorText: ""},
		{name: "test allowed", environment: "test", allowSeed: "", wantErrorText: ""},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Setenv("ENV", tc.environment)
			t.Setenv("ALLOW_CATALOG_SEED", tc.allowSeed)

			err := bootstrap.CanSeedCatalog()
			if tc.wantErrorText != "" {
				if err == nil || !strings.Contains(err.Error(), tc.wantErrorText) {
					t.Fatalf("CanSeedCatalog() error = %v, want text %q", err, tc.wantErrorText)
				}
			} else if err != nil {
				t.Fatalf("CanSeedCatalog() unexpected error = %v", err)
			}
		})
	}
}

func TestCatalogSeedIntegration(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	if err := bootstrap.SeedCatalog(t.Context(), testDatabase.ORM()); err != nil {
		t.Fatalf("SeedCatalog() first run failed: %v", err)
	}

	// Idempotency: second run must not fail
	if err := bootstrap.SeedCatalog(t.Context(), testDatabase.ORM()); err != nil {
		t.Fatalf("SeedCatalog() second run failed: %v", err)
	}
}
