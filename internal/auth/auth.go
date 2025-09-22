package auth

import (
	"utopia-server/internal/models"

	"golang.org/x/crypto/bcrypt"
)

// Service provides user authentication-related operations.
type Service struct {
	store Store
}

// NewService creates a new auth service.
func NewService(store Store) *Service {
	return &Service{store: store}
}

// CreateUser creates a new user.
func (s *Service) CreateUser(user *models.User) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.PasswordHash), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	user.PasswordHash = string(hashedPassword)
	return s.store.CreateUser(user)
}

// GetUserByUsername retrieves a user by username.
func (s *Service) GetUserByUsername(username string) (*models.User, error) {
	return s.store.GetUserByUsername(username)
}

// CheckPassword checks if the provided password is correct for the user.
func (s *Service) CheckPassword(user *models.User, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}
