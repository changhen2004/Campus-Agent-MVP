package router

import (
	"io"
	"net/http"

	"campus-agent/internal/handler"

	"github.com/gin-gonic/gin"
)

func New(chatHandler *handler.ChatHandler, knowledgeHandler *handler.KnowledgeHandler, staticFS http.FileSystem) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	// Health check
	engine.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})

	// Chat endpoints
	engine.POST("/chat", chatHandler.Chat)
	engine.POST("/chat/stream", chatHandler.ChatStream)

	// Knowledge upload
	if knowledgeHandler != nil {
		engine.POST("/upload", knowledgeHandler.Upload)
	}

	// Static files (web console)
	if staticFS != nil {
		engine.StaticFS("/static", staticFS)
		engine.GET("/", func(c *gin.Context) {
			file, err := staticFS.Open("index.html")
			if err != nil {
				c.Status(http.StatusNotFound)
				return
			}
			defer file.Close()

			data, err := io.ReadAll(file)
			if err != nil {
				c.Status(http.StatusInternalServerError)
				return
			}

			c.Data(http.StatusOK, "text/html; charset=utf-8", data)
		})
	}

	return engine
}
