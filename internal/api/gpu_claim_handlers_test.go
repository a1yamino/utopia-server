package api

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"database/sql"
	"utopia-server/internal/auth"
	"utopia-server/internal/client"
	"utopia-server/internal/config"
	"utopia-server/internal/controller"
	"utopia-server/internal/database"
	"utopia-server/internal/models"
	"utopia-server/internal/node"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// Helper function to create a user and get a token
func setupUserAndGetToken(t *testing.T, authService *auth.Service, authStore auth.Store, db *sql.DB, serverURL string) (models.User, string) {
	// Get developer role
	var devRole models.Role
	var policies []byte
	err := db.QueryRow("SELECT id, name, policies FROM roles WHERE name = ?", "developer").Scan(&devRole.ID, &devRole.Name, &policies)
	assert.NoError(t, err, "Failed to find developer role")
	assert.NotZero(t, devRole.ID, "Developer role ID should not be zero")

	// Create a test user with developer role
	password := "password123"
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	assert.NoError(t, err)

	user := models.User{
		Username:     "testdev",
		PasswordHash: string(hashedPassword),
		RoleID:       devRole.ID,
		CreatedAt:    time.Now(),
	}
	err = authStore.CreateUser(&user)
	assert.NoError(t, err)

	// Login to get token
	loginPayload := map[string]string{
		"username": user.Username,
		"password": password,
	}
	loginBody, _ := json.Marshal(loginPayload)
	loginReq, _ := http.NewRequest(http.MethodPost, serverURL+"/api/auth/login", bytes.NewBuffer(loginBody))
	loginReq.Header.Set("Content-Type", "application/json")

	loginResp, err := http.DefaultClient.Do(loginReq)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, loginResp.StatusCode)

	var loginRespBody map[string]string
	err = json.NewDecoder(loginResp.Body).Decode(&loginRespBody)
	assert.NoError(t, err)
	token := loginRespBody["token"]
	assert.NotEmpty(t, token)

	// Retrieve full user object with role
	fullUser, _, err := authStore.GetUserWithRole(user.Username)
	assert.NoError(t, err)

	return *fullUser, token
}

func TestGpuClaimAPI(t *testing.T) {
	// Setup Test DB
	testDB, dbName := database.SetupTestDB(t)
	defer database.TeardownTestDB(testDB, dbName)

	// Load config
	cfg, err := config.Load()
	require.NoError(t, err, "Failed to load config")

	// Override DSN to use the test database
	cfg.Database.DSN = fmt.Sprintf("root:password@tcp(127.0.0.1:3306)/%s?parseTime=true", dbName)
	cfg.JWT.SecretKey = "test-secret"
	cfg.JWT.TokenTTL = int(time.Hour.Seconds())

	// Create real dependencies
	authStore := auth.NewMySQLStore(testDB)
	gpuClaimStore := controller.NewMySQLStore(testDB)
	nodeStore := node.NewMySQLStore(testDB)
	authService := auth.NewService(authStore, cfg)
	nodeService := node.NewService(nodeStore)
	agentClient := client.NewAgentClient(cfg.FRP)
	server := NewServer(cfg.Server, authService, nodeService, gpuClaimStore, agentClient)

	// Start a test server
	testServer := httptest.NewServer(server.Router)
	defer testServer.Close()

	// Setup user and get token
	developerUser, developerToken := setupUserAndGetToken(t, authService, authStore, testDB, testServer.URL)

	t.Run("Create Claim Success", func(t *testing.T) {
		claimSpec := models.GpuClaimSpec{
			Image: "ubuntu:20.04",
			Resources: struct {
				GpuCount int `json:"gpuCount"`
			}{GpuCount: 1},
		}
		claimBody, _ := json.Marshal(claimSpec)
		req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/gpu-claims", bytes.NewBuffer(claimBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+developerToken)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)

		// Verify in database
		var createdClaim models.GpuClaim
		var spec, status []byte
		err = testDB.QueryRow("SELECT id, user_id, created_at, spec, status FROM gpu_claims WHERE user_id = ?", developerUser.Username).Scan(&createdClaim.ID, &createdClaim.UserID, &createdClaim.CreatedAt, &spec, &status)
		assert.NoError(t, err)

		err = json.Unmarshal(spec, &createdClaim.Spec)
		assert.NoError(t, err)
		err = json.Unmarshal(status, &createdClaim.Status)
		assert.NoError(t, err)

		assert.Equal(t, developerUser.Username, createdClaim.UserID)
		assert.Equal(t, claimSpec.Image, createdClaim.Spec.Image)
		assert.Equal(t, claimSpec.Resources.GpuCount, createdClaim.Spec.Resources.GpuCount)
		assert.Equal(t, models.GpuClaimPhasePending, createdClaim.Status.Phase)
	})

	t.Run("Create Claim Exceeds Quota", func(t *testing.T) {
		claimSpec := models.GpuClaimSpec{
			Image: "ubuntu:20.04",
			Resources: struct {
				GpuCount int `json:"gpuCount"`
			}{GpuCount: 3}, // Exceeds developer quota of 2
		}
		claimBody, _ := json.Marshal(claimSpec)
		req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/gpu-claims", bytes.NewBuffer(claimBody))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+developerToken)

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	})

	t.Run("Create Claim Unauthenticated", func(t *testing.T) {
		claimSpec := models.GpuClaimSpec{
			Image: "ubuntu:20.04",
			Resources: struct {
				GpuCount int `json:"gpuCount"`
			}{GpuCount: 1},
		}
		claimBody, _ := json.Marshal(claimSpec)
		req, _ := http.NewRequest(http.MethodPost, testServer.URL+"/api/gpu-claims", bytes.NewBuffer(claimBody))
		req.Header.Set("Content-Type", "application/json")
		// No Authorization header

		resp, err := http.DefaultClient.Do(req)
		assert.NoError(t, err)
		assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)
	})
}
