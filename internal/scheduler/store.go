package scheduler

import "utopia-server/internal/models"

// NodeStore defines the interface for accessing node data.
type NodeStore interface {
	ListNodes() ([]*models.Node, error)
}
