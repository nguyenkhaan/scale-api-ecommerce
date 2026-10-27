package backend_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
	"testing"

	"cloudian/cloudian-scale/internal/httpserver/problem"
)

func TestLoginIsThrottledAcrossInstancesAfterRepeatedFailures(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	instanceA, instanceB := newAuthRouter(t, testDatabase.ORM()), newAuthRouter(t, testDatabase.ORM())
	newTestClient(instanceA).send(t, http.MethodPost, "/api/v1/auth/register", credentials("lan@example.com", testPassword))

	for attempt := 1; attempt <= 5; attempt++ {
		response := newTestClient(instanceA).send(t, http.MethodPost, "/api/v1/auth/login", credentials("lan@example.com", "wrong password"))
		requireProblem(t, response, http.StatusUnauthorized, problem.CodeUnauthenticated)
	}

	// Even the correct password is refused on another instance while the window is open.
	blocked := newTestClient(instanceB).send(t, http.MethodPost, "/api/v1/auth/login", credentials("lan@example.com", testPassword))
	requireProblem(t, blocked, http.StatusTooManyRequests, problem.CodeRateLimited)
	if retryAfter, err := strconv.Atoi(blocked.Header().Get("Retry-After")); err != nil || retryAfter < 1 || retryAfter > 900 {
		t.Errorf("Retry-After = %q, want 1..900 seconds", blocked.Header().Get("Retry-After"))
	}

	testDatabase.ORM().Exec("UPDATE login_throttles SET window_started_at = now() - interval '16 minutes'")
	if response := newTestClient(instanceB).send(t, http.MethodPost, "/api/v1/auth/login", credentials("lan@example.com", testPassword)); response.Code != http.StatusOK {
		t.Fatalf("login after window = %d, want 200; body = %s", response.Code, response.Body.String())
	}
	var remaining int64
	testDatabase.ORM().Raw("SELECT count(*) FROM login_throttles").Scan(&remaining)
	if remaining != 0 {
		t.Errorf("login_throttles rows = %d, want cleared after success", remaining)
	}
}

// A burst of parallel guesses must not get more password checks than sequential attempts would.
func TestLoginThrottleHoldsUnderConcurrentGuesses(t *testing.T) {
	testDatabase := openIntegrationDatabase(t)
	instanceA, instanceB := newAuthRouter(t, testDatabase.ORM()), newAuthRouter(t, testDatabase.ORM())
	newTestClient(instanceA).send(t, http.MethodPost, "/api/v1/auth/register", credentials("lan@example.com", testPassword))

	const guesses = 20
	statuses := make(chan int, guesses)
	var waitGroup sync.WaitGroup
	for guess := 0; guess < guesses; guess++ {
		router := instanceA
		if guess%2 == 1 {
			router = instanceB
		}
		waitGroup.Go(func() {
			request := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login",
				strings.NewReader(`{"email":"lan@example.com","password":"wrong password"}`))
			request.Header.Set("Content-Type", "application/json")
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)
			statuses <- response.Code
		})
	}
	waitGroup.Wait()
	close(statuses)

	counts := map[int]int{}
	for status := range statuses {
		counts[status]++
	}
	if counts[http.StatusUnauthorized] > 5 || counts[http.StatusUnauthorized]+counts[http.StatusTooManyRequests] != guesses {
		t.Errorf("status counts = %v, want at most 5 x 401 and the rest 429", counts)
	}
}
