// Traefik label helpers for docker runtime
// Labels build routing rules for an app
// Entry points and host rules are included
// Service labels point to the container port
// Tls labels are added when enabled

package docker

import (
	"fmt"

	"github.com/t0gun/spacescale/internal/domain"
)

// EdgeConfig holds settings for Traefik reverse proxy labels
type EdgeConfig struct {
	BaseDomain   string // root domain e g spacescale ai apps become myapp spacescale ai
	TraefikNet   string // docker network where traefik can reach containers
	Scheme       string // traefik entrypoint string web http or websecure https
	EnableTLS    bool   // enable https termination
	CertResolver string // certificate resolver name e g lets encrypt
}

// This function handles labels for app
// It supports labels for app behavior
func labelsForApp(app domain.App, port int, cfg EdgeConfig) map[string]string {
	// build the full hostname myapp spacescale ai myapp spacescale ai
	host := fmt.Sprintf("%s.%s", app.Name, cfg.BaseDomain)

	// unique identifiers for traefik router and service
	router := "app-" + app.Name
	svc := "svc-" + app.Name

	labels := map[string]string{
		// enable traefik for this container
		"traefik.enable": "true",

		// which docker network traefik should use to reach this container
		"traefik.docker.network": cfg.TraefikNet,

		// ROUTER match requests where Host header myapp spacescale ai
		fmt.Sprintf("traefik.http.routers.%s.rule", router): fmt.Sprintf("Host(`%s`)", host),

		// ROUTER which entrypoint listener accepts traffic for this app
		fmt.Sprintf("traefik.http.routers.%s.entrypoints", router): cfg.Scheme,

		// SERVICE forward traffic to this port inside the container
		fmt.Sprintf("traefik.http.services.%s.loadbalancer.server.port", svc): fmt.Sprintf("%d", port),
	}

	// TLS config optional
	if cfg.EnableTLS {
		// enable https for this router
		labels[fmt.Sprintf("traefik.http.routers.%s.tls", router)] = "true"

		// use cert resolver e g letsencrypt for auto ssl certificates
		if cfg.CertResolver != "" {
			labels[fmt.Sprintf("traefik.http.routers.%s.tls.certresolver", router)] = cfg.CertResolver
		}
	}

	return labels
}
