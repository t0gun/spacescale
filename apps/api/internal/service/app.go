package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/t0gun/paas/internal/contracts"
	"github.com/t0gun/paas/internal/domain"
)

// AppService coordinates application and deployment operations.
// It uses a store for persistence and an optional runtime for deployments.
type AppService struct {
	store   contracts.Store
	runtime contracts.Runtime
}

// NewAppService builds a service with only a store.
// Runtime operations will be unavailable when using this constructor.
func NewAppService(store contracts.Store) *AppService {
	return &AppService{store: store}
}

// NewAppServiceWithRuntime builds a service with both a store and a runtime.
// Use this when deployment processing is required.
func NewAppServiceWithRuntime(store contracts.Store, rt contracts.Runtime) *AppService {
	return &AppService{store: store, runtime: rt}
}

// CreateAppParams collects the input needed to create a new application.
// Validation is performed in the domain constructor.
type CreateAppParams struct {
	Name  string
	Image string
	Port  int
}

// CreateApp validates input, constructs a domain app, and persists it.
// It maps storage conflict errors to a service level conflict for the API layer.
func (s *AppService) CreateApp(ctx context.Context, p CreateAppParams) (domain.App, error) {
	// Build and validate the domain object first to keep rules in one place.
	app, err := domain.NewApp(domain.NewAppParams{
		Name:  p.Name,
		Image: p.Image,
		Port:  p.Port,
	})
	if err != nil {
		return domain.App{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Persist the app and translate store conflicts to service conflicts.
	if err := s.store.CreateApp(ctx, app); err != nil {
		if errors.Is(err, contracts.ErrConflict) {
			return domain.App{}, ErrConflict
		}
		return domain.App{}, err
	}

	return app, nil
}

// ListApps returns all apps from the store without additional filtering.
// The store is responsible for the query behavior.
func (s *AppService) ListApps(ctx context.Context) ([]domain.App, error) {
	return s.store.ListApps(ctx)
}

// GetAppByID validates the input and fetches a single app by its identifier.
// Missing records are mapped to a service level not found error.
func (s *AppService) GetAppByID(ctx context.Context, id string) (domain.App, error) {
	if id == "" {
		return domain.App{}, ErrInvalidInput
	}
	app, err := s.store.GetAppByID(ctx, id)
	if err != nil {
		if errors.Is(err, contracts.ErrNotFound) {
			return domain.App{}, ErrNotFound
		}
		return domain.App{}, err
	}
	return app, nil
}
