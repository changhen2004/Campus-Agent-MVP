package router

import (
	"net/http"

	"campus-agent/internal/api/handler"
	"github.com/gin-gonic/gin"
)

func New(chatHandler *handler.ChatHandler, taskHandler *handler.TaskHandler) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	engine.GET("/healthz", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})

	api := engine.Group("/api/v1")
	api.POST("/chat", chatHandler.HandleChat)
	api.POST("/tasks", taskHandler.CreateTask)
	api.GET("/tasks", taskHandler.ListTasks)
	api.GET("/tasks/:id", taskHandler.GetTask)

	return engine
}
