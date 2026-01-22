// Tests for in memory store behavior
// Tests include create list and fetch for apps
// Tests cover deployment queue and update flows
// Tests verify not found and conflict errors
// These tests ensure store data integrity

package store_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/t0gun/paas/internal/adapters/store"
	"github.com/t0gun/paas/internal/contracts"
	"github.com/t0gun/paas/internal/domain"
)

// This function handles test memory store create app ok
// It supports test memory store create app ok behavior
func TestMemoryStore_CreateApp_OK(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemoryStore()

	app, err := domain.NewApp(domain.NewAppParams{Name: "hello", Image: "nginx:latest", Port: ptrInt(8080)})
	assert.NoError(t, err)

	err = st.CreateApp(ctx, app)
	assert.NoError(t, err)
}

// This function handles test memory store create app duplicate name conflict
// It supports test memory store create app duplicate name conflict behavior
func TestMemoryStore_CreateApp_DuplicateName_Conflict(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemoryStore()

	app1, err := domain.NewApp(domain.NewAppParams{Name: "hello", Image: "nginx:latest", Port: ptrInt(8080)})
	assert.NoError(t, err)
	app2, err := domain.NewApp(domain.NewAppParams{Name: "hello", Image: "nginx:latest", Port: ptrInt(8080)})
	assert.NoError(t, err)

	assert.NoError(t, st.CreateApp(ctx, app1))

	err = st.CreateApp(ctx, app2)
	assert.Error(t, err)
	assert.ErrorIs(t, err, contracts.ErrConflict)
}

// This function handles test memory store get app by id not found
// It supports test memory store get app by id not found behavior
func TestMemoryStore_GetAppByID_NotFound(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemoryStore()

	_, err := st.GetAppByID(ctx, "missing")
	assert.Error(t, err)
	assert.ErrorIs(t, err, contracts.ErrNotFound)
}

// This function handles test memory store get app by name not found
// It supports test memory store get app by name not found behavior
func TestMemoryStore_GetAppByName_NotFound(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemoryStore()

	_, err := st.GetAppByName(ctx, "missing")
	assert.Error(t, err)
	assert.ErrorIs(t, err, contracts.ErrNotFound)
}

// This function handles test memory store list apps count
// It supports test memory store list apps count behavior
func TestMemoryStore_ListApps_Count(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemoryStore()

	a, err := domain.NewApp(domain.NewAppParams{Name: "a", Image: "nginx:latest", Port: ptrInt(8080)})
	assert.NoError(t, err)
	b, err := domain.NewApp(domain.NewAppParams{Name: "b", Image: "nginx:latest", Port: ptrInt(8080)})
	assert.NoError(t, err)

	assert.NoError(t, st.CreateApp(ctx, a))
	assert.NoError(t, st.CreateApp(ctx, b))

	apps, err := st.ListApps(ctx)
	assert.NoError(t, err)
	assert.Len(t, apps, 2)
}

// This function handles test memory store create deployment app missing not found
// It supports test memory store create deployment app missing not found behavior
func TestMemoryStore_CreateDeployment_AppMissing_NotFound(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemoryStore()

	dep := domain.NewDeployment("missing-app")

	err := st.CreateDeployment(ctx, dep)
	assert.Error(t, err)
	assert.ErrorIs(t, err, contracts.ErrNotFound)
}

// This function handles test memory store create deployment and get by id ok
// It supports test memory store create deployment and get by id ok behavior
func TestMemoryStore_CreateDeployment_And_GetByID_OK(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemoryStore()

	app, err := domain.NewApp(domain.NewAppParams{Name: "hello", Image: "nginx:latest", Port: ptrInt(8080)})
	assert.NoError(t, err)
	assert.NoError(t, st.CreateApp(ctx, app))

	dep := domain.NewDeployment(app.ID)
	assert.NoError(t, st.CreateDeployment(ctx, dep))

	got, err := st.GetDeploymentByID(ctx, dep.ID)
	assert.NoError(t, err)
	assert.Equal(t, dep.ID, got.ID)
	assert.Equal(t, app.ID, got.AppID)
	assert.Equal(t, domain.DeploymentStatusQueued, got.Status)
}

