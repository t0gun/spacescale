package docker

import (
	"fmt"

	"github.com/t0gun/paas/internal/domain"
)

// EdgeConfig captures settings for reverse proxy labels.
type EdgeConfig struct {
	BaseDomain   string // spacescale.ai
	TraefikNet   string // traefik
	Scheme       string
	EnableTLS    bool
	CertResolver string
}

// labelsForApp builds reverse proxy labels for a single app.
func labelsForApp(app domain.App, cfg EdgeConfig) map[string]string {
	host := fmt.Sprintf("%s.%s", app.Name, cfg.BaseDomain)
	router := "app-" + app.Name
	svc := "svc-" + app.Name

	labels := map[string]string{
		"traefik.enable": "true",
		// tell traefik which docker network to use to reach the container
		"traefik.docker.network": cfg.TraefikNet,

		// Router
		fmt.Sprintf("traefik.http.routers.%s.rule", router):        fmt.Sprintf("Host(`%s`)", host),
		fmt.Sprintf("traefik.http.routers.%s.entrypoints", router): cfg.Scheme,

		// Service
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", svc): fmt.Sprintf("%d", app.Port),
	}

	// Add TLS configuration if enabled
	if cfg.EnableTLS {
		labels[fmt.Sprintf("traefik.http.routers.%s.tls", router)] = "true"
		if cfg.CertResolver != "" {
			labels[fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", router)] = cfg.CertResolver
		}
	}

	return labels
}
