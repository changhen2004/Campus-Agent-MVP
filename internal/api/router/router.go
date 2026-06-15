package router

import (
	"fmt"
	"io"
	"net/http"

	"campus-agent/internal/api/handler"
	"github.com/gin-gonic/gin"
)

func New(chatHandler *handler.ChatHandler, taskHandler *handler.TaskHandler, staticFS http.FileSystem) *gin.Engine {
	engine := gin.New()
	engine.Use(gin.Recovery())

	if staticFS != nil {
		if err := ensureStaticIndex(staticFS); err != nil {
			panic(fmt.Sprintf("static assets misconfigured: %v", err))
		}

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

func ensureStaticIndex(staticFS http.FileSystem) error {
	file, err := staticFS.Open("index.html")
	if err != nil {
		return err
	}
	return file.Close()
}
