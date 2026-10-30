package migration

import (
	"math"
	"testing"

	"github.com/pressly/goose/v3"
)

func TestEmbeddedMigrationsIncludeFlashSaleSchema(t *testing.T) {
	if err := configureGoose(); err != nil {
		t.Fatalf("configure Goose: %v", err)
	}
	migrations, err := goose.CollectMigrations(migrationDir, 0, math.MaxInt64)
	if err != nil {
		t.Fatalf("collect embedded migrations: %v", err)
	}
	latest, err := migrations.Last()
	if err != nil {
		t.Fatalf("find latest migration: %v", err)
	}
	if latest.Version != 4 || latest.Source != "sql/00004_add_flash_sale_id_defaults.sql" {
		t.Errorf("latest migration = %d %q, want 4 %q", latest.Version, latest.Source, "sql/00004_add_flash_sale_id_defaults.sql")
	}
}
