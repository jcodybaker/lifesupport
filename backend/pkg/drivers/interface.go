package drivers

import (
	"context"
	"lifesupport/backend/pkg/api"
	"lifesupport/backend/pkg/storer"
)

type Driver interface {
	DiscoverDevices(ctx context.Context, opt api.DiscoveryOptions, s *storer.Storer) (*api.DiscoveryResult, error)
}
