package shelly

import (
	"time"

	"github.com/rs/zerolog"
)

type Option func(*Driver)

func WithBaseName(baseName string) Option {
	return func(rt *Driver) {
		rt.baseName = baseName
	}
}

func WithClientName(name string) Option {
	return func(rt *Driver) {
		rt.clientName = name
	}
}

func WithDiscoveryBufferSize(size int) Option {
	return func(d *Driver) {
		d.discoveryBufferSize = size
	}
}

func WithDiscoveryTimeout(timeout time.Duration) Option {
	return func(d *Driver) {
		d.discoveryTimeout = timeout
	}
}

func WithDiscoveryWorkers(workers int) Option {
	return func(d *Driver) {
		d.discoveryWorkers = workers
	}
}

func WithLogger(logger zerolog.Logger) Option {
	return func(d *Driver) {
		d.log = logger
	}
}
