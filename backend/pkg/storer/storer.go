package storer

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"lifesupport/backend/pkg/api"
	"time"

	_ "github.com/lib/pq"
)

// Storer provides database operations for device data
type Storer struct {
	db *sql.DB
}

// New creates a new Storer instance with a PostgreSQL connection
func New(connString string) (*Storer, error) {
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Storer{db: db}, nil
}

// Close closes the database connection
func (s *Storer) Close() error {
	return s.db.Close()
}

// InitSchema creates the necessary database tables
func (s *Storer) InitSchema(ctx context.Context) error {
	schema := `
	CREATE TABLE IF NOT EXISTS systems (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		created_at TIMESTAMP NOT NULL,
		updated_at TIMESTAMP NOT NULL
	);

	CREATE TABLE IF NOT EXISTS subsystems (
		id VARCHAR(255) PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		type VARCHAR(50) NOT NULL,
		parent_id VARCHAR(255) REFERENCES subsystems(id) ON DELETE CASCADE,
		system_id VARCHAR(255) REFERENCES systems(id) ON DELETE CASCADE,
		metadata JSONB,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_subsystems_parent_id ON subsystems(parent_id);
	CREATE INDEX IF NOT EXISTS idx_subsystems_system_id ON subsystems(system_id);

	CREATE TABLE IF NOT EXISTS devices (
		id VARCHAR(255) PRIMARY KEY,
		driver VARCHAR(50) NOT NULL,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		subsystem_id VARCHAR(255) REFERENCES subsystems(id) ON DELETE CASCADE,
		metadata JSONB,
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_devices_subsystem_id ON devices(subsystem_id);

	CREATE TABLE IF NOT EXISTS sensor_readings (
		id SERIAL PRIMARY KEY,
		device_id VARCHAR(255) NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
		sensor_id VARCHAR(255) NOT NULL,
		sensor_name VARCHAR(255) NOT NULL,
		sensor_type VARCHAR(50) NOT NULL,
		value DOUBLE PRECISION NOT NULL,
		unit VARCHAR(20) NOT NULL,
		valid BOOLEAN NOT NULL DEFAULT true,
		error TEXT,
		timestamp TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_sensor_readings_device_id ON sensor_readings(device_id);
	CREATE INDEX IF NOT EXISTS idx_sensor_readings_sensor_id ON sensor_readings(sensor_id);
	CREATE INDEX IF NOT EXISTS idx_sensor_readings_timestamp ON sensor_readings(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_sensor_readings_type ON sensor_readings(sensor_type);

	CREATE TABLE IF NOT EXISTS actuator_states (
		id SERIAL PRIMARY KEY,
		device_id VARCHAR(255) NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
		actuator_id VARCHAR(255) NOT NULL,
		actuator_name VARCHAR(255) NOT NULL,
		actuator_type VARCHAR(50) NOT NULL,
		active BOOLEAN NOT NULL,
		parameters JSONB,
		error TEXT,
		timestamp TIMESTAMP NOT NULL,
		created_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_actuator_states_device_id ON actuator_states(device_id);
	CREATE INDEX IF NOT EXISTS idx_actuator_states_actuator_id ON actuator_states(actuator_id);
	CREATE INDEX IF NOT EXISTS idx_actuator_states_timestamp ON actuator_states(timestamp DESC);
	CREATE INDEX IF NOT EXISTS idx_actuator_states_type ON actuator_states(actuator_type);
	`

	_, err := s.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	return nil
}

// System operations

