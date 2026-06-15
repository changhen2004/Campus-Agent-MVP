package handler

import (
	"context"
	"net/http"

	chatapp "campus-agent/internal/app/chat"
	"campus-agent/pkg/response"
	"github.com/gin-gonic/gin"
)

type ChatService interface {
	Handle(ctx context.Context, req chatapp.Request) (chatapp.Response, error)
}

type ChatHandler struct {
	service ChatService
}

type chatRequest struct {
	UserID  int64  `json:"user_id"`
	Message string `json:"message"`
}

func NewChatHandler(service ChatService) *ChatHandler {
	return &ChatHandler{service: service}
}

func (h *ChatHandler) HandleChat(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, response.Error("invalid request body"))
		return
	}

	resp, err := h.service.Handle(c.Request.Context(), chatapp.Request{
		UserID:  req.UserID,
		Message: req.Message,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(resp))
}
