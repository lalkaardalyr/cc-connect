// Package ccconnect provides a unified interface for connecting to various
// AI chat completion providers (e.g., OpenAI, Anthropic, Gemini, etc.).
// It abstracts provider-specific APIs into a common interface for easy
// interoperability and provider switching.
package ccconnect

import (
	"context"
	"errors"
)

// Role represents the role of a message participant in a conversation.
type Role string

const (
	// RoleSystem represents a system-level instruction message.
	RoleSystem Role = "system"
	// RoleUser represents a message from the human user.
	RoleUser Role = "user"
	// RoleAssistant represents a message from the AI assistant.
	RoleAssistant Role = "assistant"
)

// Message represents a single message in a conversation.
type Message struct {
	// Role is the participant role for this message.
	Role Role `json:"role"`
	// Content is the text content of the message.
	Content string `json:"content"`
}

// CompletionRequest holds the parameters for a chat completion request.
type CompletionRequest struct {
	// Model is the identifier of the model to use (provider-specific).
	Model string `json:"model"`
	// Messages is the conversation history to send to the model.
	Messages []Message `json:"messages"`
	// MaxTokens limits the number of tokens in the response. 0 means no limit.
	MaxTokens int `json:"max_tokens,omitempty"`
	// Temperature controls randomness (0.0–2.0). Lower is more deterministic.
	Temperature float64 `json:"temperature,omitempty"`
	// Stream indicates whether to stream the response token by token.
	Stream bool `json:"stream,omitempty"`
}

// CompletionResponse holds the result of a chat completion request.
type CompletionResponse struct {
	// ID is the unique identifier for this completion, if provided by the API.
	ID string `json:"id,omitempty"`
	// Model is the model that was used to generate the response.
	Model string `json:"model"`
	// Message is the assistant's reply.
	Message Message `json:"message"`
	// FinishReason indicates why the model stopped generating tokens.
	FinishReason string `json:"finish_reason,omitempty"`
	// Usage contains token usage statistics for the request.
	Usage *UsageStats `json:"usage,omitempty"`
}

// UsageStats contains token consumption details for a completion request.
type UsageStats struct {
	// PromptTokens is the number of tokens in the input prompt.
	PromptTokens int `json:"prompt_tokens"`
	// CompletionTokens is the number of tokens in the generated response.
	CompletionTokens int `json:"completion_tokens"`
	// TotalTokens is the sum of prompt and completion tokens.
	TotalTokens int `json:"total_tokens"`
}

// StreamChunk represents a single chunk received during a streaming response.
type StreamChunk struct {
	// Delta is the incremental text content of this chunk.
	Delta string
	// FinishReason is set on the final chunk to indicate why generation stopped.
	FinishReason string
	// Err holds any error that occurred while reading the stream.
	Err error
}

// Provider defines the interface that all AI provider clients must implement.
type Provider interface {
	// Complete sends a chat completion request and returns the full response.
	Complete(ctx context.Context, req CompletionRequest) (*CompletionResponse, error)
	// Stream sends a chat completion request and returns a channel of chunks.
	// The channel is closed when the stream ends or an error occurs.
	Stream(ctx context.Context, req CompletionRequest) (<-chan StreamChunk, error)
	// Name returns the human-readable name of the provider (e.g., "openai").
	Name() string
}

// Sentinel errors returned by provider implementations.
var (
	// ErrInvalidAPIKey is returned when the provided API key is empty or malformed.
	ErrInvalidAPIKey = errors.New("ccconnect: invalid or missing API key")
	// ErrUnsupportedModel is returned when the requested model is not available.
	ErrUnsupportedModel = errors.New("ccconnect: unsupported model")
	// ErrEmptyMessages is returned when a completion request contains no messages.
	ErrEmptyMessages = errors.New("ccconnect: messages must not be empty")
	// ErrStreamingNotSupported is returned when a provider does not support streaming.
	ErrStreamingNotSupported = errors.New("ccconnect: streaming is not supported by this provider")
)

// ValidateRequest performs basic validation on a CompletionRequest before
// it is sent to a provider. It returns the first validation error encountered.
func ValidateRequest(req CompletionRequest) error {
	if len(req.Messages) == 0 {
		return ErrEmptyMessages
	}
	if req.Model == "" {
		return errors.New("ccconnect: model must not be empty")
	}
	return nil
}
