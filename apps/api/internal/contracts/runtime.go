// Runtime interface for deployment implementations
// It defines how apps are deployed
// A nil url means the app runs without exposure
// Service code depends on this contract
// Runtime adapters implement this interface

package contracts

import (
	"context"

	"github.com/t0gun/spacescale/internal/domain"
)

// Runtime defines how an app is deployed and how its URL is returned
type Runtime interface {
	// This function handles deploy
	// It supports deploy behavior
	Deploy(ctx context.Context, app domain.App) (url *string, err error)
}
