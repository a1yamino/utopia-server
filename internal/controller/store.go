package controller

import (
	"sync"

	"utopia-server/internal/models"
)

// GpuClaimStore defines the interface for GPU claim storage.
type GpuClaimStore interface {
	CreateGpuClaim(claim *models.GpuClaim) error
	ListPendingGpuClaims() ([]*models.GpuClaim, error)
	ListByPhase(phases ...models.GpuClaimPhase) ([]models.GpuClaim, error)
	Update(claim *models.GpuClaim) error
}

// memStore is an in-memory implementation of GpuClaimStore for testing.
type memStore struct {
	mu     sync.RWMutex
	claims map[string]*models.GpuClaim
}

// NewMemStore creates a new in-memory GpuClaimStore.
func NewMemStore() GpuClaimStore {
	return &memStore{
		claims: make(map[string]*models.GpuClaim),
	}
}

// CreateGpuClaim adds a new GPU claim to the store.
func (s *memStore) CreateGpuClaim(claim *models.GpuClaim) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.claims[claim.ID] = claim
	return nil
}

// ListPendingGpuClaims returns all claims with a "Pending" status.
func (s *memStore) ListPendingGpuClaims() ([]*models.GpuClaim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var pendingClaims []*models.GpuClaim
	for _, claim := range s.claims {
		if claim.Status.Phase == models.GpuClaimPhasePending {
			pendingClaims = append(pendingClaims, claim)
		}
	}
	return pendingClaims, nil
}

func (s *memStore) ListByPhase(phases ...models.GpuClaimPhase) ([]models.GpuClaim, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []models.GpuClaim
	for _, claim := range s.claims {
		for _, phase := range phases {
			if claim.Status.Phase == phase {
				result = append(result, *claim)
				break
			}
		}
	}
	return result, nil
}

// Update updates an existing GPU claim in the store.
func (s *memStore) Update(claim *models.GpuClaim) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.claims[claim.ID] = claim
	return nil
}
