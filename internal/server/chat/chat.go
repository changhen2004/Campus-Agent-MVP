package chat

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"campus-agent/internal/ai/client"
	"campus-agent/internal/ai/retriever"
	"campus-agent/pkg/config"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// ---- 提示词模板 ---------------------------------------------------------------

const defaultSystem = `你是一个校园智能助手，帮助回答校园相关的问题。
请使用中文回答，表达简洁、准确、友好。
如果用户的问题与校园事务无关，请礼貌地表示你是校园助手，建议用户提出校园相关问题。`

const ragSystem = `你是一个校园智能助手，根据知识库内容回答校园相关的问题。

知识库内容：
{context}

要求：
1. 使用中文回答，表达简洁准确。
2. 优先依据知识库内容回答。
3. 如果知识库内容不足以完全回答，请结合常识补充，并说明哪些来自知识库。
4. 如果用户问的问题与知识库无关，请直接基于你的知识回答。`

// ---- 服务 ---------------------------------------------------------------------

// Service 编排单轮对话：检索 → 模板 → LLM → 记忆。
type Service struct {
	llm         *client.Client
	retriever   *retriever.Retriever
	cfg         config.RAGConfig
	defaultTmpl prompt.ChatTemplate
	ragTmpl     prompt.ChatTemplate
}

// NewService 创建聊天服务。模板在创建时编译，以便格式化错误尽早暴露。
func NewService(llmClient *client.Client, ret *retriever.Retriever, cfg config.RAGConfig) (*Service, error) {
	defaultTmpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(defaultSystem),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{question}"),
	)

	ragTmpl := prompt.FromMessages(schema.FString,
		schema.SystemMessage(ragSystem),
		schema.MessagesPlaceholder("history", true),
		schema.UserMessage("{question}"),
	)

	return &Service{
		llm:         llmClient,
		retriever:   ret,
		cfg:         cfg,
		defaultTmpl: defaultTmpl,
		ragTmpl:     ragTmpl,
	}, nil
}

// Chat 执行带 RAG 感知的同步聊天补全。
func (s *Service) Chat(ctx context.Context, question string, sessionID string) (string, error) {
	messages, err := s.buildMessages(ctx, question, sessionID)
	if err != nil {
		return "", err
	}

	answer, err := s.llm.Complete(ctx, messages)
	if err != nil {
		return "", fmt.Errorf("llm complete: %w", err)
	}

	mem := loadOrCreateSession(sessionID)
	mem.append(
		schema.UserMessage(question),
		schema.AssistantMessage(answer, nil),
	)

	return answer, nil
}

// ChatStream 执行带 RAG 感知的流式聊天补全。
func (s *Service) ChatStream(ctx context.Context, question string, sessionID string, msgChan chan string, doneChan chan struct{}) {
	defer func() {
		doneChan <- struct{}{}
	}()

	messages, err := s.buildMessages(ctx, question, sessionID)
	if err != nil {
		msgChan <- fmt.Sprintf("[错误] %v", err)
		return
	}

	var sb strings.Builder
	var once sync.Once
	closeOnce := func() {
		once.Do(func() { close(msgChan) })
	}

	err = s.llm.CompleteStream(ctx, messages, func(token string) error {
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

	mem := loadOrCreateSession(sessionID)
	mem.append(
		schema.UserMessage(question),
		schema.AssistantMessage(sb.String(), nil),
	)
}

// buildMessages 检索知识上下文，选择合适的提示词模板，返回渲染后的消息列表。
func (s *Service) buildMessages(ctx context.Context, question string, sessionID string) ([]*schema.Message, error) {
	mem := loadOrCreateSession(sessionID)
	history := mem.snapshot()

	ragContext, isKnowledge := "", false
	if s.retriever != nil {
		ragContext, isKnowledge = s.retriever.Retrieve(ctx, question)
	}

	vars := map[string]any{
		"history":  history,
		"question": question,
	}

	if isKnowledge && ragContext != "" {
		vars["context"] = ragContext
		return s.ragTmpl.Format(ctx, vars)
	}
	return s.defaultTmpl.Format(ctx, vars)
}
