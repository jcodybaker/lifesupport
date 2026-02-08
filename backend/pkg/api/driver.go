package api

type Driver interface {
	DiscoverDevices() ([]*Device, error)
}
