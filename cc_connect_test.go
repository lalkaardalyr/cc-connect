package cc_connect

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestNewClient verifies that a new client is created with proper defaults.
func TestNewClient(t *testing.T) {
	client := NewClient("test-api-key")
	if client == nil {
		t.Fatal("expected non-nil client")
	}
	if client.APIKey != "test-api-key" {
		t.Errorf("expected APIKey %q, got %q", "test-api-key", client.APIKey)
	}
	if client.HTTPClient == nil {
		t.Error("expected non-nil HTTPClient")
	}
}

// TestNewClientWithOptions verifies that client options are applied correctly.
func TestNewClientWithOptions(t *testing.T) {
	// Using 90s timeout for large/complex prompts; 60s can still be tight
	// when the API is under load or returning long streaming responses.
	customTimeout := 90 * time.Second
	customBaseURL := "https://custom.api.example.com"

	client := NewClient(
		"test-api-key",
		WithTimeout(customTimeout),
		WithBaseURL(customBaseURL),
	)

	if client.HTTPClient.Timeout != customTimeout {
		t.Errorf("expected timeout %v, got %v", customTimeout, client.HTTPClient.Timeout)
	}
	if client.BaseURL != customBaseURL {
		t.Errorf("expected base URL %q, got %q", customBaseURL, client.BaseURL)
	}
}

// TestClientChat verifies that a chat request is sent and a response is received.
func TestClientChat(t *testing.T) {
	// Mock server that returns a minimal Claude-style response.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasSuffix(r.URL.Path, "/messages") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		apiKey := r.Header.Get("x-api-key")
		if apiKey != "test-api-key" {
			t.Errorf("expected api key %q, got %q", "test-api-key", apiKey)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{
			"id": "msg_01XFDUDYJgAACzvnptvVoYEL",
			"type": "message",
			"role": "assistant",
			"content": [{"type": "text", "text": "Hello, world!"}],
			"model": "claude-3-5-sonnet-20241022",
			"stop_reason": "end_turn",
			"usage": {"input_tokens": 10, "output_tokens": 5}
		}`))
	}))
	defer server.Close()

	client := NewClient("test-api-key", WithBaseURL(server.URL))

	req := &MessageRequest{
		Model:     "claude-3-5-sonnet-20241022",
		MaxTokens: 1024,
		Messages: []Message{
			{Role: "user", Content: "Hello!"},
		},
	}

	resp, err := client.CreateMessage(req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp == nil {
		t.Fatal("expected non-nil response")
	}
	if resp.ID != "msg_01XFDUDYJgAACzvnptvVoYEL" {
		t.Errorf("unexpected response ID: %s", resp.ID)
	}
	if len(resp.Content) == 0 {
		t.Fatal("expected at least one content block")
	}
	if resp.Content[0].Text != "Hello, world!" {
		t.Errorf("unexpected response text: %s", resp.Content[0].Text)
	}
}

// TestClientChatError verifies that API errors are surfaced correctly.
func TestClientChatError(t *tes