//go:build integration

package backend_test

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"
)

func TestSeedCommandLoadsDevelopmentFixture(t *testing.T) {
	if os.Getenv("RUN_SEED_INTEGRATION") != "true" {
		t.Skip("set RUN_SEED_INTEGRATION=true to seed the DATABASE_URL configured in ../.env")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "go", "run", "./cmd/seed")
	command.Dir = ".."
	command.Env = append(os.Environ(), "ALLOW_SEED=true")
	output, err := command.CombinedOutput()
	if err != nil {
		t.Fatalf("run unified seed command: %v\n%s", err, output)
	}
}
