package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"

	"utopia-server/internal/config"
	"utopia-server/internal/models"
)

type AgentClient struct {
	httpClient *http.Client
	config     config.FRPConfig
}

func NewAgentClient(cfg config.FRPConfig) *AgentClient {
	return &AgentClient{
		httpClient: &http.Client{},
		config:     cfg,
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
	req.Header.Set("Authorization", "Bearer "+c.config.AgentToken)

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

func (c *AgentClient) GetNodeMetrics(node *models.Node) (map[string]interface{}, error) {
	url := fmt.Sprintf("http://localhost:%d/api/v1/metrics", node.ControlPort)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+c.config.AgentToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to get metrics, status code: %d", resp.StatusCode)
	}

	var metrics map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&metrics); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return metrics, nil
}
