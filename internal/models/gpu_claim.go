package models

import "time"

// GpuClaimPhase represents the phase of a GpuClaim.
type GpuClaimPhase string

const (
	// GpuClaimPhasePending means the claim has been accepted by the system, but not yet scheduled.
	GpuClaimPhasePending GpuClaimPhase = "Pending"
	// GpuClaimPhaseScheduled means the claim has been scheduled to a node.
	GpuClaimPhaseScheduled GpuClaimPhase = "Scheduled"
	// GpuClaimPhaseRunning means the container for the claim is running on a node.
	GpuClaimPhaseRunning GpuClaimPhase = "Running"
	// GpuClaimPhaseFailed means the claim has failed.
	GpuClaimPhaseFailed GpuClaimPhase = "Failed"
	// GpuClaimPhaseCompleted means the claim has been completed.
	GpuClaimPhaseCompleted GpuClaimPhase = "Completed"
)

// GpuClaimSpec 定义了用户对 GPU 资源的期望状态。
type GpuClaimSpec struct {
	Image     string `json:"image"`
	Resources struct {
		GpuCount int `json:"gpuCount"`
	} `json:"resources"`
}

// GpuClaimStatus 定义了 GPU 资源的实际状态。
type GpuClaimStatus struct {
	Phase       GpuClaimPhase `json:"phase"`            // Pending, Running, Failed, Completed
	NodeName    string        `json:"nodeName"`         // 被调度到的节点名称
	ContainerID string        `json:"containerId"`      // 在节点上运行的容器 ID
	AccessURL   string        `json:"accessUrl"`        // 容器的公网访问地址
	Reason      string        `json:"reason,omitempty"` // 当 claim 失败时的原因
}

// GpuClaim 是一个声明式的 API 对象，用于描述对 GPU 资源的需求。
type GpuClaim struct {
	ID        string         `json:"id" gorm:"primaryKey"`
	UserID    string         `json:"userId"`
	CreatedAt time.Time      `json:"createdAt"`
	Spec      GpuClaimSpec   `json:"spec" gorm:"embedded"`
	Status    GpuClaimStatus `json:"status" gorm:"embedded"`
}
