// Docker runtime implementation for deploying apps
// This file pulls images and creates containers
// Ports are resolved from app input or image metadata
// Labels are applied when exposure is enabled
// Env values are passed into container config
// A url is returned when the app is exposed

package docker

import (
	"context"
	"fmt"
	"io"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/t0gun/paas/internal/domain"
)

// Runtime deploys apps using the local Docker engine
type Runtime struct {
	cli           *client.Client
	advertiseHost string
	namePrefix    string
	timeout       time.Duration

	// edge routing
	edge EdgeConfig
}

const errPortRequiredMsg = "port required or image must expose exactly one port"

// Option customizes runtime settings during construction
type Option func(*Runtime)

// WithAdvertiseHost This function handles with advertise host
// It supports with advertise host behavior
func WithAdvertiseHost(host string) Option { return func(r *Runtime) { r.advertiseHost = host } }

// WithNamePrefix This function handles with name prefix
// It supports with name prefix behavior
func WithNamePrefix(prefix string) Option { return func(r *Runtime) { r.namePrefix = prefix } }

// WithTimeout This function handles with timeout
// It supports with timeout behavior
func WithTimeout(d time.Duration) Option { return func(r *Runtime) { r.timeout = d } }

// WithEdge This function handles with edge
// It supports with edge behavior
func WithEdge(cfg EdgeConfig) Option { return func(r *Runtime) { r.edge = cfg } }

// New This function handles new
// It supports new behavior
func New(opts ...Option) (*Runtime, error) {
	cli, err := client.New(
		client.FromEnv,
		client.WithAPIVersionFromEnv(),
	)
	if err != nil {
		return nil, err
	}
	// default runt time before mutation
	r := &Runtime{
		cli:           cli,
		advertiseHost: "127.0.0.1",
		namePrefix:    "paas-",
		timeout:       2 * time.Minute,
		edge: EdgeConfig{
			BaseDomain: "localtest.me",
			TraefikNet: "traefik",
			Scheme:     "web",
			EnableTLS:  false,
		},
	}
	for _, opt := range opts {
		opt(r)
	}
	return r, nil
}

// This function handles deploy
// It supports deploy behavior
func (r *Runtime) Deploy(ctx context.Context, app domain.App) (*string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()
	// validate
	if strings.TrimSpace(app.Image) == "" {
		return nil, fmt.Errorf("docker runtime: empty image")
	}
	if app.Expose {
		if strings.TrimSpace(r.edge.BaseDomain) == "" {
			return nil, fmt.Errorf("docker runtime: empty base domain")
		}
		if strings.TrimSpace(r.edge.TraefikNet) == "" {
			return nil, fmt.Errorf("docker runtime: empty traefik network")
		}
		if strings.TrimSpace(r.edge.Scheme) == "" {
			// this is actually traefik entrypoint name web websecure
			r.edge.Scheme = "web"
		}
	}

	// pull docker image
	if err := r.pull(ctx, app.Image); err != nil {
		return nil, fmt.Errorf("docker runtime: pull: %w", err)
	}

	port, err := r.resolvePort(ctx, app)
	if err != nil {
		return nil, err
	}

	// replace existing container
	name := r.namePrefix + app.Name
	_ = r.removeIfExists(ctx, name)

	// traefik labels
	lbls := map[string]string{
		"paas.app": app.Name,
	}
	if app.Expose {
		if port == nil {
			return nil, fmt.Errorf(errPortRequiredMsg)
		}
		for k, v := range labelsForApp(app, *port, r.edge) {
			lbls[k] = v
		}
	}

	cfg := &container.Config{
		Image:  app.Image,
		Labels: lbls,
		Env:    envToList(app.Env),
	}
	if port != nil {
		// Expose port internally for docs metadata traefik routes to container port
		cPort, err := network.ParsePort(fmt.Sprintf("%d/tcp", *port))
		if err != nil {
			return nil, fmt.Errorf("docker runtime: parse port: %w", err)
		}
		cfg.ExposedPorts = network.PortSet{cPort: struct{}{}}
	}

	hostcfg := &container.HostConfig{
		PublishAllPorts: false,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}
	if app.Expose {
		// put container on the traefik network
		hostcfg.NetworkMode = container.NetworkMode(r.edge.TraefikNet)
	}

	created, err := r.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     cfg,
		HostConfig: hostcfg,
		Name:       name,
	})
	if err != nil {
		return nil, fmt.Errorf("docker runtime: create: %w", err)
	}
	if _, err := r.cli.ContainerStart(ctx, created.ID, client.ContainerStartOptions{}); err != nil {
		return nil, fmt.Errorf("docker runtime: start container: %w", err)
	}

	// return stable URL
	if !app.Expose {
		return nil, nil
	}
	scheme := "http"
	if r.edge.EnableTLS {
		scheme = "https"
	}
	url := fmt.Sprintf("%s://%s.%s", scheme, app.Name, r.edge.BaseDomain)
	return &url, nil
}

