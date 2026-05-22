package reconciler

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/Aneesh-382005/Orbital/internal/models"
	"github.com/Aneesh-382005/Orbital/internal/provisioner"
	"github.com/Aneesh-382005/Orbital/internal/store"
)

type Reconciler struct {
	store       *store.Store
	provisioner *provisioner.DockerProvisioner
	interval    time.Duration
	logger      *slog.Logger
}

func New(s *store.Store, p *provisioner.DockerProvisioner, interval time.Duration) *Reconciler {
	return &Reconciler{
		store:       s,
		provisioner: p,
		interval:    interval,
		logger:      slog.Default(),
	}
}

func (r *Reconciler) Start(ctx context.Context) {
	r.logger.Info("reconciler started", "interval", r.interval)
	ticker := time.NewTicker(r.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := r.reconcile(ctx); err != nil {
				r.logger.Error("reconcilation failed", "error", err)
			}
		case <-ctx.Done():
			r.logger.Info("reconciler stopped")
			return
		}
	}
}

func (r *Reconciler) reconcile(ctx context.Context) error {
	r.logger.Info("reconcilation cycle started")

	//Get all containers Docker knows about with our label
	actualContainers, err := r.provisioner.ListWorkspaceContainers(ctx)
	if err != nil {
		return fmt.Errorf("listing containers: %w", err)
	}

	//workspaceID -> containerID
	actualWorkspaceID := map[string]string{}
	for _, c := range actualContainers {
		wsID := c.Labels["orbital.workspace_id"]
		if wsID != "" {
			actualWorkspaceID[wsID] = c.ID
		}
	}

	//Get all workspaces DB thinks are running
	runningInDB, err := r.store.ListRunning()
	if err != nil {
		return fmt.Errorf("listing running workspaces: %w", err)
	}

	//Build map: workspaceID -> true
	desiredRunning := map[string]bool{}
	for _, ws := range runningInDB {
		desiredRunning[ws.ID] = true
	}

	//Case 1: DB says RUNNING but no container exists
	for _, ws := range runningInDB {
		if _, exists := actualWorkspaceID[ws.ID]; !exists {
			r.logger.Warn(
				"Container missing for running workspace",
				"workspace_id", ws.ID,
				"user_id", ws.UserID,
			)
			ws.Status = models.StatusError
			if err := r.store.Update(ws); err != nil {
				r.logger.Error(
					"failed to mark workspace as error",
					"workspace_id", ws.ID,
					"error", err,
				)
			}
		}
	}

	//Case 2: Container exists but DB has no RUNNING record for it
	for wsID, containerID := range actualWorkspaceID {
		if !desiredRunning[wsID] {
			r.logger.Warn(
				"orphaned container found, removing",
				"workspace_id", wsID,
				"container_id", containerID,
			)
			if err := r.provisioner.RemoveContainer(ctx, containerID); err != nil {
				r.logger.Error(
					"failed to remove orphaned container",
					"container_id", containerID,
					"error", err,
				)
			}
		}
	}

	r.logger.Info(
		"reconcilation cycle completed",
		"running_in_db", len(runningInDB),
		"actual_containers", len(actualContainers),
	)
	return nil
}
