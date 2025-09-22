package node

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
	"time"
	"utopia-server/internal/models"
)

// DiscoveryService 定期从 frps 发现节点隧道。
type DiscoveryService struct {
	frpsApiUrl string
	frpsUser   string
	frpsPass   string
	store      Store
}

// Run 启动发现服务。
func (s *DiscoveryService) Run(stopCh <-chan struct{}) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				s.discover()
			case <-stopCh:
				return
			}
		}
	}()
}

func (s *DiscoveryService) discover() {
	client := &http.Client{}
	req, err := http.NewRequest("GET", s.frpsApiUrl+"/api/proxy/tcp", nil)
	if err != nil {
		log.Printf("Error creating request: %v", err)
		return
	}
	req.SetBasicAuth(s.frpsUser, s.frpsPass)

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("Error getting tcp proxies: %v", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Printf("Error status code from frps: %d", resp.StatusCode)
		return
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Printf("Error reading response body: %v", err)
		return
	}

	var proxies []struct {
		Name       string `json:"name"`
		RemotePort int    `json:"remote_port"`
	}
	if err := json.Unmarshal(body, &proxies); err != nil {
		log.Printf("Error unmarshalling response: %v", err)
		return
	}

	for _, proxy := range proxies {
		if strings.HasSuffix(proxy.Name, "_control") {
			nodeID := strings.TrimSuffix(proxy.Name, "_control")
			s.updateNode(nodeID, proxy.RemotePort)
		}
	}
}

func (s *DiscoveryService) updateNode(nodeID string, controlPort int) {
	node, err := s.store.GetNode(nodeID)
	if err != nil {
		log.Printf("Error getting node %s: %v", nodeID, err)
		return
	}

	if node.ControlPort == controlPort && node.Status == models.NodeStatusOnline {
		return // No update needed
	}

	node.ControlPort = controlPort
	node.Status = models.NodeStatusOnline
	node.LastSeen = time.Now()

	if err := s.store.UpdateNode(node); err != nil {
		log.Printf("Error updating node %s: %v", nodeID, err)
	} else {
		log.Printf("Node %s is online, control port: %d", nodeID, controlPort)
	}
}

// NewDiscoveryService 创建一个新的 DiscoveryService 实例。
func NewDiscoveryService(frpsApiUrl, frpsUser, frpsPass string, store Store) *DiscoveryService {
	return &DiscoveryService{
		frpsApiUrl: frpsApiUrl,
		frpsUser:   frpsUser,
		frpsPass:   frpsPass,
		store:      store,
	}
}
