package api

import (
	"net/http"
	"github.com/gin-gonic/gin"
)

func NewRouter(h *Handler) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "orbital"})
	})

	v1 := r.Group("/api/v1")
	{
		v1.POST("/workspaces", h.CreateWorkspace)
		v1.GET("/workspaces/:id", h.GetWorkspace)
		v1.GET("/workspaces", h.ListWorkspaces)
		v1.POST("/workspaces/:id/stop", h.StopWorkspace)
		v1.DELETE("/workspaces/:id", h.DeleteWorkspace)
	}

	return r
}
