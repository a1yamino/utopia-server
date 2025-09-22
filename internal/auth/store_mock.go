package auth

import (
	"utopia-server/internal/models"

	"github.com/stretchr/testify/mock"
)

// MockStore is a mock implementation of the Store interface.
type MockStore struct {
	mock.Mock
}

// CreateUser mocks the CreateUser method.
func (m *MockStore) CreateUser(user *models.User) error {
	args := m.Called(user)
	return args.Error(0)
}

// GetUserByUsername mocks the GetUserByUsername method.
func (m *MockStore) GetUserByUsername(username string) (*models.User, error) {
	args := m.Called(username)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*models.User), args.Error(1)
}

// GetUserWithRole mocks the GetUserWithRole method.
func (m *MockStore) GetUserWithRole(username string) (*models.User, *models.Role, error) {
	args := m.Called(username)
	var user *models.User
	if args.Get(0) != nil {
		user = args.Get(0).(*models.User)
	}
	var role *models.Role
	if args.Get(1) != nil {
		role = args.Get(1).(*models.Role)
	}
	return user, role, args.Error(2)
}

// GetRoleByName mocks the GetRoleByName method.
func (m *MockStore) GetRoleByName(name string) (*models.Role, error) {
	args := m.Called(name)
	var role *models.Role
	if args.Get(0) != nil {
		role = args.Get(0).(*models.Role)
	}
	return role, args.Error(1)
}
