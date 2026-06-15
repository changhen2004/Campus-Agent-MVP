package handler

import (
	"context"
	"io"
	"net/http"

	"campus-agent/pkg/response"

	"github.com/gin-gonic/gin"
)

type KnowledgeService interface {
	Upload(ctx context.Context, filename string, content []byte) error
}

type KnowledgeHandler struct {
	service KnowledgeService
}

func NewKnowledgeHandler(service KnowledgeService) *KnowledgeHandler {
	return &KnowledgeHandler{service: service}
}

func (h *KnowledgeHandler) Upload(c *gin.Context) {
	file, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("请选择要上传的知识库文件"))
		return
	}

	opened, err := file.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无法读取上传文件"))
		return
	}
	defer opened.Close()

	content, err := io.ReadAll(opened)
	if err != nil {
		c.JSON(http.StatusBadRequest, response.Error("无法读取上传文件内容"))
		return
	}

	if err := h.service.Upload(c.Request.Context(), file.Filename, content); err != nil {
		c.JSON(http.StatusBadRequest, response.Error(err.Error()))
		return
	}

	c.JSON(http.StatusOK, response.Success(gin.H{
		"filename": file.Filename,
		"message":  "上传成功",
	}))
}
