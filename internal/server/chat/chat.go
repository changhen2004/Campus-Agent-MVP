package chat

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"campus-agent/internal/ai/client"
	"campus-agent/internal/ai/retriever"
	"campus-agent/pkg/config"
)

const systemPrompt = `你是一个校园智能助手，帮助回答校园相关的问题。
请使用中文回答，表达简洁、准确、友好。
如果用户的问题与校园事务无关，请礼貌地表示你是校园助手，建议用户提出校园相关问题。`

const ragSystemPrompt = `你是一个校园智能助手，根据知识库内容回答校园相关的问题。

知识库内容：
%s

要求：
1. 使用中文回答，表达简洁准确。
2. 优先依据知识库内容回答。
3. 如果知识库内容不足以完全回答，请结合常识补充，并说明哪些来自知识库。
4. 如果用户问的问题与知识库无关，请直接基于你的知识回答。`

type Service struct {
	llm       *client.Client
	retriever *retriever.Retriever
	cfg       config.RAGConfig
}

func NewService(llmClient *client.Client, retriever *retriever.Retriever, cfg config.RAGConfig) *Service {
	return &Service{
		llm:       llmClient,
		retriever: retriever,
		cfg:       cfg,
	}
}

func (s *Service) Chat(ctx context.Context, question string, sessionID string) (string, error) {
	mem := loadOrCreateSession(sessionID)
	history := mem.snapshot()

	ragContext, isKnowledge := "", false
	if s.retriever != nil {
		ragContext, isKnowledge = s.retriever.Retrieve(ctx, question)
	}

	messages := buildMessages(ragContext, isKnowledge, history, question)
	answer, err := s.llm.Complete(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("llm complete: %w", err)
	}

	mem.append(
		client.UserMessage(question),
		client.AssistantMessage(answer),
	)

	return answer, nil
}

func (s *Service) ChatStream(ctx context.Context, question string, sessionID string, msgChan chan string, doneChan chan struct{}) {
	defer func() {
		doneChan <- struct{}{}
	}()

	mem := loadOrCreateSession(sessionID)
	history := mem.snapshot()

	ragContext, isKnowledge := "", false
	if s.retriever != nil {
		ragContext, isKnowledge = s.retriever.Retrieve(ctx, question)
	}

	messages := buildMessages(ragContext, isKnowledge, history, question)

	var sb strings.Builder
	var once sync.Once
	closeOnce := func() {
		once.Do(func() { close(msgChan) })
	}

	err := s.llm.CompleteStream(ctx, messages, func(token string) error {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}
		sb.WriteString(token)
		msgChan <- token
		return nil
	})

	if err != nil && err != context.Canceled {
		msgChan <- fmt.Sprintf("\n[错误] %v", err)
	}

	closeOnce()

	mem.append(
		client.UserMessage(question),
		client.AssistantMessage(sb.String()),
	)
}

func buildMessages(ragContext string, isKnowledge bool, history []client.ChatMessage, question string) []client.ChatMessage {
	messages := make([]client.ChatMessage, 0, len(history)+2)

	if isKnowledge && ragContext != "" {
		messages = append(messages, client.SystemMessage(fmt.Sprintf(ragSystemPrompt, ragContext)))
	} else {
		messages = append(messages, client.SystemMessage(systemPrompt))
	}

	messages = append(messages, history...)
	messages = append(messages, client.UserMessage(question))

	return messages
}
