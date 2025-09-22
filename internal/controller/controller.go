package controller

import (
	"log"
	"time"

	"utopia-server/internal/models"
	"utopia-server/internal/scheduler"
)

// Controller is the main reconciliation loop for the system.
type Controller struct {
	store     GpuClaimStore
	scheduler *scheduler.Scheduler
}

// NewController creates a new controller.
func NewController(store GpuClaimStore, scheduler *scheduler.Scheduler) *Controller {
	return &Controller{
		store:     store,
		scheduler: scheduler,
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
	claims, err := c.store.ListPendingGpuClaims()
	if err != nil {
		log.Printf("Error listing pending GPU claims: %v", err)
		return
	}

	for _, claim := range claims {
		c.reconcile(claim)
	}
}

func (c *Controller) reconcile(claim *models.GpuClaim) {
	log.Printf("Reconciling GpuClaim %s in phase %s", claim.ID, claim.Status.Phase)

	if claim.Status.Phase == models.GpuClaimPhasePending {
		node, err := c.scheduler.Schedule(claim)
		if err != nil {
			log.Printf("Failed to schedule GpuClaim %s: %v", claim.ID, err)
			return // Keep it in Pending, will retry
		}

		log.Printf("GpuClaim %s scheduled to node %s", claim.ID, node.ID)
		claim.Status.Phase = models.GpuClaimPhaseScheduled
		claim.Status.NodeName = node.ID

		if err := c.store.UpdateGpuClaim(claim); err != nil {
			log.Printf("Failed to update GpuClaim %s after scheduling: %v", claim.ID, err)
			// If update fails, we might have a dangling resource.
			// For now, we just log it. The claim will be reconciled again.
		}
	}
}
