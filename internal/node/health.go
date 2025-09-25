package node

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"utopia-server/internal/config"
	"utopia-server/internal/models"
)

// HealthCheckService 定期轮询节点健康状况。
type HealthCheckService struct {
	store  Store
	config config.FRPConfig
}

// NewHealthCheckService 创建一个新的 HealthCheckService 实例。
func NewHealthCheckService(store Store, cfg config.FRPConfig) *HealthCheckService {
	return &HealthCheckService{store: store, config: cfg}
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
	url := fmt.Sprintf("http://localhost:%d/api/v1/metrics", node.ControlPort)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Printf("failed to create health check request for node %d: %v", node.ID, err)
		return
	}
	req.Header.Set("Authorization", "Bearer "+s.config.AgentToken)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Node %s (%d) is offline: %v", node.Hostname, node.ID, err)
		node.Status = models.NodeStatusOffline
		node.ControlPort = 0 // 清空控制端口
		node.LastSeen = time.Now()
		if err := s.store.UpdateNode(node); err != nil {
			log.Printf("Error updating node %d to offline: %v", node.ID, err)
		}
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Node %s (%d) returned non-OK status: %s", node.Hostname, node.ID, resp.Status)
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
		log.Printf("Error updating node %d with new status: %v", node.ID, err)
	}
}
