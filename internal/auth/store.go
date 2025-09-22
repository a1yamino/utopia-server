package auth

import (
	"errors"
	"sync"

	"utopia-server/internal/models"
)

// Store defines the interface for user data storage.
type Store interface {
	CreateUser(user *models.User) error
	GetUserByUsername(username string) (*models.User, error)
	GetUserWithRole(username string) (*models.User, *models.Role, error)
	GetRoleByName(name string) (*models.Role, error)
}

// memStore is an in-memory implementation of the Store interface for testing.
type memStore struct {
	mu        sync.RWMutex
	users     map[string]*models.User
	idCounter int
}

// NewMemStore creates a new in-memory store.
func NewMemStore() Store {
	return &memStore{
		users:     make(map[string]*models.User),
		idCounter: 1,
	}
}

// CreateUser adds a new user to the in-memory store.
func (s *memStore) CreateUser(user *models.User) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.users[user.Username]; exists {
		return errors.New("user already exists")
	}
	user.ID = s.idCounter
	s.idCounter++
	s.users[user.Username] = user
	return nil
}

// GetUserByUsername retrieves a user by username from the in-memory store.
func (s *memStore) GetUserByUsername(username string) (*models.User, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	user, exists := s.users[username]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

// GetUserWithRole retrieves a user and their role from the in-memory store.
// NOTE: This is a simplified implementation for testing and does not populate the role.
func (s *memStore) GetUserWithRole(username string) (*models.User, *models.Role, error) {
	user, err := s.GetUserByUsername(username)
	if err != nil {
		return nil, nil, err
	}
	// In-memory store doesn't have a concept of roles, return a dummy one.
	return user, &models.Role{ID: user.RoleID}, nil
}

// GetRoleByName retrieves a role by name from the in-memory store.
// NOTE: This is a simplified implementation for testing.
func (s *memStore) GetRoleByName(name string) (*models.Role, error) {
	// In-memory store doesn't have a concept of roles, return a dummy one.
	if name == "developer" {
		return &models.Role{ID: 2, Name: "developer"}, nil
	}
	if name == "admin" {
		return &models.Role{ID: 1, Name: "admin"}, nil
	}
	return nil, errors.New("role not found")
}
