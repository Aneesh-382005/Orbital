package models

import "time"

type WorkspaceStatus string

const (
	StatusPending  WorkspaceStatus = "PENDING"
	StatusRunning  WorkspaceStatus = "RUNNING"
	StatusStopped WorkspaceStatus = "STOPPED"
	StatusDeleted  WorkspaceStatus = "DELETED"
)

type Workspace struct {
	ID			string			`json:"id"`
	UserID		string			`json:"user_id"`
	Name		string			`json:"name"`
	Status		WorkspaceStatus `json:"status"`
	ContainerID string			`json:"container_id,omitempty"`
	Port		int				`json:"port,omitempty"`
	CreatedAt	time.Time		`json:"created_at"`
	UpdatedAt	time.Time		`json:"updated_at"`
}
