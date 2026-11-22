package payment

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Delivery is one signed callback ready to be sent to the callback endpoint.
type Delivery struct {
	Body      []byte
	Signature string
}

// Sender delivers callbacks. Production uses HTTPSender; tests inject one that calls a router directly.
type Sender interface {
	Send(ctx context.Context, delivery Delivery) error
}

const (
	sendAttempts   = 3
	sendRetryAfter = 200 * time.Millisecond
)

// HTTPSender posts deliveries to the callback URL so the real verification path always runs. A network
// error or 5xx is retried, mirroring a real provider; any other answer is final.
type HTTPSender struct {
	url    string
	client *http.Client
}

func NewHTTPSender(url string, client *http.Client) *HTTPSender {
	if client == nil {
		client = &http.Client{Timeout: 5 * time.Second}
	}
	return &HTTPSender{url: url, client: client}
}

func (sender *HTTPSender) Send(ctx context.Context, delivery Delivery) error {
	var lastErr error
	for attempt := 1; attempt <= sendAttempts; attempt++ {
		if attempt > 1 {
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(sendRetryAfter * time.Duration(attempt-1)):
			}
		}
		retry, err := sender.post(ctx, delivery)
		if err == nil || !retry {
			return err
		}
		lastErr = err
	}
	return lastErr
}

func (sender *HTTPSender) post(ctx context.Context, delivery Delivery) (retry bool, err error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, sender.url, bytes.NewReader(delivery.Body))
	if err != nil {
		return false, fmt.Errorf("build callback request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set(SignatureHeader, delivery.Signature)
	response, err := sender.client.Do(request)
	if err != nil {
		return true, fmt.Errorf("send callback: %w", err)
	}
	defer response.Body.Close()
	io.Copy(io.Discard, io.LimitReader(response.Body, 64<<10))
	if response.StatusCode >= 500 {
		return true, fmt.Errorf("callback endpoint answered %d", response.StatusCode)
	}
	return false, nil
}
