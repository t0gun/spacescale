// Service logic for app creation and lookup
// This file maps input into domain models
// It stores apps and handles conflict errors
// It exposes methods used by api handlers
// Runtime support is optional in this service

package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/t0gun/paas/internal/contracts"
	"github.com/t0gun/paas/internal/domain"
)

// AppService coordinates application and deployment operations
// It uses a store for persistence and an optional runtime for deployments
type AppService struct {
	store   contracts.Store
	runtime contracts.Runtime
}

// This function handles new app service
// It supports new app service behavior
func NewAppService(store contracts.Store) *AppService {
	return &AppService{store: store}
}

// This function handles new app service with runtime
// It supports new app service with runtime behavior
func NewAppServiceWithRuntime(store contracts.Store, rt contracts.Runtime) *AppService {
	return &AppService{store: store, runtime: rt}
}

// CreateAppParams collects the input needed to create a new application
// Validation is performed in the domain constructor
type CreateAppParams struct {
	Name   string
	Image  string
	Port   *int
	Expose *bool
	Env    map[string]string
}

// This function handles create app
// It supports create app behavior
func (s *AppService) CreateApp(ctx context.Context, p CreateAppParams) (domain.App, error) {
	// Build and validate the domain object first to keep rules in one place
	app, err := domain.NewApp(domain.NewAppParams{
		Name:   p.Name,
		Image:  p.Image,
		Port:   p.Port,
		Expose: p.Expose,
		Env:    p.Env,
	})
	if err != nil {
		return domain.App{}, fmt.Errorf("%w: %v", ErrInvalidInput, err)
	}

	// Persist the app and translate store conflicts to service conflicts
	if err := s.store.CreateApp(ctx, app); err != nil {
		if errors.Is(err, contracts.ErrConflict) {
			return domain.App{}, ErrConflict
		}
		return domain.App{}, err
	}

	return app, nil
}

// This function handles list apps
// It supports list apps behavior
func (s *AppService) ListApps(ctx context.Context) ([]domain.App, error) {
	return s.store.ListApps(ctx)
}

// This function handles get app by id
// It supports get app by id behavior
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
