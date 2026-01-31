package device

type Driver interface {
	DiscoverDevices() ([]*Device, error)
}
