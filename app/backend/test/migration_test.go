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

func TestMigrationRepairsMissingPaymentSessionsTable(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	database := testDatabase.ORM()
	if err := database.Exec("DROP TABLE payment_sessions CASCADE").Error; err != nil {
		t.Fatalf("simulate missing payment_sessions table in isolated test schema: %v", err)
	}
	if err := database.Exec("DELETE FROM goose_db_version WHERE version_id IN (11, 12)").Error; err != nil {
		t.Fatalf("make repair migration eligible in isolated test schema: %v", err)
	}
	if err := migration.Run(context.Background(), testDatabase.SQL(), migration.CommandUp); err != nil {
		t.Fatalf("apply repair migration: %v", err)
	}

	var tableCount int
	if err := database.Raw(`SELECT count(*) FROM information_schema.tables WHERE table_schema = current_schema() AND table_name = 'payment_sessions'`).Scan(&tableCount).Error; err != nil {
		t.Fatalf("check repaired table: %v", err)
	}
	if tableCount != 1 {
		t.Fatalf("payment_sessions table count = %d, want 1", tableCount)
	}

}

func TestMigrationRepairsMissingPaymentEventTables(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	database := testDatabase.ORM()
	if err := database.Exec("DROP TABLE payment_reconciliations, payment_events CASCADE").Error; err != nil {
		t.Fatalf("simulate missing payment event tables in isolated test schema: %v", err)
	}
	if err := database.Exec("DELETE FROM goose_db_version WHERE version_id = 12").Error; err != nil {
		t.Fatalf("make event-table repair migration eligible in isolated test schema: %v", err)
	}
	if err := migration.Run(context.Background(), testDatabase.SQL(), migration.CommandUp); err != nil {
		t.Fatalf("apply event-table repair migration: %v", err)
	}

	var repairedTableCount int
	if err := database.Raw(`SELECT count(*) FROM information_schema.tables
		WHERE table_schema = current_schema() AND table_name IN ('payment_events', 'payment_reconciliations')`).
		Scan(&repairedTableCount).Error; err != nil {
		t.Fatalf("check repaired payment event tables: %v", err)
	}
	if repairedTableCount != 2 {
		t.Fatalf("payment event table count = %d, want 2", repairedTableCount)
	}

	var repairedReferences int
	if err := database.Raw(`SELECT count(*) FROM pg_constraint
		WHERE contype = 'f' AND conrelid = to_regclass('payment_reconciliations')
		AND confrelid IN (to_regclass('payment_events'), to_regclass('orders'))`).
		Scan(&repairedReferences).Error; err != nil {
		t.Fatalf("check repaired reconciliation references: %v", err)
	}
	if repairedReferences != 2 {
		t.Fatalf("payment reconciliation foreign key count = %d, want 2", repairedReferences)
	}
}