// This function handles pull
// It supports pull behavior
func (r *Runtime) pull(ctx context.Context, ref string) error {
	rc, err := r.cli.ImagePull(ctx, ref, client.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer rc.Close()
	_, _ = io.Copy(io.Discard, rc)
	return nil
}

// This function handles remove if exists
// It supports remove if exists behavior
func (r *Runtime) removeIfExists(ctx context.Context, name string) error {
	_, err := r.cli.ContainerRemove(ctx, name, client.ContainerRemoveOptions{Force: true})
	if err == nil {
		return nil
	}
	msg := strings.ToLower(err.Error())
	if strings.Contains(msg, "no such container") || strings.Contains(msg, "not found") {
		return nil
	}
	return err
}

// This function handles resolve port
// It supports resolve port behavior
func (r *Runtime) resolvePort(ctx context.Context, app domain.App) (*int, error) {
	if app.Port != nil {
		if *app.Port < 1 || *app.Port > 65535 {
			return nil, fmt.Errorf("docker runtime: invalid port %d", *app.Port)
		}
		return app.Port, nil
	}
	if !app.Expose {
		return nil, nil
	}
	port, err := r.portFromImage(ctx, app.Image)
	if err != nil {
		return nil, err
	}
	return &port, nil
}

// This function handles port from image
// It supports port from image behavior
func (r *Runtime) portFromImage(ctx context.Context, ref string) (int, error) {
	inspect, err := r.cli.ImageInspect(ctx, ref)
	if err != nil {
		return 0, fmt.Errorf("docker runtime: inspect image: %w", err)
	}
	exposed := exposedPortsFromInspect(inspect.InspectResponse)
	if len(exposed) != 1 {
		return 0, fmt.Errorf(errPortRequiredMsg)
	}
	port, err := parseExposedPort(exposed[0])
	if err != nil {
		return 0, fmt.Errorf("docker runtime: parse exposed port: %w", err)
	}
	if port < 1 || port > 65535 {
		return 0, fmt.Errorf("docker runtime: invalid port %d", port)
	}
	return port, nil
}

// This function handles exposed ports from inspect
// It supports exposed ports from inspect behavior
func exposedPortsFromInspect(inspect image.InspectResponse) []string {
	if inspect.Config == nil || len(inspect.Config.ExposedPorts) == 0 {
		return nil
	}
	ports := make([]string, 0, len(inspect.Config.ExposedPorts))
	for port := range inspect.Config.ExposedPorts {
		ports = append(ports, port)
	}
	sort.Strings(ports)
	return ports
}

// This function handles parse exposed port
// It supports parse exposed port behavior
func parseExposedPort(spec string) (int, error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return 0, fmt.Errorf("empty exposed port")
	}
	if idx := strings.Index(spec, "/"); idx != -1 {
		spec = spec[:idx]
	}
	return strconv.Atoi(spec)
}

// This function handles env to list
// It supports env to list behavior
func envToList(env map[string]string) []string {
	if len(env) == 0 {
		return nil
	}
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	out := make([]string, 0, len(env))
	for _, k := range keys {
		out = append(out, fmt.Sprintf("%s=%s", k, env[k]))
	}
	return out
}

// This function handles host port
// It supports host port behavior
func (r *Runtime) hostPort(ctx context.Context, containerID string, cPort network.Port) (string, error) {
	ins, err := r.cli.ContainerInspect(ctx, containerID, client.ContainerInspectOptions{})
	if err != nil {
		return "", err
	}
	if ins.Container.NetworkSettings == nil {
		return "", fmt.Errorf("missing network settings")
	}

	b := ins.Container.NetworkSettings.Ports[cPort]
	if len(b) == 0 || strings.TrimSpace(b[0].HostPort) == "" {
		return "", fmt.Errorf("no host port mapped for %s", cPort)
	}
	return b[0].HostPort, nil
}
