package auth

import (
	"database/sql"
	"testing"
	"utopia-server/internal/config"
	"utopia-server/internal/models"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"golang.org/x/crypto/bcrypt"
)

func TestCreateUser_Success(t *testing.T) {
	mockStore := new(MockStore)
	cfg := &config.Config{JWT: config.JWTConfig{SecretKey: "secret", TokenTTL: 3600}}
	service := NewService(mockStore, cfg)

	mockStore.On("GetUserWithRole", "newuser").Return(nil, nil, sql.ErrNoRows)
	mockStore.On("CreateUser", mock.AnythingOfType("*models.User")).Return(nil)

	err := service.CreateUser("newuser", "password")

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)
}

func TestCreateUser_UserAlreadyExists(t *testing.T) {
	mockStore := new(MockStore)
	cfg := &config.Config{JWT: config.JWTConfig{SecretKey: "secret", TokenTTL: 3600}}
	service := NewService(mockStore, cfg)

	mockStore.On("GetUserWithRole", "existinguser").Return(&models.User{}, &models.Role{}, nil)

	err := service.CreateUser("existinguser", "password")

	assert.Error(t, err)
	mockStore.AssertExpectations(t)
}

func TestPasswordHashing(t *testing.T) {
	mockStore := new(MockStore)
	cfg := &config.Config{JWT: config.JWTConfig{SecretKey: "secret", TokenTTL: 3600}}
	service := NewService(mockStore, cfg)

	password := "password"
	var capturedUser *models.User

	mockStore.On("GetUserWithRole", "newuser").Return(nil, nil, sql.ErrNoRows)
	mockStore.On("CreateUser", mock.AnythingOfType("*models.User")).Run(func(args mock.Arguments) {
		capturedUser = args.Get(0).(*models.User)
	}).Return(nil)

	err := service.CreateUser("newuser", password)

	assert.NoError(t, err)
	mockStore.AssertExpectations(t)

	assert.NotNil(t, capturedUser)
	err = bcrypt.CompareHashAndPassword([]byte(capturedUser.PasswordHash), []byte(password))
	assert.NoError(t, err, "Password hash does not match original password")
}
