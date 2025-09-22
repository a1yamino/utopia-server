package node

import (
	"time"

	"utopia-server/internal/models"

	"github.com/google/uuid"
)

// Service 封装了节点管理的业务逻辑。
type Service struct {
	store Store
}

// NewService 创建一个新的节点服务实例。
func NewService(store Store) *Service {
	return &Service{
		store: store,
	}
}

// CreateNode 创建一个新节点。
func (s *Service) CreateNode(hostname string) (*models.Node, error) {
	newNode := &models.Node{
		ID:       "node-" + uuid.New().String(),
		Hostname: hostname,
		Status:   "Registering",
		LastSeen: time.Now(),
	}

	if err := s.store.CreateNode(newNode); err != nil {
		return nil, err
	}

	return newNode, nil
}