// CreateSystem creates a new system in the database
func (s *Storer) CreateSystem(ctx context.Context, sys *api.System) error {
	query := `
		INSERT INTO systems (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err := s.db.ExecContext(ctx, query, sys.ID, sys.Name, sys.Description, sys.CreatedAt, sys.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to create system: %w", err)
	}

	// Create subsystems recursively
	for _, subsystem := range sys.Subsystems {
		if err := s.createSubsystem(ctx, subsystem, sys.ID, nil); err != nil {
			return err
		}
	}

	return nil
}

// GetSystem retrieves a system by ID with all its subsystems and devices
func (s *Storer) GetSystem(ctx context.Context, id string) (*api.System, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM systems WHERE id = $1`

	var sys api.System
	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&sys.ID, &sys.Name, &sys.Description, &sys.CreatedAt, &sys.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("system not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get system: %w", err)
	}

	// Load subsystems
	subsystems, err := s.getSubsystemsBySystemID(ctx, sys.ID)
	if err != nil {
		return nil, err
	}
	sys.Subsystems = subsystems

	return &sys, nil
}

// UpdateSystem updates an existing system
func (s *Storer) UpdateSystem(ctx context.Context, sys *api.System) error {
	query := `
		UPDATE systems 
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
	`
	sys.UpdatedAt = time.Now()
	result, err := s.db.ExecContext(ctx, query, sys.ID, sys.Name, sys.Description, sys.UpdatedAt)
	if err != nil {
		return fmt.Errorf("failed to update system: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("system not found: %s", sys.ID)
	}

	return nil
}

// DeleteSystem deletes a system and all its children (cascading)
func (s *Storer) DeleteSystem(ctx context.Context, id string) error {
	query := `DELETE FROM systems WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete system: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("system not found: %s", id)
	}

	return nil
}

// ListSystems retrieves all systems with their subsystems and devices
func (s *Storer) ListSystems(ctx context.Context) ([]*api.System, error) {
	query := `SELECT id, name, description, created_at, updated_at FROM systems ORDER BY name`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query systems: %w", err)
	}
	defer rows.Close()

	var systems []*api.System
	for rows.Next() {
		var sys api.System
		err := rows.Scan(
			&sys.ID, &sys.Name, &sys.Description, &sys.CreatedAt, &sys.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan system: %w", err)
		}

		// Load subsystems
		subsystems, err := s.getSubsystemsBySystemID(ctx, sys.ID)
		if err != nil {
			return nil, err
		}
		sys.Subsystems = subsystems

		systems = append(systems, &sys)
	}

	return systems, rows.Err()
}

// Subsystem operations

func (s *Storer) createSubsystem(ctx context.Context, sub *api.Subsystem, systemID string, parentID *string) error {
	metadata, err := json.Marshal(sub.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO subsystems (id, name, description, type, parent_id, system_id, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, NOW(), NOW())
	`
	_, err = s.db.ExecContext(ctx, query, sub.ID, sub.Name, sub.Description, sub.Type, parentID, systemID, metadata)
	if err != nil {
		return fmt.Errorf("failed to create subsystem: %w", err)
	}

	// Create child subsystems
	for _, child := range sub.Subsystems {
		childParentID := sub.ID
		if err := s.createSubsystem(ctx, child, systemID, &childParentID); err != nil {
			return err
		}
	}

	// Create devices
	for _, dev := range sub.Devices {
		if err := s.createDevice(ctx, dev, sub.ID); err != nil {
			return err
		}
	}

	return nil
}

// CreateSubsystem creates a new subsystem
func (s *Storer) CreateSubsystem(ctx context.Context, sub *api.Subsystem, systemID string) error {
	return s.createSubsystem(ctx, sub, systemID, nil)
}

// GetSubsystem retrieves a subsystem by ID with all its devices and children
func (s *Storer) GetSubsystem(ctx context.Context, id string) (*api.Subsystem, error) {
	query := `
		SELECT id, name, description, type, parent_id, metadata
		FROM subsystems 
		WHERE id = $1
	`

	var sub api.Subsystem
	var parentID sql.NullString
	var metadataJSON []byte

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&sub.ID, &sub.Name, &sub.Description, &sub.Type, &parentID, &metadataJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("subsystem not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get subsystem: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &sub.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Load devices
	devices, err := s.getDevicesBySubsystemID(ctx, sub.ID)
	if err != nil {
		return nil, err
	}
	sub.Devices = devices

	// Load child subsystems
	children, err := s.getChildSubsystems(ctx, sub.ID)
	if err != nil {
		return nil, err
	}
	sub.Subsystems = children

	return &sub, nil
}

func (s *Storer) getSubsystemsBySystemID(ctx context.Context, systemID string) ([]*api.Subsystem, error) {
	query := `
		SELECT id, name, description, type, parent_id, metadata
		FROM subsystems 
		WHERE system_id = $1 AND parent_id IS NULL
	`

	rows, err := s.db.QueryContext(ctx, query, systemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query subsystems: %w", err)
	}
	defer rows.Close()

	var subsystems []*api.Subsystem
	for rows.Next() {
		var sub api.Subsystem
		var parentID sql.NullString
		var metadataJSON []byte

		err := rows.Scan(&sub.ID, &sub.Name, &sub.Description, &sub.Type, &parentID, &metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subsystem: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &sub.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		// Load devices
		devices, err := s.getDevicesBySubsystemID(ctx, sub.ID)
		if err != nil {
			return nil, err
		}
		sub.Devices = devices

		// Load child subsystems recursively
		children, err := s.getChildSubsystems(ctx, sub.ID)
		if err != nil {
			return nil, err
		}
		sub.Subsystems = children

		subsystems = append(subsystems, &sub)
	}

	return subsystems, rows.Err()
}

func (s *Storer) getChildSubsystems(ctx context.Context, parentID string) ([]*api.Subsystem, error) {
	query := `
		SELECT id, name, description, type, metadata
		FROM subsystems 
		WHERE parent_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, parentID)
	if err != nil {
		return nil, fmt.Errorf("failed to query child subsystems: %w", err)
	}
	defer rows.Close()

	var subsystems []*api.Subsystem
	for rows.Next() {
		var sub api.Subsystem
		var metadataJSON []byte

		err := rows.Scan(&sub.ID, &sub.Name, &sub.Description, &sub.Type, &metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan child subsystem: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &sub.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		// Load devices
		devices, err := s.getDevicesBySubsystemID(ctx, sub.ID)
		if err != nil {
			return nil, err
		}
		sub.Devices = devices

		// Load child subsystems recursively
		children, err := s.getChildSubsystems(ctx, sub.ID)
		if err != nil {
			return nil, err
		}
		sub.Subsystems = children

		subsystems = append(subsystems, &sub)
	}

	return subsystems, rows.Err()
}

// UpdateSubsystem updates an existing subsystem
func (s *Storer) UpdateSubsystem(ctx context.Context, sub *api.Subsystem) error {
	metadata, err := json.Marshal(sub.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE subsystems 
		SET name = $2, description = $3, type = $4, metadata = $5, updated_at = NOW()
		WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query, sub.ID, sub.Name, sub.Description, sub.Type, metadata)
	if err != nil {
		return fmt.Errorf("failed to update subsystem: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("subsystem not found: %s", sub.ID)
	}

	return nil
}

// DeleteSubsystem deletes a subsystem and all its children (cascading)
func (s *Storer) DeleteSubsystem(ctx context.Context, id string) error {
	query := `DELETE FROM subsystems WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete subsystem: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("subsystem not found: %s", id)
	}

	return nil
}

// ListSubsystems retrieves all subsystems across all systems with their devices and children
func (s *Storer) ListSubsystems(ctx context.Context) ([]*api.Subsystem, error) {
	query := `
		SELECT id, name, description, type, parent_id, metadata
		FROM subsystems 
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query subsystems: %w", err)
	}
	defer rows.Close()

	var subsystems []*api.Subsystem
	for rows.Next() {
		var sub api.Subsystem
		var parentID sql.NullString
		var metadataJSON []byte

		err := rows.Scan(&sub.ID, &sub.Name, &sub.Description, &sub.Type, &parentID, &metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan subsystem: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &sub.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		// Load devices
		devices, err := s.getDevicesBySubsystemID(ctx, sub.ID)
		if err != nil {
			return nil, err
		}
		sub.Devices = devices

		// Load child subsystems
		children, err := s.getChildSubsystems(ctx, sub.ID)
		if err != nil {
			return nil, err
		}
		sub.Subsystems = children

		subsystems = append(subsystems, &sub)
	}

	return subsystems, rows.Err()
}

// Device operations

func (s *Storer) createDevice(ctx context.Context, dev *api.Device, subsystemID string) error {
	metadata, err := json.Marshal(dev.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		INSERT INTO devices (id, driver, name, description, subsystem_id, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err = s.db.ExecContext(ctx, query, dev.ID, dev.Driver, dev.Name, dev.Description, subsystemID, metadata)
	if err != nil {
		return fmt.Errorf("failed to create device: %w", err)
	}

	return nil
}

// CreateDevice creates a new device
func (s *Storer) CreateDevice(ctx context.Context, dev *api.Device, subsystemID string) error {
	return s.createDevice(ctx, dev, subsystemID)
}

// GetDevice retrieves a device by ID
func (s *Storer) GetDevice(ctx context.Context, id string) (*api.Device, error) {
	query := `
		SELECT id, driver, name, description, metadata
		FROM devices 
		WHERE id = $1
	`

	var dev api.Device
	var metadataJSON []byte

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&dev.ID, &dev.Driver, &dev.Name, &dev.Description, &metadataJSON,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("device not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &dev.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	// Note: Sensors and Actuators are not stored in DB as they are interfaces
	// They would be reconstructed by the application layer

	return &dev, nil
}

func (s *Storer) getDevicesBySubsystemID(ctx context.Context, subsystemID string) ([]*api.Device, error) {
	query := `
		SELECT id, driver, name, description, metadata
		FROM devices 
		WHERE subsystem_id = $1
	`

	rows, err := s.db.QueryContext(ctx, query, subsystemID)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []*api.Device
	for rows.Next() {
		var dev api.Device
		var metadataJSON []byte

		err := rows.Scan(&dev.ID, &dev.Driver, &dev.Name, &dev.Description, &metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &dev.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		devices = append(devices, &dev)
	}

	return devices, rows.Err()
}

// UpdateDevice updates an existing device
func (s *Storer) UpdateDevice(ctx context.Context, dev *api.Device) error {
	metadata, err := json.Marshal(dev.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE devices 
		SET driver = $2, name = $3, description = $4, metadata = $5, updated_at = NOW()
		WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query, dev.ID, dev.Driver, dev.Name, dev.Description, metadata)
	if err != nil {
		return fmt.Errorf("failed to update device: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("device not found: %s", dev.ID)
	}

	return nil
}

// DeleteDevice deletes a device and all its sensor readings and actuator states (cascading)
func (s *Storer) DeleteDevice(ctx context.Context, id string) error {
	query := `DELETE FROM devices WHERE id = $1`
	result, err := s.db.ExecContext(ctx, query, id)
	if err != nil {
		return fmt.Errorf("failed to delete device: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("device not found: %s", id)
	}

	return nil
}

// ListDevices retrieves all devices across all subsystems
func (s *Storer) ListDevices(ctx context.Context) ([]*api.Device, error) {
	query := `
		SELECT id, driver, name, description, metadata
		FROM devices 
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query devices: %w", err)
	}
	defer rows.Close()

	var devices []*api.Device
	for rows.Next() {
		var dev api.Device
		var metadataJSON []byte

		err := rows.Scan(&dev.ID, &dev.Driver, &dev.Name, &dev.Description, &metadataJSON)
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &dev.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		devices = append(devices, &dev)
	}

	return devices, rows.Err()
}

// Sensor Reading operations

// StoreSensorReading stores a sensor reading in the database
func (s *Storer) StoreSensorReading(ctx context.Context, deviceID, sensorID, sensorName string, sensorType api.SensorType, reading *api.SensorReading) error {
	query := `
		INSERT INTO sensor_readings (device_id, sensor_id, sensor_name, sensor_type, value, unit, valid, error, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
	`
	_, err := s.db.ExecContext(ctx, query,
		deviceID, sensorID, sensorName, sensorType,
		reading.Value, reading.Unit, reading.Valid, reading.Error, reading.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to store sensor reading: %w", err)
	}

	return nil
}

// GetSensorReadings retrieves sensor readings with optional filters
func (s *Storer) GetSensorReadings(ctx context.Context, filters SensorReadingFilters) ([]*api.SensorReading, error) {
	query := `
		SELECT value, unit, valid, error, timestamp
		FROM sensor_readings
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if filters.DeviceID != nil {
		query += fmt.Sprintf(" AND device_id = $%d", argCount)
		args = append(args, *filters.DeviceID)
		argCount++
	}

	if filters.SensorID != nil {
		query += fmt.Sprintf(" AND sensor_id = $%d", argCount)
		args = append(args, *filters.SensorID)
		argCount++
	}

	if filters.SensorType != nil {
		query += fmt.Sprintf(" AND sensor_type = $%d", argCount)
		args = append(args, *filters.SensorType)
		argCount++
	}

	if filters.StartTime != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argCount)
		args = append(args, *filters.StartTime)
		argCount++
	}

	if filters.EndTime != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argCount)
		args = append(args, *filters.EndTime)
		argCount++
	}

	query += " ORDER BY timestamp DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query sensor readings: %w", err)
	}
	defer rows.Close()

	var readings []*api.SensorReading
	for rows.Next() {
		var reading api.SensorReading
		var errorMsg sql.NullString

		err := rows.Scan(&reading.Value, &reading.Unit, &reading.Valid, &errorMsg, &reading.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan sensor reading: %w", err)
		}

		if errorMsg.Valid {
			reading.Error = errorMsg.String
		}

		readings = append(readings, &reading)
	}

	return readings, rows.Err()
}

// GetLatestSensorReading retrieves the most recent sensor reading for a sensor
func (s *Storer) GetLatestSensorReading(ctx context.Context, sensorID string) (*api.SensorReading, error) {
	query := `
		SELECT value, unit, valid, error, timestamp
		FROM sensor_readings
		WHERE sensor_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var reading api.SensorReading
	var errorMsg sql.NullString

	err := s.db.QueryRowContext(ctx, query, sensorID).Scan(
		&reading.Value, &reading.Unit, &reading.Valid, &errorMsg, &reading.Timestamp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no readings found for sensor: %s", sensorID)
		}
		return nil, fmt.Errorf("failed to get latest sensor reading: %w", err)
	}

	if errorMsg.Valid {
		reading.Error = errorMsg.String
	}

	return &reading, nil
}

