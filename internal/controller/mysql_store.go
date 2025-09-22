package controller

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"utopia-server/internal/models"
)

type mysqlStore struct {
	db *sql.DB
}

func NewMySQLStore(db *sql.DB) GpuClaimStore {
	return &mysqlStore{db: db}
}

func (s *mysqlStore) CreateGpuClaim(claim *models.GpuClaim) error {
	spec, err := json.Marshal(claim.Spec)
	if err != nil {
		return fmt.Errorf("failed to marshal spec: %w", err)
	}

	status, err := json.Marshal(claim.Status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	query := "INSERT INTO gpu_claims (id, user_id, created_at, spec, status) VALUES (?, ?, ?, ?, ?)"
	_, err = s.db.Exec(query, claim.ID, claim.UserID, claim.CreatedAt, spec, status)
	if err != nil {
		return fmt.Errorf("failed to create gpu claim: %w", err)
	}
	return nil
}

func (s *mysqlStore) ListPendingGpuClaims() ([]*models.GpuClaim, error) {
	query := `SELECT id, user_id, created_at, spec, status FROM gpu_claims WHERE status->>"$.phase" = 'Pending'`
	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list pending gpu claims: %w", err)
	}
	defer rows.Close()

	var claims []*models.GpuClaim
	for rows.Next() {
		var claim models.GpuClaim
		var spec, status []byte
		err := rows.Scan(&claim.ID, &claim.UserID, &claim.CreatedAt, &spec, &status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan gpu claim: %w", err)
		}

		if err := json.Unmarshal(spec, &claim.Spec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
		}
		if err := json.Unmarshal(status, &claim.Status); err != nil {
			return nil, fmt.Errorf("failed to unmarshal status: %w", err)
		}
		claims = append(claims, &claim)
	}

	return claims, nil
}

func (s *mysqlStore) Update(claim *models.GpuClaim) error {
	status, err := json.Marshal(claim.Status)
	if err != nil {
		return fmt.Errorf("failed to marshal status: %w", err)
	}

	query := "UPDATE gpu_claims SET status = ? WHERE id = ?"
	_, err = s.db.Exec(query, status, claim.ID)
	if err != nil {
		return fmt.Errorf("failed to update gpu claim: %w", err)
	}
	return nil
}

func (s *mysqlStore) ListByPhase(phases ...models.GpuClaimPhase) ([]models.GpuClaim, error) {
	if len(phases) == 0 {
		return []models.GpuClaim{}, nil
	}

	query := `SELECT id, user_id, created_at, spec, status FROM gpu_claims WHERE status->>"$.phase" IN (`
	args := make([]interface{}, len(phases))
	placeholders := ""
	for i, phase := range phases {
		args[i] = phase
		placeholders += "?"
		if i < len(phases)-1 {
			placeholders += ","
		}
	}
	query += placeholders + ")"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list gpu claims by phase: %w", err)
	}
	defer rows.Close()

	var claims []models.GpuClaim
	for rows.Next() {
		var claim models.GpuClaim
		var spec, status []byte
		err := rows.Scan(&claim.ID, &claim.UserID, &claim.CreatedAt, &spec, &status)
		if err != nil {
			return nil, fmt.Errorf("failed to scan gpu claim: %w", err)
		}

		if err := json.Unmarshal(spec, &claim.Spec); err != nil {
			return nil, fmt.Errorf("failed to unmarshal spec: %w", err)
		}
		if err := json.Unmarshal(status, &claim.Status); err != nil {
			return nil, fmt.Errorf("failed to unmarshal status: %w", err)
		}
		claims = append(claims, claim)
	}

	return claims, nil
}
