// Command expiry-worker expires pending orders whose payment deadline has passed.
// PostgreSQL time is the only clock; a restart simply continues with whatever is still due.
package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"cloudian/cloudian-scale/internal/config"
	"cloudian/cloudian-scale/internal/database"
	"cloudian/cloudian-scale/internal/module/order"
)

const (
	defaultInterval  = 2 * time.Second
	defaultBatchSize = 100
)

func main() {
	if err := run(); err != nil {
		slog.Error("expiry worker stopped", "error", err)
		os.Exit(1)
	}
}

func run() error {
	applicationConfig, err := config.Load()
	if err != nil {
		return err
	}
	interval, batchSize, err := loadWorkerSettings()
	if err != nil {
		return err
	}
	stop, cancel := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer cancel()

	startup, cancelStartup := context.WithTimeout(stop, 15*time.Second)
	postgres, err := database.Open(startup, applicationConfig.DatabaseUrl)
	cancelStartup()
	if err != nil {
		return err
	}
	defer postgres.Close()

	slog.Info("expiry worker started", "interval", interval.String(), "batch_size", batchSize)
	for {
		stats, err := order.ExpireDue(stop, postgres.ORM(), batchSize)
		if stop.Err() != nil {
			slog.Info("expiry worker stopping")
			return nil
		}
		logCycle(stats, err)
		// Keep draining without waiting while a pass made progress and more orders are still due.
		if err == nil && stats.Remaining > 0 && stats.Expired > 0 {
			continue
		}
		select {
		case <-stop.Done():
			slog.Info("expiry worker stopping")
			return nil
		case <-time.After(interval):
		}
	}
}

// logCycle writes one structured line, only when a cycle did something or failed.
func logCycle(stats order.ExpiryStats, err error) {
	if err == nil && stats.Scanned == 0 && stats.Errors == 0 {
		return
	}
	attrs := []any{
		"scanned", stats.Scanned, "expired", stats.Expired, "lost_race", stats.LostRace,
		"errors", stats.Errors, "remaining", stats.Remaining, "max_lateness_seconds", stats.MaxLatenessSeconds,
	}
	if err != nil {
		slog.Error("expiry cycle failed", append(attrs, "error", err)...)
		return
	}
	slog.Info("expiry cycle", attrs...)
}

func loadWorkerSettings() (time.Duration, int, error) {
	interval := defaultInterval
	if raw := strings.TrimSpace(os.Getenv("EXPIRY_WORKER_INTERVAL")); raw != "" {
		parsed, err := time.ParseDuration(raw)
		if err != nil || parsed <= 0 {
			return 0, 0, fmt.Errorf("EXPIRY_WORKER_INTERVAL must be a positive duration, got %q", raw)
		}
		interval = parsed
	}
	batchSize := defaultBatchSize
	if raw := strings.TrimSpace(os.Getenv("EXPIRY_WORKER_BATCH_SIZE")); raw != "" {
		parsed, err := strconv.Atoi(raw)
		if err != nil || parsed < 1 {
			return 0, 0, fmt.Errorf("EXPIRY_WORKER_BATCH_SIZE must be a positive integer, got %q", raw)
		}
		batchSize = parsed
	}
	return interval, batchSize, nil
}
