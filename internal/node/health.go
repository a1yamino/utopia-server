package node

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"utopia-server/internal/models"
)

// HealthCheckService 定期轮询节点健康状况。
type HealthCheckService struct {
	store Store
}

// NewHealthCheckService 创建一个新的 HealthCheckService 实例。
func NewHealthCheckService(store Store) *HealthCheckService {
	return &HealthCheckService{store: store}
}

// Run 启动健康检查轮询循环。
// 它会阻塞直到 stopCh 被关闭。
func (s *HealthCheckService) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	log.Println("Health check service started")

	for {
		select {
		case <-ticker.C:
			s.performCheck()
		case <-stopCh:
			log.Println("Health check service stopped")
			return
		}
	}
}

func (s *HealthCheckService) performCheck() {
	nodes, err := s.store.ListNodes()
	if err != nil {
		log.Printf("Error listing nodes for health check: %v", err)
		return
	}

	for _, node := range nodes {
		if node.Status == models.NodeStatusOnline {
			go s.checkNode(node)
		}
	}
}

func (s *HealthCheckService) checkNode(node *models.Node) {
	url := fmt.Sprintf("http://localhost:%d/status", node.ControlPort)
	resp, err := http.Get(url)
	if err != nil {
		log.Printf("Node %s (%s) is offline: %v", node.Hostname, node.ID, err)
		node.Status = models.NodeStatusOffline
		node.ControlPort = 0 // 清空控制端口
		node.LastSeen = time.Now()
		if err := s.store.UpdateNode(node); err != nil {
			log.Printf("Error updating node %s to offline: %v", node.ID, err)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Node %s (%s) returned non-OK status: %s", node.Hostname, node.ID, resp.Status)
		return
	}

	var status struct {
		Gpus []models.GpuInfo `json:"gpus"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		log.Printf("Error decoding status from node %s (%s): %v", node.Hostname, node.ID, err)
		return
	}

	node.Gpus = status.Gpus
	node.LastSeen = time.Now()
	if err := s.store.UpdateNode(node); err != nil {
		log.Printf("Error updating node %s with new status: %v", node.ID, err)
	}
}
