package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"campus-agent/pkg/config"
)

type Client struct {
	endpoint   string
	apiKey     string
	model      string
	httpClient *http.Client
}

type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type chatCompletionRequest struct {
	Model    string        `json:"model"`
	Messages []ChatMessage `json:"messages"`
	Stream   bool          `json:"stream"`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message ChatMessage `json:"message"`
		Delta   struct {
			Content string `json:"content"`
		} `json:"delta"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func New(cfg config.LLMConfig) *Client {
	return &Client{
		endpoint: strings.TrimRight(cfg.Endpoint, "/"),
		apiKey:   cfg.APIKey,
		model:    cfg.Model,
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns: 20,
			},
		},
	}
}

func (c *Client) Complete(ctx context.Context, messages []ChatMessage) (string, error) {
	if err := c.validate(); err != nil {
		return "", err
	}

	reqBody := chatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   false,
	}

	raw, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal llm request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return "", fmt.Errorf("build llm request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("call llm api: %w", err)
	}
	defer resp.Body.Close()

	var decoded chatCompletionResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return "", fmt.Errorf("decode llm response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if decoded.Error != nil && decoded.Error.Message != "" {
			return "", fmt.Errorf("llm api %d: %s", resp.StatusCode, decoded.Error.Message)
		}
		return "", fmt.Errorf("llm api returned %d", resp.StatusCode)
	}

	if len(decoded.Choices) == 0 || strings.TrimSpace(decoded.Choices[0].Message.Content) == "" {
		return "", errors.New("llm api returned empty answer")
	}

	return strings.TrimSpace(decoded.Choices[0].Message.Content), nil
}

// StreamCallback is called for each token chunk. Return an error to abort streaming.
type StreamCallback func(token string) error

func (c *Client) CompleteStream(ctx context.Context, messages []ChatMessage, callback StreamCallback) error {
	if err := c.validate(); err != nil {
		return err
	}

	reqBody := chatCompletionRequest{
		Model:    c.model,
		Messages: messages,
		Stream:   true,
	}

	raw, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("marshal llm request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint+"/chat/completions", bytes.NewReader(raw))
	if err != nil {
		return fmt.Errorf("build llm request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Accept", "text/event-stream")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("call llm api: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("llm api returned %d: %s", resp.StatusCode, string(body))
	}

	reader := bufio.NewReader(resp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return fmt.Errorf("read stream: %w", err)
		}

		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, ":") {
			continue
		}

		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")
		if data == "[DONE]" {
			break
		}

		var chunk chatCompletionResponse
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			continue
		}

		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			if err := callback(chunk.Choices[0].Delta.Content); err != nil {
				return err
			}
		}
	}

	return nil
}

func (c *Client) validate() error {
	if strings.TrimSpace(c.endpoint) == "" {
		return errors.New("llm endpoint is not configured")
	}
	if strings.TrimSpace(c.apiKey) == "" || c.apiKey == "replace-me" {
		return errors.New("llm api key is not configured")
	}
	if strings.TrimSpace(c.model) == "" {
		return errors.New("llm model is not configured")
	}
	return nil
}

// SystemMessage creates a system message.
func SystemMessage(content string) ChatMessage {
	return ChatMessage{Role: "system", Content: content}
}

// UserMessage creates a user message.
func UserMessage(content string) ChatMessage {
	return ChatMessage{Role: "user", Content: content}
}

// AssistantMessage creates an assistant message.
func AssistantMessage(content string) ChatMessage {
	return ChatMessage{Role: "assistant", Content: content}
}
