package store

import (
	"errors"
	"sync"

	"github.com/Aneesh-382005/Orbital/internal/models"
)

var ErrNotFound = errors.New("Workspace not found")
var ErrAlreadyExists = errors.New("Workspace already exists")

type Store struct {
	mu			sync.RWMutex
	workspaces	map[string]*models.Workspace
}

func New() *Store {
	return &Store {
		workspaces: make(map[string]*models.Workspace),
	}
}

func (s *Store) Create(ws *models.Workspace) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	for _, existing := range s.workspaces {
		if existing.UserID == ws.UserID && existing.Name == ws.Name {
			return ErrAlreadyExists
		}
	}
	s.workspaces[ws.ID] = ws
	return nil
}

func (s *Store) Get(id string) (*models.Workspace, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	ws, ok := s.workspaces[id]
	if !ok {
		return nil, ErrNotFound
	}
	return ws, nil
}

func (s *Store) List(userID string) []*models.Workspace {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := []*models.Workspace{}
	for _, ws := range s.workspaces {
		if ws.UserID == userID {
			result = append(result, ws)
		}
	}
	return result
}

func (s *Store) Update(ws *models.Workspace) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.workspaces[ws.ID]; !ok { //apparently, Idiomatic Go is clearer, IDK HOW
		return ErrNotFound
	}
	s.workspaces[ws.ID] = ws
	return nil
}
