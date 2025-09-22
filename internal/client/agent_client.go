package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"utopia-server/internal/models"
)

type AgentClient struct {
	httpClient *http.Client
}

func NewAgentClient() *AgentClient {
	return &AgentClient{
		httpClient: &http.Client{},
	}
}

func (c *AgentClient) CreateContainer(node *models.Node, claim *models.GpuClaim) (string, error) {
	url := fmt.Sprintf("http://localhost:%d/containers", node.ControlPort)

	body, err := json.Marshal(claim.Spec)
	if err != nil {
		return "", fmt.Errorf("failed to marshal claim spec: %w", err)
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("failed to create container, status code: %d", resp.StatusCode)
	}

	var result struct {
		ContainerID string `json:"container_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("failed to decode response: %w", err)
	}

	return result.ContainerID, nil
}
