package contracts

import (
	"context"

	"github.com/t0gun/paas/internal/domain"
)

// Runtime defines how an app is deployed and how its URL is returned.
type Runtime interface {
	// Deploy runs the app and returns the reachable URL when exposed.
	// A nil URL indicates the app is running without exposure.
	Deploy(ctx context.Context, app domain.App) (url *string, err error)
}
