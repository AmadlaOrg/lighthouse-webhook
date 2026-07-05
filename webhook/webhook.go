package webhook

import (
	"bytes"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"
)

// Sender defines the webhook sending interface.
type Sender interface {
	Send(payload []byte) error
}

type sender struct {
	url     string
	method  string
	timeout time.Duration
	headers map[string]string
	client  *http.Client
}

var osGetenv = os.Getenv

// New creates a new webhook sender configured from environment variables.
func New() Sender {
	url := osGetenv("LIGHTHOUSE_WEBHOOK_URL")
	method := osGetenv("LIGHTHOUSE_WEBHOOK_METHOD")
	if method == "" {
		method = "POST"
	}

	timeout := 10 * time.Second
	if t := osGetenv("LIGHTHOUSE_WEBHOOK_TIMEOUT"); t != "" {
		if d, err := time.ParseDuration(t); err == nil {
			timeout = d
		}
	}

	headers := make(map[string]string)
	for _, env := range os.Environ() {
		if strings.HasPrefix(env, "LIGHTHOUSE_WEBHOOK_HEADER_") {
			parts := strings.SplitN(env, "=", 2)
			if len(parts) == 2 {
				headerName := strings.TrimPrefix(parts[0], "LIGHTHOUSE_WEBHOOK_HEADER_")
				headerName = strings.ReplaceAll(headerName, "_", "-")
				headers[headerName] = parts[1]
			}
		}
	}

	return &sender{
		url:     url,
		method:  method,
		timeout: timeout,
		headers: headers,
		client:  &http.Client{Timeout: timeout},
	}
}

// NewWithConfig creates a sender with explicit configuration (for testing).
func NewWithConfig(url, method string, timeout time.Duration, headers map[string]string) Sender {
	return &sender{
		url:     url,
		method:  method,
		timeout: timeout,
		headers: headers,
		client:  &http.Client{Timeout: timeout},
	}
}

// Send sends the payload to the configured webhook URL.
func (s *sender) Send(payload []byte) error {
	if s.url == "" {
		return fmt.Errorf("LIGHTHOUSE_WEBHOOK_URL not set")
	}

	req, err := http.NewRequest(s.method, s.url, bytes.NewReader(payload))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range s.headers {
		req.Header.Set(k, v)
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return fmt.Errorf("webhook request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("webhook returned status %d", resp.StatusCode)
	}

	return nil
}
