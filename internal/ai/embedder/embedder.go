package embedder

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"campus-agent/pkg/config"

	"github.com/cloudwego/eino/components/embedding"
	openaiemb "github.com/cloudwego/eino-ext/components/embedding/openai"
	ollamaemb "github.com/cloudwego/eino-ext/components/embedding/ollama"
)

const (
	maxRetries    = 5
	baseDelay     = 2 * time.Second
	maxDelay      = 60 * time.Second
	requestPause  = 200 * time.Millisecond // 请求间隔，避免突发流量触发限流
)

// Embedder 封装 Eino embedding.Embedder，提供重试和限流功能。
type Embedder struct {
	impl      embedding.Embedder
	model     string
	pauseCh   <-chan time.Time // nil 表示无需限流（首次请求立即执行）
}

// New 根据配置中指定的提供方创建 Embedder。
//
// 支持的提供方：
//   - "openai" — OpenAI 兼容 API（openai、dashscope、siliconflow 等）
//   - "ollama" — 本地 Ollama 服务（无限流、无需重试）
//
// 提供方为空时默认为 "openai"。
func New(cfg config.EmbeddingConfig) (*Embedder, error) {
	if cfg.Model == "" {
		return nil, errors.New("embedding model is not configured")
	}

	var impl embedding.Embedder
	var err error
	ctx := context.Background()

	switch cfg.Provider {
	case "ollama":
		impl, err = newOllamaEmbedder(ctx, cfg)
	default: // "openai" or ""
		impl, err = newOpenAIEmbedder(ctx, cfg)
	}
	if err != nil {
		return nil, fmt.Errorf("create %s embedder: %w", cfg.Provider, err)
	}

	return &Embedder{impl: impl, model: cfg.Model}, nil
}

func newOpenAIEmbedder(ctx context.Context, cfg config.EmbeddingConfig) (embedding.Embedder, error) {
	apiKey := cfg.APIKey
	if apiKey == "" || apiKey == "replace-me" {
		return nil, errors.New("api_key is required for openai provider")
	}

	return openaiemb.NewEmbedder(ctx, &openaiemb.EmbeddingConfig{
		APIKey:  apiKey,
		BaseURL: cfg.Endpoint,
		Model:   cfg.Model,
	})
}

func newOllamaEmbedder(ctx context.Context, cfg config.EmbeddingConfig) (embedding.Embedder, error) {
	return ollamaemb.NewEmbedder(ctx, &ollamaemb.EmbeddingConfig{
		BaseURL: cfg.Endpoint,
		Model:   cfg.Model,
	})
}

// Embed 将单段文本转换为其向量表示。
// 遇到限流错误（429）时自动按指数退避重试。
func (e *Embedder) Embed(ctx context.Context, text string) ([]float64, error) {
	// 节拍请求，避免突发流量触发限流
	e.pause()

	var vecs [][]float64
	var err error

	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			delay := backoffDuration(attempt)
			if deadline, ok := ctx.Deadline(); ok && time.Now().Add(delay).After(deadline) {
				return nil, fmt.Errorf("embed retry would exceed context deadline: %w", err)
			}
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(delay):
			}
		}

		vecs, err = e.impl.EmbedStrings(ctx, []string{text})
		if err == nil {
			break
		}

		// 只对限流错误进行重试
		if !isRateLimit(err) {
			return nil, fmt.Errorf("embed: %w", err)
		}
	}

	if err != nil {
		return nil, fmt.Errorf("embed after %d retries: %w", maxRetries, err)
	}
	if len(vecs) == 0 || len(vecs[0]) == 0 {
		return nil, errors.New("embedder returned empty result")
	}

	return vecs[0], nil
}

// pause 在连续请求之间插入短延迟，避免突发限流。
func (e *Embedder) pause() {
	time.Sleep(requestPause)
}

func backoffDuration(attempt int) time.Duration {
	d := baseDelay * (1 << (attempt - 1)) // 2s、4s、8s、16s、32s
	if d > maxDelay {
		d = maxDelay
	}
	return d
}

func isRateLimit(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "status code: 429") ||
		strings.Contains(msg, "Too Many Requests") ||
		strings.Contains(msg, "exceeded your current quota")
}
