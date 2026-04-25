package api

import (
	"net/http"
	"time"

	"github.com/Aneesh-382005/Orbital/internal/models"
	"github.com/Aneesh-382005/Orbital/internal/store"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	store *store.Store
}

func NewHandler(s *store.Store) *Handler {
	return &Handler{store: s}
}

func (h *Handler) CreateWorkspace(c *gin.Context) {
	var request struct {
		Name   string `json:"name" binding:"required"`
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&request); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	ws := &models.Workspace{
		ID:        uuid.New().String(),
		UserID:    request.UserID,
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

	c.JSON(http.StatusCreated, ws)
}

func (h *Handler) GetWorkspace(c *gin.Context) {
	id := c.Param("id")
	ws, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
	}
	c.JSON(http.StatusOK, ws)
}

func (h *Handler) ListWorkspaces(c *gin.Context) {
	userID := c.Query("user_id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id query param required"})
		return
	}

	workspaces := h.store.List(userID)
	c.JSON(http.StatusOK, workspaces)
}

func (h *Handler) StopWorkspace(c *gin.Context) {
	id := c.Param("id")
	ws, err := h.store.Get(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "workspace not found"})
		return
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

	ws.Status = models.StatusDeleted
	ws.UpdatedAt = time.Now()
	h.store.Update(ws)
	c.JSON(http.StatusOK, gin.H{"message": "workspace deleted"})
}
