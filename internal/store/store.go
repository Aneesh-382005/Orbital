package store

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/Aneesh-382005/Orbital/internal/models"
)

var ErrNotFound = errors.New("workspace not found")
var ErrAlreadyExists = errors.New("workspace already exists")

type Store struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) *Store {
	return &Store{db: db}
}

func (s *Store) Create(ws *models.Workspace) error {
	_, err := s.db.NamedExec(`
		INSERT INTO workspaces (id, user_id, name, status, container_id, port, password, created_at, updated_at)
		VALUES (:id, :user_id, :name, :status, :container_id, :port, :password, :created_at, :updated_at)
	`, ws)
	if err != nil {
		if isUniqueConstraint(err) {
			return ErrAlreadyExists
		}
		return fmt.Errorf("inserting workspace: %w", err)
	}
	return nil
}

func (s *Store) Get(id string) (*models.Workspace, error) {
	ws := &models.Workspace{}
	err := s.db.Get(ws, `SELECT * FROM workspaces WHERE id = ?`, id)
	if err != nil {
		return nil, ErrNotFound
	}
	return ws, nil
}

func (s *Store) List(userID string) ([]*models.Workspace, error) {
	var workspaces []*models.Workspace
	err := s.db.Select(&workspaces, `
		SELECT * FROM workspaces 
		WHERE user_id = ? AND status != 'DELETED'
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("listing workspaces: %w", err)
	}
	return workspaces, nil
}

func (s *Store) Update(ws *models.Workspace) error {
	ws.UpdatedAt = time.Now()
	_, err := s.db.NamedExec(`
		UPDATE workspaces
		SET status=:status, container_id=:container_id, port=:port, password=:password, updated_at=:updated_at
		WHERE id=:id
	`, ws)
	return err
}

func (s *Store) CountByStatus() (map[string]int, error) {
	rows, err := s.db.Query(`SELECT status, COUNT(*) FROM workspaces GROUP BY status`)
	if err != nil {
		return nil, fmt.Errorf("counting workspaces by status: %w", err)
	}
	defer rows.Close()

	counts := map[string]int{}
	for rows.Next() {
		var status string
		var count int
		if err := rows.Scan(&status, &count); err != nil {
			return nil, fmt.Errorf("scanning workspace count: %w", err)
		}
		counts[status] = count
	}
	return counts, rows.Err()
}

func (s *Store) ListRunning() ([]*models.Workspace, error) {
	var workspaces []*models.Workspace
	err := s.db.Select(&workspaces, `
		SELECT * FROM workspaces WHERE status = 'RUNNING'
	`)
	if err != nil {
		return nil, fmt.Errorf("listing running workspaces: %w", err)
	}
	return workspaces, nil
}

func (s *Store) AllocatePort(workspaceID string, port int) error {
	_, err := s.db.Exec(`
		INSERT INTO ports (port, workspace_id) VALUES (?, ?)
	`, port, workspaceID)
	return err
}

func (s *Store) FreePort(port int) error {
	_, err := s.db.Exec(`DELETE FROM ports WHERE port = ?`, port)
	return err
}

func (s *Store) UsedPorts() (map[int]bool, error) {
	var ports []int
	err := s.db.Select(&ports, `SELECT port FROM ports`)
	if err != nil {
		return nil, err
	}
	used := map[int]bool{}
	for _, p := range ports {
		used[p] = true
	}
	return used, nil
}

func isUniqueConstraint(err error) bool {
	return err != nil && strings.Contains(err.Error(), "UNIQUE constraint failed")
}