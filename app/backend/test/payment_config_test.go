package backend_test

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/module/payment"
)

func TestPaymentLoadConfigValidatesSecretOutsideLocalEnvironments(t *testing.T) {
	cases := []struct {
		name, env, secret string
		wantErr           bool
	}{
		{"local placeholder", "dev", payment.DevPlaceholderSecret, false},
		{"test short secret", "test", "short", false},
		{"production placeholder", "production", payment.DevPlaceholderSecret, true},
		{"production short", "production", strings.Repeat("s", 31), true},
		{"unset env short", "", "short", true},
		{"production strong", "production", strings.Repeat("s", 32), false},
		{"production empty secret disables", "production", "", false},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			t.Setenv("ENV", test.env)
			t.Setenv("PAYMENT_SIMULATOR_SECRET", test.secret)
			cfg, err := payment.LoadConfig()
			if (err != nil) != test.wantErr {
				t.Fatalf("LoadConfig error = %v, wantErr %v", err, test.wantErr)
			}
			if err == nil && cfg.SimulatorSecret != test.secret {
				t.Errorf("secret was altered")
			}
		})
	}
}

func TestPaymentLoadConfigDefaults(t *testing.T) {
	t.Setenv("PAYMENT_SIMULATOR_ENABLED", "")
	t.Setenv("PAYMENT_CALLBACK_URL", "")
	t.Setenv("PAYMENT_SIMULATOR_SECRET", "")
	t.Setenv("PORT", "4100")
	t.Setenv("ENV", "development")
	cfg, err := payment.LoadConfig()
	if err != nil || !cfg.SimulatorEnabled || cfg.CallbackURL != "http://127.0.0.1:4100/api/payment-callbacks/simulator" {
		t.Errorf("dev config = %+v, %v", cfg, err)
	}
	t.Setenv("ENV", "production")
	t.Setenv("PAYMENT_SIMULATOR_SECRET", strings.Repeat("s", 40))
	if cfg, err = payment.LoadConfig(); err != nil || cfg.SimulatorEnabled {
		t.Errorf("production config = %+v, %v; simulator must default off", cfg, err)
	}
	t.Setenv("PAYMENT_SIMULATOR_ENABLED", "maybe")
	if _, err = payment.LoadConfig(); err == nil {
		t.Error("an invalid PAYMENT_SIMULATOR_ENABLED must be a startup error")
	}
}

func TestPaymentEnvExampleSecretIsTheRefusedPlaceholder(t *testing.T) {
	content, err := os.ReadFile("../.env.example")
	if err != nil {
		t.Fatalf("read .env.example: %v", err)
	}
	if !strings.Contains(string(content), "PAYMENT_SIMULATOR_SECRET="+payment.DevPlaceholderSecret+"\n") {
		t.Error(".env.example secret drifted from payment.DevPlaceholderSecret, so it would be accepted in production")
	}
}

func TestPaymentEmptySecretMountsNoCallbackOrSimulatorRoutes(t *testing.T) {
	env := newPaymentEnv(t, func(cfg *payment.Config) { cfg.SimulatorSecret = "" })
	for _, route := range []struct{ method, path string }{
		{http.MethodPost, paymentCallbackPath},
		{http.MethodPost, "/api/payment-simulator/events"},
		{http.MethodGet, "/api/payment-simulator/sessions"},
	} {
		request := httptest.NewRequest(route.method, route.path, strings.NewReader("{}"))
		request.Header.Set("X-Operator-Key", paymentTestOperatorKey)
		response := httptest.NewRecorder()
		env.router.ServeHTTP(response, request)
		if response.Code != http.StatusNotFound {
			t.Errorf("%s %s = %d, want 404 (not mounted)", route.method, route.path, response.Code)
		}
	}
	if session := env.newFixture(t, 300).createSession(t); session.ID == "" {
		t.Error("payment sessions must still work without the simulator secret")
	}
}
