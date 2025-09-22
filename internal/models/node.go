package models

import "time"

// GpuInfo 描述了节点上单个 GPU 的信息。
type GpuInfo struct {
	ID          string `json:"id"`                    // GPU 的唯一标识符
	Model       string `json:"model"`                 // GPU 型号
	Status      string `json:"status"`                // Available, InUse
	ContainerID string `json:"containerId,omitempty"` // 如果在使用中，关联的容器 ID
}

// Node 代表一个计算节点，可以承载 GPU 工作负载。
type Node struct {
	ID          string    `json:"id" gorm:"primaryKey"`
	Hostname    string    `json:"hostname"`
	Status      string    `json:"status"` // Online, Offline, Registering
	Gpus        []GpuInfo `json:"gpus" gorm:"type:json"`
	ControlPort int       `json:"controlPort"`
	LastSeen    time.Time `json:"lastSeen"`
}
