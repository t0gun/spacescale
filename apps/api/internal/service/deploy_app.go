package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/t0gun/paas/internal/contracts"
	"github.com/t0gun/paas/internal/domain"
)

// DeployAppParams contains the input needed to request a deployment.
type DeployAppParams struct {
	AppID string
}

// ListDeploymentsParams identifies which app to list deployments for.
type ListDeploymentsParams struct {
	AppID string
}

// DeployApp validates the request, ensures the app exists, and enqueues a deployment.
// The returned deployment starts in a queued state for processing by a worker.
func (s *AppService) DeployApp(ctx context.Context, p DeployAppParams) (domain.Deployment, error) {
	if p.AppID == "" {
		return domain.Deployment{}, fmt.Errorf("%w: app id is required", ErrInvalidInput)
	}

	// Ensure the app exists before creating a deployment record.
	_, err := s.store.GetAppByID(ctx, p.AppID)
	if err != nil {
		if errors.Is(err, contracts.ErrNotFound) {
			return domain.Deployment{}, ErrNotFound
		}
		return domain.Deployment{}, err
	}

	// Create a queued deployment record.
	dep := domain.NewDeployment(p.AppID)
	if err := s.store.CreateDeployment(ctx, dep); err != nil {
		if errors.Is(err, contracts.ErrNotFound) {
			// Store enforces that the app must exist.
			return domain.Deployment{}, ErrNotFound
		}
		return domain.Deployment{}, ErrNotFound
	}
	return dep, nil
}

// ProcessNextDeployment takes the next queued deployment and runs the runtime.
// It updates status transitions and records errors on the deployment when needed.
func (s *AppService) ProcessNextDeployment(ctx context.Context) (domain.Deployment, error) {
	if s.runtime == nil {
		return domain.Deployment{}, ErrNoRuntime
	}

	// Grab the next queued deployment in FIFO order.
	dep, err := s.store.TakeNextQueuedDeployment(ctx)
	if err != nil {
		if errors.Is(err, contracts.ErrNotFound) {
			return domain.Deployment{}, ErrNoWork
		}
		return domain.Deployment{}, err
	}

	// Mark deployment as building before interacting with the runtime.
	dep.Status = domain.DeploymentStatusBuilding
	dep.UpdatedAt = time.Now()
	if err := s.store.UpdateDeployment(ctx, dep); err != nil {
		return domain.Deployment{}, err
	}

	// Load app data needed for the runtime deployment.
	app, err := s.store.GetAppByID(ctx, dep.AppID)
	if err != nil {
		msg := err.Error()
		dep.Status = domain.DeploymentStatusFailed
		dep.Error = &msg
		dep.UpdatedAt = time.Now().UTC()
		_ = s.store.UpdateDeployment(ctx, dep)
		return dep, fmt.Errorf("runtime deploy failed: %w", err)
	}

	// Run the runtime deploy and capture a URL or an error.
	url, err := s.runtime.Deploy(ctx, app)
	if err != nil {
		msg := err.Error()
		dep.Status = domain.DeploymentStatusFailed
		dep.Error = &msg
		dep.UpdatedAt = time.Now().UTC()
		_ = s.store.UpdateDeployment(ctx, dep)
		return dep, fmt.Errorf("runtime deploy failed: %w", err)
	}

	// Mark deployment as running with the resolved URL.
	dep.Status = domain.DeploymentStatusRunning
	dep.URL = &url
	dep.Error = nil
	dep.UpdatedAt = time.Now().UTC()
	if err := s.store.UpdateDeployment(ctx, dep); err != nil {
		return domain.Deployment{}, err
	}

	return dep, nil

}

// ListDeployments validates input, ensures the app exists, and returns its deployments.
// This avoids returning an empty list for a missing app.
func (s *AppService) ListDeployments(ctx context.Context, p ListDeploymentsParams) ([]domain.Deployment, error) {
	if p.AppID == "" {
		return nil, ErrInvalidInput
	}

	// Ensure the app exists so missing apps return a not found error.
	if _, err := s.store.GetAppByID(ctx, p.AppID); err != nil {
		if errors.Is(err, contracts.ErrNotFound) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	deps, err := s.store.ListDeploymentsByAppID(ctx, p.AppID)
	if err != nil {
		return nil, err
	}
	return deps, nil
}