// DeleteOldSensorReadings deletes sensor readings older than the specified time
func (s *Storer) DeleteOldSensorReadings(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM sensor_readings WHERE timestamp < $1`
	result, err := s.db.ExecContext(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old sensor readings: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// Actuator State operations

// StoreActuatorState stores an actuator state in the database
func (s *Storer) StoreActuatorState(ctx context.Context, deviceID, actuatorID, actuatorName string, actuatorType api.ActuatorType, state *api.ActuatorState) error {
	parameters, err := json.Marshal(state.Parameters)
	if err != nil {
		return fmt.Errorf("failed to marshal parameters: %w", err)
	}

	query := `
		INSERT INTO actuator_states (device_id, actuator_id, actuator_name, actuator_type, active, parameters, error, timestamp)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err = s.db.ExecContext(ctx, query,
		deviceID, actuatorID, actuatorName, actuatorType,
		state.Active, parameters, state.Error, state.Timestamp,
	)
	if err != nil {
		return fmt.Errorf("failed to store actuator state: %w", err)
	}

	return nil
}

// GetActuatorStates retrieves actuator states with optional filters
func (s *Storer) GetActuatorStates(ctx context.Context, filters ActuatorStateFilters) ([]*api.ActuatorState, error) {
	query := `
		SELECT active, parameters, error, timestamp
		FROM actuator_states
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	if filters.DeviceID != nil {
		query += fmt.Sprintf(" AND device_id = $%d", argCount)
		args = append(args, *filters.DeviceID)
		argCount++
	}

	if filters.ActuatorID != nil {
		query += fmt.Sprintf(" AND actuator_id = $%d", argCount)
		args = append(args, *filters.ActuatorID)
		argCount++
	}

	if filters.ActuatorType != nil {
		query += fmt.Sprintf(" AND actuator_type = $%d", argCount)
		args = append(args, *filters.ActuatorType)
		argCount++
	}

	if filters.StartTime != nil {
		query += fmt.Sprintf(" AND timestamp >= $%d", argCount)
		args = append(args, *filters.StartTime)
		argCount++
	}

	if filters.EndTime != nil {
		query += fmt.Sprintf(" AND timestamp <= $%d", argCount)
		args = append(args, *filters.EndTime)
		argCount++
	}

	query += " ORDER BY timestamp DESC"

	if filters.Limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, filters.Limit)
	}

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query actuator states: %w", err)
	}
	defer rows.Close()

	var states []*api.ActuatorState
	for rows.Next() {
		var state api.ActuatorState
		var parametersJSON []byte
		var errorMsg sql.NullString

		err := rows.Scan(&state.Active, &parametersJSON, &errorMsg, &state.Timestamp)
		if err != nil {
			return nil, fmt.Errorf("failed to scan actuator state: %w", err)
		}

		if len(parametersJSON) > 0 {
			if err := json.Unmarshal(parametersJSON, &state.Parameters); err != nil {
				return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
			}
		}

		if errorMsg.Valid {
			state.Error = errorMsg.String
		}

		states = append(states, &state)
	}

	return states, rows.Err()
}

// GetLatestActuatorState retrieves the most recent state for an actuator
func (s *Storer) GetLatestActuatorState(ctx context.Context, actuatorID string) (*api.ActuatorState, error) {
	query := `
		SELECT active, parameters, error, timestamp
		FROM actuator_states
		WHERE actuator_id = $1
		ORDER BY timestamp DESC
		LIMIT 1
	`

	var state api.ActuatorState
	var parametersJSON []byte
	var errorMsg sql.NullString

	err := s.db.QueryRowContext(ctx, query, actuatorID).Scan(
		&state.Active, &parametersJSON, &errorMsg, &state.Timestamp,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("no states found for actuator: %s", actuatorID)
		}
		return nil, fmt.Errorf("failed to get latest actuator state: %w", err)
	}

	if len(parametersJSON) > 0 {
		if err := json.Unmarshal(parametersJSON, &state.Parameters); err != nil {
			return nil, fmt.Errorf("failed to unmarshal parameters: %w", err)
		}
	}

	if errorMsg.Valid {
		state.Error = errorMsg.String
	}

	return &state, nil
}

// DeleteOldActuatorStates deletes actuator states older than the specified time
func (s *Storer) DeleteOldActuatorStates(ctx context.Context, before time.Time) (int64, error) {
	query := `DELETE FROM actuator_states WHERE timestamp < $1`
	result, err := s.db.ExecContext(ctx, query, before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old actuator states: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rows, nil
}

// Filter types

// SensorReadingFilters defines optional filters for querying sensor readings
type SensorReadingFilters struct {
	DeviceID   *string
	SensorID   *string
	SensorType *api.SensorType
	StartTime  *time.Time
	EndTime    *time.Time
	Limit      int
}

// ActuatorStateFilters defines optional filters for querying actuator states
type ActuatorStateFilters struct {
	DeviceID     *string
	ActuatorID   *string
	ActuatorType *api.ActuatorType
	StartTime    *time.Time
	EndTime      *time.Time
	Limit        int
}
