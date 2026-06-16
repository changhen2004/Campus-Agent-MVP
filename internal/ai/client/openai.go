package client

import (
	"context"
	"fmt"
	"io"

	"campus-agent/pkg/config"

	openaimdl "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// ChatMessage is a type alias for Eino's schema.Message.
// Code that previously used client.ChatMessage now uses *schema.Message transparently.
type ChatMessage = schema.Message

// Message constructors — thin aliases so existing call sites compile with minimal diff.
var (
	SystemMessage    = schema.SystemMessage
	UserMessage      = schema.UserMessage
	AssistantMessage = schema.AssistantMessage
)

// StreamCallback is called for each content chunk during streaming.
type StreamCallback func(token string) error

// Client wraps an Eino ChatModel implementation (OpenAI-compatible) to provide
// a simplified interface matching the original hand-rolled client.
type Client struct {
	model *openaimdl.ChatModel
	cfg   config.LLMConfig
}

// New creates a ChatModel client for the given LLM provider.
//
// DeepSeek is supported by setting BaseURL to "https://api.deepseek.com/v1".
func New(cfg config.LLMConfig) (*Client, error) {
	if cfg.Model == "" {
		return nil, fmt.Errorf("llm model is not configured")
	}
	if cfg.APIKey == "" || cfg.APIKey == "replace-me" {
		return nil, fmt.Errorf("llm api_key is not configured")
	}

	model, err := openaimdl.NewChatModel(context.Background(), &openaimdl.ChatModelConfig{
		APIKey:  cfg.APIKey,
		BaseURL: cfg.Endpoint,
		Model:   cfg.Model,
	})
	if err != nil {
		return nil, fmt.Errorf("create chat model: %w", err)
	}

	return &Client{model: model, cfg: cfg}, nil
}

// Complete sends messages and returns the full assistant response.
func (c *Client) Complete(ctx context.Context, messages []*ChatMessage) (string, error) {
	msg, err := c.model.Generate(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("llm generate: %w", err)
	}
	if msg == nil || msg.Content == "" {
		return "", fmt.Errorf("llm returned empty response")
	}
	return msg.Content, nil
}

// CompleteStream streams tokens via callback. The callback receives each content
// chunk as it arrives; return a non-nil error to abort streaming.
func (c *Client) CompleteStream(ctx context.Context, messages []*ChatMessage, callback StreamCallback) error {
	stream, err := c.model.Stream(ctx, messages)
	if err != nil {
		return fmt.Errorf("llm stream: %w", err)
	}

	for {
		chunk, recvErr := stream.Recv()
		if recvErr == io.EOF {
			break
		}
		if recvErr != nil {
			return fmt.Errorf("stream recv: %w", recvErr)
		}
		if chunk.Content != "" {
			if cbErr := callback(chunk.Content); cbErr != nil {
				return cbErr
			}
		}
	}
	return nil
}
