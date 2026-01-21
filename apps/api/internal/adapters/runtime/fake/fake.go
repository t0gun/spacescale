package fake

import (
	"context"
	"fmt"
	"strings"

	"github.com/t0gun/paas/internal/domain"
)

// Runtime is a fake runtime that returns a predictable URL.
type Runtime struct {
	BaseDomain string
	Scheme     string // https or http
}

// New constructs a fake runtime with a base domain.
func New(baseDomain string) *Runtime {
	return &Runtime{
		BaseDomain: strings.TrimSpace(baseDomain),
		Scheme:     "https",
	}
}

// Deploy returns a synthetic URL for the app without doing real work.
func (r *Runtime) Deploy(ctx context.Context, app domain.App) (*string, error) {
	if !app.Expose {
		return nil, nil
	}
	url := fmt.Sprintf("%s://%s.%s", r.Scheme, app.Name, r.BaseDomain)
	return &url, nil
}