// This function handles test memory store list deployments by app id order and reflects updates
// It supports test memory store list deployments by app id order and reflects updates behavior
func TestMemoryStore_ListDeploymentsByAppID_OrderAndReflectsUpdates(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemoryStore()

	app, err := domain.NewApp(domain.NewAppParams{Name: "hello", Image: "nginx:latest", Port: ptrInt(8080)})
	require.NoError(t, err)
	require.NoError(t, st.CreateApp(ctx, app))

	d1 := domain.NewDeployment(app.ID)
	d2 := domain.NewDeployment(app.ID)

	require.NoError(t, st.CreateDeployment(ctx, d1))
	require.NoError(t, st.CreateDeployment(ctx, d2))

	updatedURL := "https://example.com"
	d1.Status = domain.DeploymentStatusRunning
	d1.URL = &updatedURL
	require.NoError(t, st.UpdateDeployment(ctx, d1))

	deps, err := st.ListDeploymentsByAppID(ctx, app.ID)
	require.NoError(t, err)
	require.Len(t, deps, 2)
	assert.Equal(t, d1.ID, deps[0].ID)
	assert.Equal(t, domain.DeploymentStatusRunning, deps[0].Status)
	require.NotNil(t, deps[0].URL)
	assert.Equal(t, updatedURL, *deps[0].URL)
	assert.Equal(t, d2.ID, deps[1].ID)
}

// This function handles test memory store list deployments by app id empty slice
// It supports test memory store list deployments by app id empty slice behavior
func TestMemoryStore_ListDeploymentsByAppID_EmptySlice(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemoryStore()

	deps, err := st.ListDeploymentsByAppID(ctx, "missing-app")
	assert.NoError(t, err)
	assert.Empty(t, deps)
}

// This function handles test memory store take next queued deployment fifo
// It supports test memory store take next queued deployment fifo behavior
func TestMemoryStore_TakeNextQueuedDeployment_FIFO(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemoryStore()

	app, err := domain.NewApp(domain.NewAppParams{Name: "hello", Image: "nginx:latest", Port: ptrInt(8080)})
	require.NoError(t, err)
	require.NoError(t, st.CreateApp(ctx, app))

	first := domain.NewDeployment(app.ID)
	second := domain.NewDeployment(app.ID)
	require.NoError(t, st.CreateDeployment(ctx, first))
	require.NoError(t, st.CreateDeployment(ctx, second))

	dep, err := st.TakeNextQueuedDeployment(ctx)
	require.NoError(t, err)
	assert.Equal(t, first.ID, dep.ID)

	dep, err = st.TakeNextQueuedDeployment(ctx)
	require.NoError(t, err)
	assert.Equal(t, second.ID, dep.ID)

	_, err = st.TakeNextQueuedDeployment(ctx)
	assert.ErrorIs(t, err, contracts.ErrNotFound)
}

// This function handles test memory store take next queued deployment skips non queued
// It supports test memory store take next queued deployment skips non queued behavior
func TestMemoryStore_TakeNextQueuedDeployment_SkipsNonQueued(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemoryStore()

	app, err := domain.NewApp(domain.NewAppParams{Name: "hello", Image: "nginx:latest", Port: ptrInt(8080)})
	require.NoError(t, err)
	require.NoError(t, st.CreateApp(ctx, app))

	skip := domain.NewDeployment(app.ID)
	next := domain.NewDeployment(app.ID)
	require.NoError(t, st.CreateDeployment(ctx, skip))
	require.NoError(t, st.CreateDeployment(ctx, next))

	skip.Status = domain.DeploymentStatusRunning
	require.NoError(t, st.UpdateDeployment(ctx, skip))

	dep, err := st.TakeNextQueuedDeployment(ctx)
	require.NoError(t, err)
	assert.Equal(t, next.ID, dep.ID)

	_, err = st.TakeNextQueuedDeployment(ctx)
	assert.ErrorIs(t, err, contracts.ErrNotFound)
}

// This function handles test memory store update deployment not found
// It supports test memory store update deployment not found behavior
func TestMemoryStore_UpdateDeployment_NotFound(t *testing.T) {
	ctx := context.Background()
	st := store.NewMemoryStore()

	dep := domain.NewDeployment("missing-app")

	err := st.UpdateDeployment(ctx, dep)
	assert.Error(t, err)
	assert.ErrorIs(t, err, contracts.ErrNotFound)
}

// This function handles ptr int
// It supports ptr int behavior
func ptrInt(v int) *int {
	return &v
}
