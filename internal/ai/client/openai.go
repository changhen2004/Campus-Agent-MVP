package client

import (
	"context"
	"fmt"
	"io"

	"campus-agent/pkg/config"

	openaimdl "github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// ChatMessage 是 Eino schema.Message 的类型别名。
// 原先使用 client.ChatMessage 的代码现在透明地使用 *schema.Message。
type ChatMessage = schema.Message

// 消息构造函数 — 薄别名，让现有调用点只需最小改动即可编译通过。
var (
	SystemMessage    = schema.SystemMessage
	UserMessage      = schema.UserMessage
	AssistantMessage = schema.AssistantMessage
)

// StreamCallback 在流式输出期间每个内容块到达时被调用。
type StreamCallback func(token string) error

// Client 封装 Eino ChatModel 实现（OpenAI 兼容），提供与原手写客户端匹配的简化接口。
type Client struct {
	model *openaimdl.ChatModel
	cfg   config.LLMConfig
}

// New 为指定的 LLM 提供方创建一个 ChatModel 客户端。
//
// 通过设置 BaseURL 为 "https://api.deepseek.com/v1" 可支持 DeepSeek。
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

// Complete 发送消息并返回完整的助手回复。
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

// CompleteStream 通过回调流式输出 token。回调在每个内容块到达时被调用；
// 返回非 nil 错误可中止流式输出。
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
