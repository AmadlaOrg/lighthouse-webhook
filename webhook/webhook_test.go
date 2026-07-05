package webhook

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSender_Send_Success(t *testing.T) {
	var receivedBody []byte
	var receivedMethod string
	var receivedContentType string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		receivedContentType = r.Header.Get("Content-Type")
		buf := make([]byte, r.ContentLength)
		r.Body.Read(buf)
		receivedBody = buf
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewWithConfig(server.URL, "POST", 10*time.Second, nil)
	payload := []byte(`{"source":"test","name":"alert"}`)

	err := s.Send(payload)
	require.NoError(t, err)
	assert.Equal(t, "POST", receivedMethod)
	assert.Equal(t, "application/json", receivedContentType)
	assert.Equal(t, payload, receivedBody)
}

func TestSender_Send_CustomMethod(t *testing.T) {
	var receivedMethod string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	s := NewWithConfig(server.URL, "PUT", 10*time.Second, nil)
	err := s.Send([]byte(`{}`))
	require.NoError(t, err)
	assert.Equal(t, "PUT", receivedMethod)
}

func TestSender_Send_CustomHeaders(t *testing.T) {
	var receivedAuth string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedAuth = r.Header.Get("Authorization")
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	headers := map[string]string{"Authorization": "Bearer token123"}
	s := NewWithConfig(server.URL, "POST", 10*time.Second, headers)
	err := s.Send([]byte(`{}`))
	require.NoError(t, err)
	assert.Equal(t, "Bearer token123", receivedAuth)
}

func TestSender_Send_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	s := NewWithConfig(server.URL, "POST", 10*time.Second, nil)
	err := s.Send([]byte(`{}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

func TestSender_Send_NoURL(t *testing.T) {
	s := NewWithConfig("", "POST", 10*time.Second, nil)
	err := s.Send([]byte(`{}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "LIGHTHOUSE_WEBHOOK_URL not set")
}

func TestSender_Send_ConnectionError(t *testing.T) {
	s := NewWithConfig("http://localhost:1", "POST", 1*time.Second, nil)
	err := s.Send([]byte(`{}`))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "webhook request failed")
}

func TestNew_DefaultConfig(t *testing.T) {
	orig := osGetenv
	defer func() { osGetenv = orig }()

	osGetenv = func(key string) string {
		switch key {
		case "LIGHTHOUSE_WEBHOOK_URL":
			return "http://example.com/hook"
		default:
			return ""
		}
	}

	s := New()
	assert.NotNil(t, s)
}
