package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"utopia-server/internal/auth"
	"utopia-server/internal/config"
	"utopia-server/internal/database"

	"github.com/stretchr/testify/assert"
)

func TestAuthAPI(t *testing.T) {
	// Setup Test DB
	testDB, dbName := database.SetupTestDB(t)
	defer database.TeardownTestDB(testDB, dbName)

	// Create real dependencies
	cfg := &config.Config{
		JWT: config.JWTConfig{
			SecretKey: "test-secret",
			TokenTTL:  int(time.Hour.Seconds()),
		},
		Database: config.DatabaseConfig{
			DSN: fmt.Sprintf("root:password@tcp(127.0.0.1:3306)/%s?parseTime=true", dbName),
		},
	}
	authStore := auth.NewMySQLStore(testDB)
	authService := auth.NewService(authStore, cfg)
	server := NewServer(cfg.Server, authService, nil, nil) // Pass nil for other services for now

	// Start a test server
	testServer := httptest.NewServer(server.Router)
	defer testServer.Close()

	// Sub-tests
	t.Run("Register and Login Success", func(t *testing.T) {
		// Register
		registerPayload := map[string]string{
			"username": "testuser",
			"password": "password123",
		}
		registerBody, _ := json.Marshal(registerPayload)
		registerReq, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/auth/register", bytes.NewBuffer(registerBody))
		registerReq.Header.Set("Content-Type", "application/json")

		registerResp, err := http.DefaultClient.Do(registerReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, registerResp.StatusCode)

		// Login
		loginPayload := map[string]string{
			"username": "testuser",
			"password": "password123",
		}
		loginBody, _ := json.Marshal(loginPayload)
		loginReq, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/auth/login", bytes.NewBuffer(loginBody))
		loginReq.Header.Set("Content-Type", "application/json")

		loginResp, err := http.DefaultClient.Do(loginReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusOK, loginResp.StatusCode)

		var loginRespBody map[string]string
		err = json.NewDecoder(loginResp.Body).Decode(&loginRespBody)
		assert.NoError(t, err)
		assert.NotEmpty(t, loginRespBody["token"])
	})

	t.Run("Register Existing User", func(t *testing.T) {
		// Register first user
		registerPayload := map[string]string{
			"username": "existinguser",
			"password": "password123",
		}
		registerBody, _ := json.Marshal(registerPayload)
		registerReq, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/auth/register", bytes.NewBuffer(registerBody))
		registerReq.Header.Set("Content-Type", "application/json")

		registerResp, err := http.DefaultClient.Do(registerReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, registerResp.StatusCode)

		// Try to register again
		registerAgainReq, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/auth/register", bytes.NewBuffer(registerBody))
		registerAgainReq.Header.Set("Content-Type", "application/json")

		registerAgainResp, err := http.DefaultClient.Do(registerAgainReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusConflict, registerAgainResp.StatusCode)
	})

	t.Run("Login with Wrong Password", func(t *testing.T) {
		// Register user
		registerPayload := map[string]string{
			"username": "wrongpassuser",
			"password": "password123",
		}
		registerBody, _ := json.Marshal(registerPayload)
		registerReq, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/auth/register", bytes.NewBuffer(registerBody))
		registerReq.Header.Set("Content-Type", "application/json")

		registerResp, err := http.DefaultClient.Do(registerReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusCreated, registerResp.StatusCode)

		// Login with wrong password
		loginPayload := map[string]string{
			"username": "wrongpassuser",
			"password": "wrongpassword",
		}
		loginBody, _ := json.Marshal(loginPayload)
		loginReq, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/auth/login", bytes.NewBuffer(loginBody))
		loginReq.Header.Set("Content-Type", "application/json")

		loginResp, err := http.DefaultClient.Do(loginReq)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, loginResp.StatusCode)
	})
}
