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

// LabelsForApp builds a set of Traefik v2 Docker labels for a single application container.
//
// Goal
//   - Given an app name, a container port, and edge (Traefik) configuration,
//     return a map of labels that Traefik can read from Docker and turn into:
//     1) a Router (request matching + entrypoints)
//     2) a Service (where to forward the traffic)
//
// What this function configures
//   - Host-based routing: requests for <app>.<base-domain> are routed to the container.
//   - EntryPoints: choose which Traefik listener(s) accept those requests.
//   - Service port: which internal container port Traefik forwards to.
//   - Optional TLS: enable TLS for the router and optionally select a cert resolver.
//
// Inputs
//   - app: the domain app model. We use app.Name to construct names and the hostname.
//   - port: the INTERNAL port exposed by the app process inside the container.
//     This is not the published host port.
//   - cfg: EdgeConfig settings that control routing and TLS behavior.
//
// Naming conventions used here
//   - host   = "<app.Name>.<cfg.BaseDomain>"  (example: "myapp.spacescale.ai")
//   - router = "app-<app.Name>"              (example: "app-myapp")
//   - svc    = "svc-<app.Name>"              (example: "svc-myapp")
//
// Traefik label mental model (very short)
//   - Router labels: traefik.http.routers.<name>.*
//     Define how to match requests (rule) and where they enter Traefik (entrypoints).
//   - Service labels: traefik.http.services.<name>.*
//     Define the upstream target(s) for matched requests (port, load balancer, etc.).
//
// Network note
//   - traefik.docker.network is important when containers are attached to multiple
//     networks. It tells Traefik which network to use to reach the container IP.
//
// TLS note
//   - When cfg.EnableTLS is true, we set router TLS to true.
//   - If cfg.CertResolver is set, Traefik will use that resolver (must be configured
//     in Traefik's static config) to obtain/renew certificates.
func LabelsForApp(app domain.App, port int, cfg EdgeConfig) map[string]string {
	// build the full hostname myapp spacescale ai myapp spacescale ai
	host := fmt.Sprintf("%s.%s", app.Name, cfg.BaseDomain)

	// unique identifiers for traefik router and service
	router := "app-" + app.Name
	svc := "svc-" + app.Name

	labels := map[string]string{
		// NOTE: All keys here are Traefik v2 label syntax.
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
