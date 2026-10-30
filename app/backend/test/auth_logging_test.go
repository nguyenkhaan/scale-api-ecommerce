package backend_test

import (
	"bytes"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"testing"

	"cloudian/cloudian-scale/internal/module/auth"
	"github.com/gin-gonic/gin"
)

// Covers every log sink the API uses: the Gin request log, GORM's query log on stdout and the standard logger.
func TestAuthLogsContainNoSecretsOrEmails(t *testing.T) {
	readStdout := captureStdout(t)
	var logs bytes.Buffer
	previousGinWriter, previousLogWriter := gin.DefaultWriter, log.Writer()
	gin.DefaultWriter = &logs
	log.SetOutput(&logs)
	t.Cleanup(func() {
		gin.DefaultWriter = previousGinWriter
		log.SetOutput(previousLogWriter)
	})

	// The database is opened after stdout is captured because GORM binds its logger to stdout at open time.
	testDatabase := openIntegrationDatabase(t)
	router := newAuthRouter(t, testDatabase.ORM())
	client := newTestClient(router)
	client.send(t, http.MethodPost, "/api/v1/auth/register", credentials("lan@example.com", testPassword))
	client.send(t, http.MethodPost, "/api/v1/auth/login", credentials("lan@example.com", testPassword))
	// A duplicate insert is logged by GORM as a failed statement; an unknown email is a missing row.
	newTestClient(router).send(t, http.MethodPost, "/api/v1/auth/register", credentials("lan@example.com", testPassword))
	newTestClient(router).send(t, http.MethodPost, "/api/v1/auth/login", credentials("nobody@example.com", testPassword))

	var passwordHash string
	testDatabase.ORM().Raw("SELECT password_hash FROM users").Scan(&passwordHash)
	output := logs.String() + readStdout()

	if !strings.Contains(output, "/api/v1/auth/login") || !strings.Contains(output, `INSERT INTO "users"`) {
		t.Fatalf("request and query logs were not both captured: %q", output)
	}
	secrets := []string{
		testPassword, passwordHash, "lan@example.com", "nobody@example.com",
		client.cookies[auth.SessionCookieName], client.cookies[auth.CSRFCookieName],
	}
	for _, secret := range secrets {
		if secret != "" && strings.Contains(output, secret) {
			t.Errorf("logs contain %q: %q", secret, output)
		}
	}
}

// captureStdout redirects os.Stdout until the returned function is called, which restores it and returns the output.
func captureStdout(t *testing.T) func() string {
	t.Helper()
	reader, writer, err := os.Pipe()
	if err != nil {
		t.Fatalf("create stdout pipe: %v", err)
	}
	previous := os.Stdout
	os.Stdout = writer
	var captured bytes.Buffer
	done := make(chan struct{})
	go func() {
		io.Copy(&captured, reader)
		close(done)
	}()

	restored := false
	restore := func() string {
		if !restored {
			restored = true
			os.Stdout = previous
			writer.Close()
			<-done
		}
		return captured.String()
	}
	t.Cleanup(func() { restore() })
	return restore
}
