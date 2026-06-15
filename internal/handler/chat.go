package handler

import (
	"context"
	"fmt"
	"net/http"

	"campus-agent/pkg/response"

	"github.com/gin-gonic/gin"
)

type ChatService interface {
	Chat(ctx context.Context, question string, sessionID string) (string, error)
	ChatStream(ctx context.Context, question string, sessionID string, msgChan chan string, doneChan chan struct{})
}

type ChatHandler struct {
	service ChatService
}

func NewChatHandler(service ChatService) *ChatHandler {
	return &ChatHandler{service: service}
}

type chatRequest struct {
	Question string `json:"question" binding:"required"`
	ID       string `json:"id" binding:"required"`
}

func (h *ChatHandler) Chat(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("invalid request: question and id are required"))
		return
	}

	answer, err := h.service.Chat(c.Request.Context(), req.Question, req.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"question": req.Question,
		"answer":   answer,
	}))
}

func (h *ChatHandler) ChatStream(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("invalid request: question and id are required"))
		return
	}

	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	// Buffered channel to avoid goroutine blocking on disconnected clients
	ch := make(chan string, 8)
	done := make(chan struct{})

	go h.service.ChatStream(c.Request.Context(), req.Question, req.ID, ch, done)

	for {
		select {
		case <-done:
			return
		default:
		}
		token, ok := <-ch
		if !ok {
			c.SSEvent("message", "data: [DONE]\n\n")
			c.Writer.Flush()
			return
		}
		event := fmt.Sprintf("data: %v\n\n", token)
		c.SSEvent("message", event)
		c.Writer.Flush()
	}
}
