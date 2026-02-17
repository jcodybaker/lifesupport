package drivers

import "lifesupport/backend/pkg/api"

type Manager struct {
	drivers map[api.DriverName]Driver
}

func NewManager() *Manager {
	return &Manager{
		drivers: make(map[api.DriverName]Driver),
	}
}

func (m *Manager) Register(name api.DriverName, driver Driver) {
	m.drivers[name] = driver
}

func (m *Manager) Get(name api.DriverName) (Driver, bool) {
	driver, exists := m.drivers[name]
	return driver, exists
}
