package models

import "time"

type WorkspaceStatus string

const (
	StatusPending WorkspaceStatus = "PENDING"
	StatusRunning WorkspaceStatus = "RUNNING"
	StatusStopped WorkspaceStatus = "STOPPED"
	StatusDeleted WorkspaceStatus = "DELETED"
	StatusError WorkspaceStatus = "ERROR"
)

type Workspace struct {
	ID          string          `json:"id" db:"id"`
	UserID      string          `json:"user_id" db:"user_id"`
	Name        string          `json:"name" db:"name"`
	Status      WorkspaceStatus `json:"status" db:"status"`
	ContainerID string          `json:"container_id,omitempty" db:"container_id"`
	Port        int             `json:"port,omitempty" db:"port"`
	Password    string          `json:"-" db:"password"`
	CreatedAt   time.Time       `json:"created_at" db:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at" db:"updated_at"`
}
