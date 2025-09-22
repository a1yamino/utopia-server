package controller

import (
	"log"
	"time"

	"utopia-server/internal/client"
	"utopia-server/internal/models"
	"utopia-server/internal/node"
	"utopia-server/internal/scheduler"
)

// Controller is the main reconciliation loop for the system.
type Controller struct {
	store       GpuClaimStore
	scheduler   *scheduler.Scheduler
	nodeStore   node.Store
	agentClient *client.AgentClient
}

// NewController creates a new controller.
func NewController(store GpuClaimStore, scheduler *scheduler.Scheduler, nodeStore node.Store, agentClient *client.AgentClient) *Controller {
	return &Controller{
		store:       store,
		scheduler:   scheduler,
		nodeStore:   nodeStore,
		agentClient: agentClient,
	}
}

// Run starts the controller's reconciliation loop.
func (c *Controller) Run(stopCh <-chan struct{}) {
	log.Println("Starting controller")
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	go func() {
		for {
			select {
			case <-ticker.C:
				c.reconcileClaims()
			case <-stopCh:
				log.Println("Stopping controller")
				return
			}
		}
	}()
}

func (c *Controller) reconcileClaims() {
	claims, err := c.store.ListByPhase(models.GpuClaimPhasePending, models.GpuClaimPhaseScheduled)
	if err != nil {
		log.Printf("Error listing GPU claims: %v", err)
		return
	}

	for _, claim := range claims {
		c.reconcile(&claim)
	}
}

func (c *Controller) reconcile(claim *models.GpuClaim) {
	log.Printf("Reconciling GpuClaim %s in phase %s", claim.ID, claim.Status.Phase)

	switch claim.Status.Phase {
	case models.GpuClaimPhasePending:
		c.reconcilePending(claim)
	case models.GpuClaimPhaseScheduled:
		c.reconcileScheduled(claim)
	}
}

func (c *Controller) reconcilePending(claim *models.GpuClaim) {
	node, err := c.scheduler.Schedule(claim)
	if err != nil {
		log.Printf("Failed to schedule GpuClaim %s: %v", claim.ID, err)
		return // Keep it in Pending, will retry
	}

	log.Printf("GpuClaim %s scheduled to node %s", claim.ID, node.ID)
	claim.Status.Phase = models.GpuClaimPhaseScheduled
	claim.Status.NodeName = node.ID

	if err := c.store.Update(claim); err != nil {
		log.Printf("Failed to update GpuClaim %s after scheduling: %v", claim.ID, err)
	}
}

func (c *Controller) reconcileScheduled(claim *models.GpuClaim) {
	node, err := c.nodeStore.GetNode(claim.Status.NodeName)
	if err != nil {
		log.Printf("Failed to get node %s for GpuClaim %s: %v", claim.Status.NodeName, claim.ID, err)
		claim.Status.Phase = models.GpuClaimPhaseFailed
		claim.Status.Reason = "NodeNotFound"
		if err := c.store.Update(claim); err != nil {
			log.Printf("Failed to update GpuClaim %s to Failed: %v", claim.ID, err)
		}
		return
	}

	containerID, err := c.agentClient.CreateContainer(node, claim)
	if err != nil {
		log.Printf("Failed to create container for GpuClaim %s on node %s: %v", claim.ID, node.ID, err)
		claim.Status.Phase = models.GpuClaimPhaseFailed
		claim.Status.Reason = "ContainerCreationError"
		if err := c.store.Update(claim); err != nil {
			log.Printf("Failed to update GpuClaim %s to Failed: %v", claim.ID, err)
		}
		return
	}

	log.Printf("Container %s created for GpuClaim %s on node %s", containerID, claim.ID, node.ID)
	claim.Status.Phase = models.GpuClaimPhaseRunning
	claim.Status.ContainerID = containerID

	if err := c.store.Update(claim); err != nil {
		log.Printf("Failed to update GpuClaim %s to Running: %v", claim.ID, err)
	}
}
