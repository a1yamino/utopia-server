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
