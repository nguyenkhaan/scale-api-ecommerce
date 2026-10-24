package backend_test

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"cloudian/cloudian-scale/internal/database"
)

func TestNeonDatabaseConnection(t *testing.T) {
	connectionString := strings.TrimSpace(os.Getenv("TEST_DATABASE_URL"))
	if connectionString == "" {
		t.Skip("set TEST_DATABASE_URL to run the Neon integration test")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	postgresDatabase, err := database.Open(ctx, connectionString)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() {
		if err := postgresDatabase.Close(); err != nil {
			t.Errorf("close test database: %v", err)
		}
	})

	if err := postgresDatabase.Ping(ctx); err != nil {
		t.Fatalf("ping test database: %v", err)
	}
}
