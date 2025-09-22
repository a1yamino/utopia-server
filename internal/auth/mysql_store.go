package auth

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"utopia-server/internal/models"
)

type mysqlStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) Store {
	return &mysqlStore{db: db}
}

func (s *mysqlStore) CreateUser(user *models.User) error {
	query := "INSERT INTO users (username, password_hash, role_id, created_at) VALUES (?, ?, ?, ?)"
	_, err := s.db.Exec(query, user.Username, user.PasswordHash, user.RoleID, user.CreatedAt)
	if err != nil {
		return fmt.Errorf("failed to create user: %w", err)
	}
	return nil
}

func (s *mysqlStore) GetUserByUsername(username string) (*models.User, error) {
	query := "SELECT id, username, password_hash, role_id, created_at FROM users WHERE username = ?"
	row := s.db.QueryRow(query, username)

	var user models.User
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.RoleID, &user.CreatedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, fmt.Errorf("failed to get user by username: %w", err)
	}
	return &user, nil
}

func (s *mysqlStore) GetUserWithRole(username string) (*models.User, *models.Role, error) {
	query := `
		SELECT u.id, u.username, u.password_hash, u.created_at, r.id, r.name, r.policies
		FROM users u
		JOIN roles r ON u.role_id = r.id
		WHERE u.username = ?
	`
	row := s.db.QueryRow(query, username)

	var user models.User
	var role models.Role
	err := row.Scan(&user.ID, &user.Username, &user.PasswordHash, &user.CreatedAt, &role.ID, &role.Name, &role.Policies)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil, fmt.Errorf("user not found")
		}
		return nil, nil, fmt.Errorf("failed to get user with role: %w", err)
	}
	user.Role = role
	user.RoleID = role.ID
	return &user, &role, nil
}

func (s *mysqlStore) GetRoleByName(name string) (*models.Role, error) {
	query := "SELECT id, name, policies FROM roles WHERE name = ?"
	row := s.db.QueryRow(query, name)

	var role models.Role
	var policies []byte
	err := row.Scan(&role.ID, &role.Name, &policies)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("role not found")
		}
		return nil, fmt.Errorf("failed to get role by name: %w", err)
	}

	// Unmarshal policies from JSON
	if err := json.Unmarshal(policies, &role.Policies); err != nil {
		return nil, fmt.Errorf("failed to unmarshal role policies: %w", err)
	}

	return &role, nil
}
