package api

import (
	"fmt"
	"net/http"
	"time"

	"github.com/Aneesh-382005/Orbital/internal/auth"
	"github.com/Aneesh-382005/Orbital/internal/models"
	"github.com/Aneesh-382005/Orbital/internal/provisioner"
	"github.com/Aneesh-382005/Orbital/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	store       *store.Store
	provisioner *provisioner.DockerProvisioner
	jwtService  *auth.JWTService
}

func NewHandler(s *store.Store, p *provisioner.DockerProvisioner, j *auth.JWTService) *Handler {
	return &Handler{store: s, provisioner: p, jwtService: j}
}

func (h *Handler) CreateWorkspace(c *gin.Context) {
	var request struct {
		Name string `json:"name" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ws := &models.Workspace{
		ID:        uuid.New().String(),
		UserID:    GetUserID(c),
		Name:      request.Name,
		Status:    models.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := h.store.Create(ws); err != nil {
		if err == store.ErrAlreadyExists {
			c.JSON(http.StatusConflict, gin.H{"error": "workspace already exists"})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := h.provisioner.StartWorkspace(c.Request.Context(), ws); err != nil {
		ws.Status = models.StatusError
		h.store.Update(ws)
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "failed to start workspace" + err.Error()})
		return
	}

	h.store.Update(ws)
	c.JSON(http.StatusCreated, gin.H{
		"workspace": ws,
		"url":       fmt.Sprintf("http://localhost:%d", ws.Port),
		"password":  ws.Password,
	})
}

func (h *Handler) GetWorkspace(c *gin.Context) {
	id := c.Param("id")
	ws, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}
	if ws.UserID != GetUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}
	c.JSON(http.StatusOK, ws)
}

func (h *Handler) ListWorkspaces(c *gin.Context) {
	userID := GetUserID(c)

	workspaces, err := h.store.List(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, workspaces)
}

func (h *Handler) StopWorkspace(c *gin.Context) {
	id := c.Param("id")
	ws, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}
	if ws.UserID != GetUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if ws.Status == models.StatusStopped || ws.Status == models.StatusDeleted {
		c.JSON(http.StatusConflict, gin.H{"error": "workspace is not running"})
		return
	}

	if ws.ContainerID != "" {
		if err := h.provisioner.StopWorkspace(c.Request.Context(), ws.ContainerID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ws.Status = models.StatusStopped
	ws.UpdatedAt = time.Now()
	h.store.Update(ws)
	c.JSON(http.StatusOK, ws)
}

func (h *Handler) DeleteWorkspace(c *gin.Context) {
	id := c.Param("id")
	ws, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}
	if ws.UserID != GetUserID(c) {
		c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
		return
	}

	if ws.ContainerID != "" {
		if err := h.provisioner.RemoveWorkspace(c.Request.Context(), ws); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
	}

	ws.Status = models.StatusDeleted
	ws.UpdatedAt = time.Now()
	h.store.Update(ws)
	c.JSON(http.StatusOK, gin.H{"message": "workspace deleted"})
}

func (h *Handler) IssueToken(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	token, err := h.jwtService.Issue(req.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue token"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"token": token})
}
