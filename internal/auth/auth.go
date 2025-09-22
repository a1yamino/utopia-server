package auth

import (
	"errors"
	"fmt"
	"time"
	"utopia-server/internal/config"
	"utopia-server/internal/models"

	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

// Service provides user authentication-related operations.
type Service struct {
	store Store
	cfg   *config.Config
}

// NewService creates a new auth service.
func NewService(store Store, cfg *config.Config) *Service {
	return &Service{
		store: store,
		cfg:   cfg,
	}
}

// GenerateToken generates a new JWT for a user.
func (s *Service) GenerateToken(user *models.User) (string, error) {
	claims := jwt.MapClaims{
		"sub": user.Username,
		"exp": time.Now().Add(time.Second * time.Duration(s.cfg.JWT.TokenTTL)).Unix(),
		"iat": time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(s.cfg.JWT.SecretKey))
}

// CreateUser creates a new user.
func (s *Service) CreateUser(username string, password string) error {
	// Check if user already exists
	_, err := s.store.GetUserByUsername(username)
	if err == nil {
		return errors.New("user already exists")
	}
	if err.Error() != "user not found" {
		return err
	}

	// Get the default "developer" role
	role, err := s.store.GetRoleByName("developer")
	if err != nil {
		return fmt.Errorf("could not find default role 'developer': %w", err)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		RoleID:       role.ID,
		CreatedAt:    time.Now(),
	}

	return s.store.CreateUser(user)
}

// GetUserByUsername retrieves a user by username.
func (s *Service) GetUserByUsername(username string) (*models.User, error) {
	return s.store.GetUserByUsername(username)
}

// GetUserWithRole retrieves a user and their role by username.
func (s *Service) GetUserWithRole(username string) (*models.User, *models.Role, error) {
	return s.store.GetUserWithRole(username)
}

// CheckPassword checks if the provided password is correct for the user.
func (s *Service) CheckPassword(user *models.User, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
}

// GetJWTSecret returns the JWT secret key.
func (s *Service) GetJWTSecret() []byte {
	return []byte(s.cfg.JWT.SecretKey)
}
