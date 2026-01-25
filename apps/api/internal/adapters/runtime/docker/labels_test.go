package docker_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/t0gun/spacescale/internal/adapters/runtime/docker"
	"github.com/t0gun/spacescale/internal/domain"
)

func TestLabelsForApp_NOTLS(t *testing.T) {
	app := domain.App{Name: "hello"}
	cfg := docker.EdgeConfig{
		BaseDomain: "example.com",
		TraefikNet: "traefik",
		Scheme:     "web",
		EnableTLS:  false,
	}
	labels := docker.LabelsForApp(app, 8080, cfg)
	assert.Equal(t, "true", labels["traefik.enable"])
	assert.Equal(t, "Host(`hello.example.com`)", labels["traefik.http.routers.app-hello.rule"])
	assert.Equal(t, "web", labels["traefik.http.routers.app-hello.entrypoints"])
	assert.Equal(t, "8080", labels["traefik.http.services.svc-hello.loadbalancer.server.port"])
	_, hasTLS := labels["traefik.http.routers.app-hello.tls"]
	assert.False(t, hasTLS)
}

func TestLabelsForApp_TLS(t *testing.T) {
	app := domain.App{Name: "hello"}
	cfg := docker.EdgeConfig{
		BaseDomain:   "example.com",
		TraefikNet:   "traefik",
		Scheme:       "web",
		EnableTLS:    true,
		CertResolver: "lets",
	}
	labels := docker.LabelsForApp(app, 443, cfg)
	assert.Equal(t, "true", labels["traefik.enable"])
	assert.Equal(t, "Host(`hello.example.com`)", labels["traefik.http.routers.app-hello.rule"])
	assert.Equal(t, "web", labels["traefik.http.routers.app-hello.entrypoints"])
	assert.Equal(t, "443", labels["traefik.http.services.svc-hello.loadbalancer.server.port"])
	assert.Equal(t, "true", labels["traefik.http.routers.app-hello.tls"])
	assert.Equal(t, "lets", labels["traefik.http.routers.app-hello.tls.certresolver"])
}
