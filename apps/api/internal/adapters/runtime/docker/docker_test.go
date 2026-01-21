package docker_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/t0gun/paas/internal/adapters/runtime/docker"
	"github.com/t0gun/paas/internal/domain"
)

func TestDockerRuntime_Deploy(t *testing.T) {
	if os.Getenv("RUN_DOCKER_TESTS") != "1" {
		t.Skip("set RUN_DOCKER_TESTS=1 to run docker integration tests")
	}

	rt, err := docker.New(
		docker.WithEdge(docker.EdgeConfig{
			BaseDomain: "localtest.me",
			TraefikNet: "traefik",
			Scheme:     "web", // entrypoint name
			EnableTLS:  false,
		}),
	)
	assert.NoError(t, err)

	app, err := domain.NewApp(domain.NewAppParams{
		Name:  "hello",
		Image: "nginx:latest",
		Port:  ptrInt(80),
	})
	assert.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	url, err := rt.Deploy(ctx, app)
	assert.NoError(t, err)

	assert.NotNil(t, url)
	assert.Equal(t, "http://hello.localtest.me", *url)
}

func ptrInt(v int) *int {
	return &v
}
