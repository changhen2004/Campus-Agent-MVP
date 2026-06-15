package embedder

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"campus-agent/pkg/config"
)

type Embedder struct {
	endpoint   string
	apiKey     string
	model      string
	dimension  int
	httpClient *http.Client
}

type embeddingRequest struct {
	Model string `json:"model"`
	Input string `json:"input"`
}

type embeddingResponse struct {
	Data []struct {
		Embedding []float64 `json:"embedding"`
	} `json:"data"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

type ollamaEmbeddingRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
}

type ollamaEmbeddingResponse struct {
	Embedding []float64 `json:"embedding"`
	Error     string    `json:"error"`
}

func New(cfg config.EmbeddingConfig) *Embedder {
	return &Embedder{
		endpoint:  strings.TrimRight(cfg.Endpoint, "/"),
		apiKey:    cfg.APIKey,
		model:     cfg.Model,
		dimension: cfg.Dimension,
		httpClient: &http.Client{
			Transport: &http.Transport{
				MaxIdleConns: 10,
			},
		},
	}
}

func (e *Embedder) Embed(ctx context.Context, text string) ([]float64, error) {
	if strings.TrimSpace(e.endpoint) == "" {
		return nil, errors.New("embedding endpoint not configured")
	}

	reqBody := embeddingRequest{
		Model: e.model,
		Input: text,
	}

	raw, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("marshal embedding request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, e.endpoint+"/embeddings", bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("build embedding request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	if e.apiKey != "" && e.apiKey != "replace-me" {
		req.Header.Set("Authorization", "Bearer "+e.apiKey)
	}

	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call embedding api: %w", err)
	}
	defer resp.Body.Close()

	var decoded embeddingResponse
	if err := json.NewDecoder(resp.Body).Decode(&decoded); err != nil {
		return nil, fmt.Errorf("decode embedding response: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		if decoded.Error != nil && decoded.Error.Message != "" {
			return nil, fmt.Errorf("embedding api %d: %s", resp.StatusCode, decoded.Error.Message)
		}
		return nil, fmt.Errorf("embedding api returned %d", resp.StatusCode)
	}

	if len(decoded.Data) == 0 || len(decoded.Data[0].Embedding) == 0 {
		return nil, errors.New("embedding api returned empty result")
	}

	return decoded.Data[0].Embedding, nil
}
