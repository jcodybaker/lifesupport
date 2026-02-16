package drivers

import (
	"context"
	"lifesupport/backend/pkg/storer"
)

type DiscoveryOptions struct {
	// Add any options needed for device discovery, e.g. timeouts, concurrency limits, etc.
}

type DiscoveryResult struct {
	DiscoveredTags []string
}

type Driver interface {
	DiscoverDevices(ctx context.Context, opt DiscoveryOptions, s *storer.Storer) (*DiscoveryResult, error)
}
