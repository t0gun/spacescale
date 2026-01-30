// Tests for domain app creation and defaults
// Tests cover optional port and expose behavior
// Tests confirm status values and timestamps
// Tests verify deployment creation data
// These tests guard domain rules

package domain_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestNewAppDefault verifies app defaults and validation.
func TestNewAppDefault(t *testing.T) {
	tests := []struct {
		label      string
		in         NewAppParams
		ok         bool
		wantExpose bool
	}{
		{label: "valid app", in: NewAppParams{Name: "hello", Image: "nginx:latest", Port: ptrInt(8080)}, ok: true, wantExpose: true},
		{label: "valid app no port", in: NewAppParams{Name: "hello", Image: "nginx:latest"}, ok: true, wantExpose: true},
		{label: "valid app expose false", in: NewAppParams{Name: "hello", Image: "nginx:latest", Port: ptrInt(8080), Expose: ptrBool(false)}, ok: true, wantExpose: false},
		{label: "invalid name", in: NewAppParams{Name: "Bad_Name", Image: "nginx:latest", Port: ptrInt(8080)}, ok: false},
		{label: "empty image", in: NewAppParams{Name: "hello", Image: "", Port: ptrInt(8080)}, ok: false},
		{label: "invalid port", in: NewAppParams{Name: "hello", Image: "nginx:latest", Port: ptrInt(0)}, ok: false},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			app, err := NewApp(tt.in)

			if tt.ok {
				assert.NoError(t, err)
				assert.NotEmpty(t, app.ID)
				assert.Equal(t, tt.in.Name, app.Name)
				if tt.in.Port == nil {
					assert.Nil(t, app.Port)
				} else {
					assert.NotNil(t, app.Port)
					assert.Equal(t, *tt.in.Port, *app.Port)
				}
				assert.Equal(t, tt.in.Image, app.Image)
				assert.Equal(t, AppStatusCreated, app.Status)
				assert.Equal(t, tt.wantExpose, app.Expose)
			}

			if !tt.ok {
				assert.Error(t, err)
				assert.Empty(t, app.ID)
			}

		})
	}
}

// ptrBool returns a pointer to v.
func ptrBool(v bool) *bool {
	return &v
}

// ptrInt returns a pointer to v.
func ptrInt(v int) *int {
	return &v
}

// TestNewDeployment verifies deployment defaults.
func TestNewDeployment(t *testing.T) {
	tests := []struct {
		label string
		appID string
	}{
		{"normal", "app-123"},
		{"uuid-like", "8e6c9a65-6f01-4db3-a3c2-9c9c9c9c9c9c"},
	}

	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			dep := NewDeployment(tt.appID)

			assert.NotEmpty(t, dep.ID)
			assert.Equal(t, tt.appID, dep.AppID)
			assert.Equal(t, DeploymentStatusQueued, dep.Status)

			assert.False(t, dep.CreatedAt.IsZero())
			assert.False(t, dep.UpdatedAt.IsZero())
			assert.WithinDuration(t, time.Now().UTC(), dep.CreatedAt, 2*time.Second)
		})
	}
}
