package backend_test

import (
	"context"
	"testing"

	"cloudian/cloudian-scale/internal/migration"
)

func TestMigrationRejectsUnknownCommand(t *testing.T) {
	err := migration.Run(context.Background(), nil, "drop-everything")
	if err == nil {
		t.Fatal("Run() error = nil, want unsupported-command error")
	}
}

func TestDatabaseResetRequiresTestEnvironmentAndConfirmation(t *testing.T) {
	testCases := []struct {
		name         string
		environment  string
		confirmation string
		wantAllowed  bool
	}{
		{name: "test confirmed", environment: "test", confirmation: "true", wantAllowed: true},
		{name: "development rejected", environment: "development", confirmation: "true"},
		{name: "missing confirmation", environment: "test", confirmation: ""},
		{name: "production rejected", environment: "production", confirmation: "true"},
	}

	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			allowed := migration.CanResetDatabase(testCase.environment, testCase.confirmation)
			if allowed != testCase.wantAllowed {
				t.Errorf("CanResetDatabase() = %t, want %t", allowed, testCase.wantAllowed)
			}
		})
	}
}
