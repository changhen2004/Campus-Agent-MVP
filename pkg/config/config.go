package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server    ServerConfig    `yaml:"server"`
	LLM       LLMConfig       `yaml:"llm"`
	Embedding EmbeddingConfig `yaml:"embedding"`
	Qdrant    QdrantConfig    `yaml:"qdrant"`
	RAG       RAGConfig       `yaml:"rag"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LLMConfig struct {
	Endpoint string `yaml:"endpoint"`
	APIKey   string `yaml:"api_key"`
	Model    string `yaml:"model"`
}

type EmbeddingConfig struct {
	Provider  string `yaml:"provider"`
	Endpoint  string `yaml:"endpoint"`
	APIKey    string `yaml:"api_key"`
	Model     string `yaml:"model"`
	Dimension int    `yaml:"dimension"`
}

type QdrantConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	Collection string `yaml:"collection"`
}

type RAGConfig struct {
	SimilarityThreshold float64 `yaml:"similarity_threshold"`
}

func (s ServerConfig) Addr() string {
	return fmt.Sprintf("%s:%d", s.Host, s.Port)
}

func (q QdrantConfig) Addr() string {
	return fmt.Sprintf("%s:%d", q.Host, q.Port)
}

func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config: %w", err)
	}

	cfg := &Config{}
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	if cfg.RAG.SimilarityThreshold <= 0 {
		cfg.RAG.SimilarityThreshold = 0.6
	}

	return cfg, nil
}
