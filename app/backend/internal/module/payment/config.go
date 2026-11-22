package payment

import (
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"cloudian/cloudian-scale/internal/config"
)

// Config holds the payment simulator settings. All of them come from the environment.
type Config struct {
	// SimulatorEnabled exposes the simulator control and admin listing endpoints; disabled means 404.
	SimulatorEnabled bool
	// SimulatorSecret keys the HMAC that signs and verifies callbacks. Empty rejects every callback.
	SimulatorSecret string
	// OperatorKey lets automation call simulator control with X-Operator-Key. Empty disables that access.
	OperatorKey string
	// CallbackURL is where the simulator delivers events, so the real verification path always runs.
	CallbackURL string
}

// LoadConfig reads PAYMENT_* variables. The simulator defaults to enabled only on local environments.
func LoadConfig() (Config, error) {
	port := strings.TrimSpace(os.Getenv("PORT"))
	if port == "" {
		port = "4000"
	}
	cfg := Config{
		SimulatorEnabled: config.IsLocalEnvironment(os.Getenv("ENV")),
		SimulatorSecret:  os.Getenv("PAYMENT_SIMULATOR_SECRET"),
		OperatorKey:      os.Getenv("PAYMENT_OPERATOR_KEY"),
		CallbackURL:      strings.TrimSpace(os.Getenv("PAYMENT_CALLBACK_URL")),
	}
	if raw := strings.TrimSpace(os.Getenv("PAYMENT_SIMULATOR_ENABLED")); raw != "" {
		enabled, err := strconv.ParseBool(raw)
		if err != nil {
			return Config{}, fmt.Errorf("PAYMENT_SIMULATOR_ENABLED must be a boolean: %w", err)
		}
		cfg.SimulatorEnabled = enabled
	}
	if cfg.CallbackURL == "" {
		cfg.CallbackURL = "http://" + net.JoinHostPort("127.0.0.1", port) + "/api/payment-callbacks/simulator"
	}
	if err := cfg.validate(os.Getenv("ENV")); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// DevPlaceholderSecret is the sample value shipped in .env.example. It is public, so it is refused outside
// local environments.
const DevPlaceholderSecret = "dev-only-simulator-secret-change-me"

// minProductionSecretBytes is the shortest HMAC secret accepted outside local environments.
const minProductionSecretBytes = 32

// validate rejects a weak secret outside local environments. An empty secret is allowed: it switches the
// callback and simulator off instead of running them with a guessable key.
func (cfg Config) validate(environment string) error {
	if cfg.SimulatorSecret == "" || config.IsLocalEnvironment(environment) {
		return nil
	}
	if cfg.SimulatorSecret == DevPlaceholderSecret {
		return errors.New("PAYMENT_SIMULATOR_SECRET is the .env.example placeholder; set a private random secret")
	}
	if len(cfg.SimulatorSecret) < minProductionSecretBytes {
		return fmt.Errorf("PAYMENT_SIMULATOR_SECRET must be at least %d bytes outside local environments", minProductionSecretBytes)
	}
	return nil
}
