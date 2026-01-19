package docker

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/client"
	"github.com/t0gun/paas/internal/domain"
)

// Runtime deploys apps using the local Docker engine.
type Runtime struct {
	cli           *client.Client
	advertiseHost string
	namePrefix    string
	timeout       time.Duration
}

// Option customizes runtime settings during construction.
type Option func(*Runtime)

// WithAdvertiseHost sets the hostname or IP that will be used in the returned URL.
// This is useful when the Docker host is not the same as the request origin.
// If not set, a local address is used by default.
func WithAdvertiseHost(host string) Option { return func(r *Runtime) { r.advertiseHost = host } }

// WithNamePrefix sets the container name prefix used for created containers.
// This makes it easy to find containers that belong to this runtime.
// If not set, a simple default prefix is used.
func WithNamePrefix(prefix string) Option  { return func(r *Runtime) { r.namePrefix = prefix } }

// WithTimeout sets a maximum duration for runtime operations like pull and start.
// The timeout applies to the full deploy sequence for a single app.
// If not set, a reasonable default is used.
func WithTimeout(d time.Duration) Option   { return func(r *Runtime) { r.timeout = d } }

// New creates a Docker runtime client and applies any optional configuration.
// The runtime starts with safe defaults and each option can override one part.
// This keeps the call site clean while still allowing customization.
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
	}
	for _, opt := range opts {
		opt(r)
	}
	return r, nil
}

// Deploy pulls the image, creates a container, starts it, and returns a URL.
// It validates the app input and uses a timeout to avoid hanging operations.
// The returned URL is built using the advertised host and the mapped host port.
func (r *Runtime) Deploy(ctx context.Context, app domain.App) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, r.timeout)
	defer cancel()

	if strings.TrimSpace(app.Image) == "" {
		return "", fmt.Errorf("docker runtime: empty image")
	}
	if app.Port < 1 || app.Port > 65535 {
		return "", fmt.Errorf("docker runtime: invalid port %d", app.Port)
	}
	// pull docker image
	if err := r.pull(ctx, app.Image); err != nil {
		return "", fmt.Errorf("docker runtime: pull: %w", err)
	}

	// replace existing container
	name := r.namePrefix + app.Name
	_ = r.removeIfExists(ctx, name)

	// create container with random host port binding
	cPort, err := network.ParsePort(fmt.Sprintf("%d/tcp", app.Port))
	if err != nil {
		return "", fmt.Errorf("docker runtime: parse port: %w", err)
	}
	exposed := network.PortSet{cPort: struct{}{}}

	cfg := &container.Config{
		Image:        app.Image,
		ExposedPorts: exposed,
		Labels: map[string]string{
			"paas.app": app.Name,
		},
	}

	hostcfg := &container.HostConfig{
		PublishAllPorts: true,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	created, err := r.cli.ContainerCreate(ctx, client.ContainerCreateOptions{
		Config:     cfg,
		HostConfig: hostcfg,
		Name:       name,
	})
	if err != nil {
		return "", fmt.Errorf("docker runtime: create: %w", err)
	}
	if _, err := r.cli.ContainerStart(ctx, created.ID, client.ContainerStartOptions{}); err != nil {
		return "", fmt.Errorf("docker runtime: start container: %w", err)
	}

	hostPort, err := r.hostPort(ctx, created.ID, cPort)
	if err != nil {
		return "", fmt.Errorf("docker runtime: resolve host port: %w", err)
	}

	return fmt.Sprintf("http://%s:%s", r.advertiseHost, hostPort), nil
}

// pull downloads the image if it is not present on the host.
// The response stream must be fully read so the pull can finish cleanly.
// Errors are returned directly to the caller for context handling.
func (r *Runtime) pull(ctx context.Context, ref string) error {
	rc, err := r.cli.ImagePull(ctx, ref, client.ImagePullOptions{})
	if err != nil {
		return err
	}
	defer rc.Close()
	_, _ = io.Copy(io.Discard, rc)
	return nil
}

// removeIfExists deletes an existing container by name so deploys can be repeated.
// Not found errors are treated as a no op so the deploy can continue.
// Any other error is returned to the caller.
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

// hostPort inspects the container to find which host port was mapped.
// Docker assigns a random port when PublishAllPorts is enabled.
// The result is used to build the external URL for the deployment.
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
