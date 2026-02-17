package drivers

import (
	"context"
	"errors"
	"lifesupport/backend/pkg/api"
	"lifesupport/backend/pkg/storer"
)

var ErrNoData = errors.New("no data available")

type Statuser interface {
	GetID() string
	GetDeviceID() string
}

type Driver interface {
	DiscoverDevices(ctx context.Context, opt api.DiscoveryOptions, s *storer.Storer) (*api.DiscoveryResult, error)
	GetLastStatus(ctx context.Context, opt api.StatusOptions, resource Statuser) (*api.SensorReading, error)
}

