package scheduler

import (
	"errors"
	"utopia-server/internal/models"
)

var (
	// ErrNoSuitableNodeFound is returned when no node can satisfy the claim's requirements.
	ErrNoSuitableNodeFound = errors.New("no suitable node found")
)

// Scheduler decides which node a GpuClaim should be assigned to.
type Scheduler struct {
	nodeStore NodeStore
}

// NewScheduler creates a new Scheduler.
func NewScheduler(nodeStore NodeStore) *Scheduler {
	return &Scheduler{
		nodeStore: nodeStore,
	}
}

// Schedule finds a suitable node for the given GpuClaim.
// The current algorithm is a simple first-fit: it finds the first online node
// that has enough available GPUs to satisfy the claim.
func (s *Scheduler) Schedule(claim *models.GpuClaim) (*models.Node, error) {
	nodes, err := s.nodeStore.ListNodes()
	if err != nil {
		return nil, err
	}

	requiredGpuCount := claim.Spec.Resources.GpuCount

	for _, node := range nodes {
		if node.Status != "Online" {
			continue
		}

		availableGpuCount := 0
		for _, gpu := range node.Gpus {
			if !gpu.Busy {
				availableGpuCount++
			}
		}

		if availableGpuCount >= requiredGpuCount {
			return node, nil
		}
	}

	return nil, ErrNoSuitableNodeFound
}
