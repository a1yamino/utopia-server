package models

import "time"

const (
	NodeStatusOnline      = "Online"
	NodeStatusOffline     = "Offline"
	NodeStatusRegistering = "Registering"
)

// GpuInfo 描述了节点上单个 GPU 的信息。
type GpuInfo struct {
	ID            int    `json:"id"`
	TemperatureC  int    `json:"temperature_c"`
	MemoryTotalMB int    `json:"memory_total_mb"`
	MemoryUsedMB  int    `json:"memory_used_mb"`
	Name          string `json:"name"`
	UUID          string `json:"uuid"`
	Busy          bool   `json:"busy"`
	UsagePercent  int    `json:"usage_percent"`
	ContainerID   string `json:"containerId,omitempty"` // This field is not in the new JSON, but we might need it.
}

// SystemMetrics 代表节点的系统级指标。
type SystemMetrics struct {
	CPUUsagePercent    float64 `json:"cpu_usage_percent"`
	MemoryUsagePercent float64 `json:"memory_usage_percent"`
	MemoryTotalMB      int     `json:"memory_total_mb"`
	MemoryUsedMB       int     `json:"memory_used_mb"`
	DiskUsagePercent   int     `json:"disk_usage_percent"`
	LoadAverage        float64 `json:"load_average"`
	Uptime             int     `json:"uptime"`
}

// Node 代表一个计算节点，可以承载 GPU 工作负载。
type Node struct {
	ID          int64     `json:"id" gorm:"primaryKey"`
	Hostname    string    `json:"hostname"`
	Status      string    `json:"status"` // Online, Offline, Registering
	Gpus        []GpuInfo `json:"gpus" gorm:"type:json"`
	ControlPort int       `json:"controlPort"`
	LastSeen    time.Time `json:"lastSeen"`
}

// NodeMetrics 代表从节点 agent 返回的完整指标。
type NodeMetrics struct {
	NodeID             string        `json:"node_id"`
	CPUUsagePercent    float64       `json:"cpu_usage_percent"`
	MemoryUsagePercent float64       `json:"memory_usage_percent"`
	Gpus               []GpuInfo     `json:"gpus"`
	System             SystemMetrics `json:"system"`
}
