// Fake runtime for tests and local use
// It builds a predictable url for an app
// It respects exposure and can return nil
// It avoids docker dependencies in tests
// It keeps runtime behavior simple and stable

package fake

import (
	"context"
	"fmt"
	"strings"

	"github.com/t0gun/spacescale/internal/domain"
)

// Runtime is a fake runtime that returns a predictable URL
type Runtime struct {
	BaseDomain string
	Scheme     string // https or http
}

// This function handles new
// It supports new behavior
func New(baseDomain string) *Runtime {
	return &Runtime{
		BaseDomain: strings.TrimSpace(baseDomain),
		Scheme:     "https",
	}
}

// This function handles deploy
// It supports deploy behavior
func (r *Runtime) Deploy(ctx context.Context, app domain.App) (*string, error) {
	if !app.Expose {
		return nil, nil
	}
	url := fmt.Sprintf("%s://%s.%s", r.Scheme, app.Name, r.BaseDomain)
	return &url, nil
}
