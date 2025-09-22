package node

import (
	"fmt"
	"sync"

	"utopia-server/internal/models"
)

// Store 定义了节点数据的持久化接口。
type Store interface {
	CreateNode(node *models.Node) error
	GetNode(id string) (*models.Node, error)
	ListNodes() ([]*models.Node, error)
}

// memStore 是 Store 接口的一个内存实现，主要用于测试。
type memStore struct {
	mu    sync.RWMutex
	nodes map[string]*models.Node
}

// NewMemStore 创建一个新的 memStore 实例。
func NewMemStore() Store {
	return &memStore{
		nodes: make(map[string]*models.Node),
	}
}

// CreateNode 将一个新节点存储在内存中。
func (s *memStore) CreateNode(node *models.Node) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.nodes[node.ID]; exists {
		return fmt.Errorf("node with id %s already exists", node.ID)
	}
	s.nodes[node.ID] = node
	return nil
}

// GetNode 从内存中检索一个节点。
func (s *memStore) GetNode(id string) (*models.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	node, exists := s.nodes[id]
	if !exists {
		return nil, fmt.Errorf("node with id %s not found", id)
	}
	return node, nil
}

// ListNodes returns all nodes from the store.
func (s *memStore) ListNodes() ([]*models.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	nodes := make([]*models.Node, 0, len(s.nodes))
	for _, node := range s.nodes {
		nodes = append(nodes, node)
	}
	return nodes, nil
}
