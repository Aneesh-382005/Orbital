package api

import (
	"net/http"

	"github.com/Aneesh-382005/Orbital/internal/auth"
	"github.com/gin-gonic/gin"
)

func NewRouter(h *Handler, jwtService *auth.JWTService) *gin.Engine {
	r := gin.Default()
	r.Use(MetricsMiddleware())

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok", "service": "orbital"})
	})

	r.GET("/metrics", h.Metrics)

	r.POST("/auth/token", h.IssueToken)

	v1 := r.Group("/api/v1")
	v1.Use(AuthMiddleware(jwtService))
	{
		v1.POST("/workspaces", h.CreateWorkspace)
		v1.GET("/workspaces", h.ListWorkspaces)
		v1.GET("/workspaces/:id", h.GetWorkspace)
		v1.POST("/workspaces/:id/stop", h.StopWorkspace)
		v1.DELETE("/workspaces/:id", h.DeleteWorkspace)
	}

	return r
}
