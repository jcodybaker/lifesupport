package storer

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"lifesupport/backend/pkg/api"

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Exported errors
var (
	ErrNotFound      = errors.New("not found")
	ErrAlreadyExists = errors.New("already exists")
)

// Storer provides database operations for device data
type Storer struct {
	db  *sql.DB
	log zerolog.Logger
}

// New creates a new Storer instance with a PostgreSQL connection
func New(connString string, opts ...Option) (*Storer, error) {
	s := &Storer{}
	for _, opt := range opts {
		opt(s)
	}
	db, err := sql.Open("postgres", connString)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}
	s.db = db

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return s, nil
}

func (s *Storer) logCtx(ctx context.Context, sub string) zerolog.Logger {
	var ll zerolog.Context
	if ctxLog := log.Ctx(ctx); ctxLog.GetLevel() != zerolog.Disabled {
		ll = ctxLog.With()
	} else {
		ll = s.log.With()
	}
	ll = ll.Str("component", "storer")
	if sub != "" {
		ll = ll.Str("subcomponent", sub)
	}
	return ll.Logger()
}

// Close closes the database connection
func (s *Storer) Close() error {
	log.Debug().Msg("closing database connection")
	return s.db.Close()
}

// InitSchema creates the necessary database tables
func (s *Storer) InitSchema(ctx context.Context) error {
	ll := s.logCtx(ctx, "schema")
	ll.Debug().Msg("initializing database schema")
	schema := `
	CREATE TABLE IF NOT EXISTS devices (
		id VARCHAR(255) PRIMARY KEY,
		driver VARCHAR(50) NOT NULL,
		name VARCHAR(255) NOT NULL,
		description TEXT,
		metadata JSONB,
		tags TEXT[],
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW()
	);

	CREATE INDEX IF NOT EXISTS idx_devices_tags ON devices USING GIN(tags);

	CREATE TABLE IF NOT EXISTS sensors (
		id VARCHAR(255) NOT NULL,
		device_id VARCHAR(255) NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		sensor_type VARCHAR(50) NOT NULL,
		metadata JSONB,
		tags TEXT[],
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		PRIMARY KEY (device_id, id)
	);

	CREATE INDEX IF NOT EXISTS idx_sensors_device_id ON sensors(device_id);
	CREATE INDEX IF NOT EXISTS idx_sensors_tags ON sensors USING GIN(tags);
	CREATE INDEX IF NOT EXISTS idx_sensors_type ON sensors(sensor_type);

	CREATE TABLE IF NOT EXISTS actuators (
		id VARCHAR(255) NOT NULL,
		device_id VARCHAR(255) NOT NULL REFERENCES devices(id) ON DELETE CASCADE,
		name VARCHAR(255) NOT NULL,
		actuator_type VARCHAR(50) NOT NULL,
		metadata JSONB,
		tags TEXT[],
		created_at TIMESTAMP NOT NULL DEFAULT NOW(),
		updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
		PRIMARY KEY (device_id, id)
	);

	CREATE INDEX IF NOT EXISTS idx_actuators_device_id ON actuators(device_id);
	CREATE INDEX IF NOT EXISTS idx_actuators_tags ON actuators USING GIN(tags);
	CREATE INDEX IF NOT EXISTS idx_actuators_type ON actuators(actuator_type);
	`

	// Create trigger functions to enforce tag uniqueness
	triggerFunctions := `
	-- Function to check unique tags for devices
	CREATE OR REPLACE FUNCTION check_device_tags_unique()
	RETURNS TRIGGER AS $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM devices 
			WHERE id != NEW.id 
			AND tags && NEW.tags
		) THEN
			RAISE EXCEPTION 'Tag already exists in another device';
		END IF;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	DROP TRIGGER IF EXISTS device_tags_unique_trigger ON devices;
	CREATE TRIGGER device_tags_unique_trigger
		BEFORE INSERT OR UPDATE ON devices
		FOR EACH ROW EXECUTE FUNCTION check_device_tags_unique();

	-- Function to check unique tags for sensors
	CREATE OR REPLACE FUNCTION check_sensor_tags_unique()
	RETURNS TRIGGER AS $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM sensors 
			WHERE (device_id != NEW.device_id OR id != NEW.id)
			AND tags && NEW.tags
		) THEN
			RAISE EXCEPTION 'Tag already exists in another sensor';
		END IF;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	DROP TRIGGER IF EXISTS sensor_tags_unique_trigger ON sensors;
	CREATE TRIGGER sensor_tags_unique_trigger
		BEFORE INSERT OR UPDATE ON sensors
		FOR EACH ROW EXECUTE FUNCTION check_sensor_tags_unique();

	-- Function to check unique tags for actuators
	CREATE OR REPLACE FUNCTION check_actuator_tags_unique()
	RETURNS TRIGGER AS $$
	BEGIN
		IF EXISTS (
			SELECT 1 FROM actuators 
			WHERE (device_id != NEW.device_id OR id != NEW.id)
			AND tags && NEW.tags
		) THEN
			RAISE EXCEPTION 'Tag already exists in another actuator';
		END IF;
		RETURN NEW;
	END;
	$$ LANGUAGE plpgsql;

	DROP TRIGGER IF EXISTS actuator_tags_unique_trigger ON actuators;
	CREATE TRIGGER actuator_tags_unique_trigger
		BEFORE INSERT OR UPDATE ON actuators
		FOR EACH ROW EXECUTE FUNCTION check_actuator_tags_unique();
	`

	_, err := s.db.ExecContext(ctx, schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	_, err = s.db.ExecContext(ctx, triggerFunctions)
	if err != nil {
		return fmt.Errorf("failed to create trigger functions: %w", err)
	}

	return nil
}

// Device operations

func (s *Storer) createDevice(ctx context.Context, dev *api.Device) error {
	metadata, err := json.Marshal(dev.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Ensure default tag is present
	dev.EnsureDefaultTag()

	query := `
		INSERT INTO devices (id, driver, name, description, metadata, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err = s.db.ExecContext(ctx, query, dev.ID, dev.Driver, dev.Name, dev.Description, metadata, pq.Array(dev.Tags))
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("%w: device with id %s", ErrAlreadyExists, dev.ID)
			}
		}
		return fmt.Errorf("failed to create device: %w", err)
	}

	return nil
}

// CreateDevice creates a new device with its nested sensors and actuators in a transaction
func (s *Storer) CreateDevice(ctx context.Context, dev *api.Device) error {
	ll := s.logCtx(ctx, "device")
	ll.Debug().Str("device_id", dev.ID).Str("driver", string(dev.Driver)).Msg("creating device")
	// Start a transaction
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Create the device
	metadata, err := json.Marshal(dev.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Ensure default tag is present
	dev.EnsureDefaultTag()

	query := `
		INSERT INTO devices (id, driver, name, description, metadata, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err = tx.ExecContext(ctx, query, dev.ID, dev.Driver, dev.Name, dev.Description, metadata, pq.Array(dev.Tags))
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("%w: device with id %s", ErrAlreadyExists, dev.ID)
			}
		}
		return fmt.Errorf("failed to create device: %w", err)
	}

	// Insert nested sensors
	for _, sensor := range dev.Sensors {
		if baseSensor, ok := sensor.(*api.BaseSensor); ok {
			// Ensure device_id is set
			baseSensor.DeviceID = dev.ID

			// Generate default tag if not provided
			if len(baseSensor.Tags) == 0 {
				baseSensor.Tags = []string{baseSensor.DefaultTag(dev.ID)}
			}

			sensorMetadata, err := json.Marshal(baseSensor.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal sensor metadata: %w", err)
			}

			sensorQuery := `
				INSERT INTO sensors (id, device_id, name, sensor_type, metadata, tags, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			`
			_, err = tx.ExecContext(ctx, sensorQuery, baseSensor.ID, baseSensor.DeviceID, baseSensor.Name, baseSensor.SensorType, sensorMetadata, pq.Array(baseSensor.Tags))
			if err != nil {
				if pqErr, ok := err.(*pq.Error); ok {
					if pqErr.Code == "23505" { // unique_violation
						return fmt.Errorf("%w: sensor %s/%s", ErrAlreadyExists, baseSensor.DeviceID, baseSensor.ID)
					}
				}
				return fmt.Errorf("failed to create sensor: %w", err)
			}
		}
	}

	// Insert nested actuators
	for _, actuator := range dev.Actuators {
		if baseActuator, ok := actuator.(*api.BaseActuator); ok {
			// Ensure device_id is set
			baseActuator.DeviceID = dev.ID

			// Generate default tag if not provided
			if len(baseActuator.Tags) == 0 {
				baseActuator.Tags = []string{baseActuator.DefaultTag(dev.ID)}
			}

			actuatorMetadata, err := json.Marshal(baseActuator.Metadata)
			if err != nil {
				return fmt.Errorf("failed to marshal actuator metadata: %w", err)
			}

			actuatorQuery := `
				INSERT INTO actuators (id, device_id, name, actuator_type, metadata, tags, created_at, updated_at)
				VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
			`
			_, err = tx.ExecContext(ctx, actuatorQuery, baseActuator.ID, baseActuator.DeviceID, baseActuator.Name, baseActuator.ActuatorType, actuatorMetadata, pq.Array(baseActuator.Tags))
			if err != nil {
				if pqErr, ok := err.(*pq.Error); ok {
					if pqErr.Code == "23505" { // unique_violation
						return fmt.Errorf("%w: actuator %s/%s", ErrAlreadyExists, baseActuator.DeviceID, baseActuator.ID)
					}
				}
				return fmt.Errorf("failed to create actuator: %w", err)
			}
		}
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetDevice retrieves a device by ID
func (s *Storer) GetDevice(ctx context.Context, id string) (*api.Device, error) {
	ll := s.logCtx(ctx, "device")
	ll.Debug().Str("device_id", id).Msg("getting device")
	query := `
		SELECT id, driver, name, description, metadata, tags
		FROM devices 
		WHERE id = $1
	`

	var dev api.Device
	var metadataJSON []byte
	var tags []string

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&dev.ID, &dev.Driver, &dev.Name, &dev.Description, &metadataJSON, pq.Array(&tags),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: device %s", ErrNotFound, id)
		}
		return nil, fmt.Errorf("failed to get device: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &dev.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	dev.Tags = tags

	// Note: Sensors and Actuators are not stored in DB as they are interfaces
	// They would be reconstructed by the application layer

	return &dev, nil
}

// UpdateDevice updates an existing device
func (s *Storer) UpdateDevice(ctx context.Context, dev *api.Device) error {
	ll := s.logCtx(ctx, "device")
	ll.Debug().Str("device_id", dev.ID).Msg("updating device")
	metadata, err := json.Marshal(dev.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Ensure default tag is present
	dev.EnsureDefaultTag()

	query := `
		UPDATE devices 
		SET driver = $2, name = $3, description = $4, metadata = $5, tags = $6, updated_at = NOW()
		WHERE id = $1
	`
	result, err := s.db.ExecContext(ctx, query, dev.ID, dev.Driver, dev.Name, dev.Description, metadata, pq.Array(dev.Tags))
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("%w: tag conflict", ErrAlreadyExists)
			}
		}
		return fmt.Errorf("failed to update device: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: device %s", ErrNotFound, dev.ID)
	}

	return nil
}

// DeleteDevice deletes a device and all its sensor readings and actuator states (cascading)
func (s *Storer) DeleteDevice(ctx context.Context, id string) error {
	ll := s.logCtx(ctx, "device")
	ll.Debug().Str("device_id", id).Msg("deleting device")
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
		return fmt.Errorf("%w: device %s", ErrNotFound, id)
	}

	return nil
}

// ListDevices retrieves all devices.
func (s *Storer) ListDevices(ctx context.Context) ([]*api.Device, error) {
	ll := s.logCtx(ctx, "device")
	ll.Debug().Msg("listing all devices")
	query := `
		SELECT id, driver, name, description, metadata, tags
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
		var tags []string

		err := rows.Scan(&dev.ID, &dev.Driver, &dev.Name, &dev.Description, &metadataJSON, pq.Array(&tags))
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &dev.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		dev.Tags = tags

		devices = append(devices, &dev)
	}

	return devices, rows.Err()
}

// GetDeviceByTag retrieves a device with a specific tag
func (s *Storer) GetDeviceByTag(ctx context.Context, tag string) (*api.Device, error) {
	ll := s.logCtx(ctx, "device")
	ll.Debug().Str("tag", tag).Msg("getting device by tag")
	query := `
		SELECT id, driver, name, description, metadata, tags
		FROM devices 
		WHERE $1 = ANY(tags)
		LIMIT 1
	`

	var dev api.Device
	var metadataJSON []byte
	var tags []string

	err := s.db.QueryRowContext(ctx, query, tag).Scan(
		&dev.ID, &dev.Driver, &dev.Name, &dev.Description, &metadataJSON, pq.Array(&tags),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: device with tag %s", ErrNotFound, tag)
		}
		return nil, fmt.Errorf("failed to get device by tag: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &dev.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	dev.Tags = tags
	return &dev, nil
}

// ListDevicesByTagPrefix retrieves devices with tags matching a prefix
func (s *Storer) ListDevicesByTagPrefix(ctx context.Context, prefix string) ([]*api.Device, error) {
	ll := s.logCtx(ctx, "device")
	ll.Debug().Str("prefix", prefix).Msg("listing devices by tag prefix")
	query := `
		SELECT DISTINCT id, driver, name, description, metadata, tags
		FROM devices, unnest(tags) AS tag
		WHERE tag LIKE $1
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query, prefix+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query devices by tag prefix: %w", err)
	}
	defer rows.Close()

	return s.scanDevices(rows)
}

// scanDevices is a helper to scan device rows
func (s *Storer) scanDevices(rows *sql.Rows) ([]*api.Device, error) {
	var devices []*api.Device
	for rows.Next() {
		var dev api.Device
		var metadataJSON []byte
		var tags []string

		err := rows.Scan(&dev.ID, &dev.Driver, &dev.Name, &dev.Description, &metadataJSON, pq.Array(&tags))
		if err != nil {
			return nil, fmt.Errorf("failed to scan device: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &dev.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		dev.Tags = tags
		devices = append(devices, &dev)
	}

	return devices, rows.Err()
}

// Sensor operations

// CreateSensor creates a new sensor
func (s *Storer) CreateSensor(ctx context.Context, sensor *api.BaseSensor) error {
	ll := s.logCtx(ctx, "sensor")
	ll.Debug().Str("device_id", sensor.DeviceID).Str("sensor_id", sensor.ID).Str("sensor_type", string(sensor.SensorType)).Msg("creating sensor")
	metadata, err := json.Marshal(sensor.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Generate default tag if not provided
	if len(sensor.Tags) == 0 {
		sensor.Tags = []string{fmt.Sprintf("device.%s.sensor.%s", sensor.DeviceID, sensor.ID)}
	}

	query := `
		INSERT INTO sensors (id, device_id, name, sensor_type, metadata, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err = s.db.ExecContext(ctx, query, sensor.ID, sensor.DeviceID, sensor.Name, sensor.SensorType, metadata, pq.Array(sensor.Tags))
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("%w: sensor %s/%s", ErrAlreadyExists, sensor.DeviceID, sensor.ID)
			}
		}
		return fmt.Errorf("failed to create sensor: %w", err)
	}

	return nil
}

// GetSensor retrieves a sensor by device ID and sensor ID
func (s *Storer) GetSensor(ctx context.Context, deviceID, sensorID string) (*api.BaseSensor, error) {
	ll := s.logCtx(ctx, "sensor")
	ll.Debug().Str("device_id", deviceID).Str("sensor_id", sensorID).Msg("getting sensor")
	query := `
		SELECT id, device_id, name, sensor_type, metadata, tags
		FROM sensors 
		WHERE device_id = $1 AND id = $2
	`

	var sensor api.BaseSensor
	var metadataJSON []byte
	var tags []string

	err := s.db.QueryRowContext(ctx, query, deviceID, sensorID).Scan(
		&sensor.ID, &sensor.DeviceID, &sensor.Name, &sensor.SensorType, &metadataJSON, pq.Array(&tags),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: sensor %s/%s", ErrNotFound, deviceID, sensorID)
		}
		return nil, fmt.Errorf("failed to get sensor: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &sensor.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	sensor.Tags = tags
	return &sensor, nil
}

// UpdateSensor updates an existing sensor
func (s *Storer) UpdateSensor(ctx context.Context, sensor *api.BaseSensor) error {
	ll := s.logCtx(ctx, "sensor")
	ll.Debug().Str("device_id", sensor.DeviceID).Str("sensor_id", sensor.ID).Msg("updating sensor")
	metadata, err := json.Marshal(sensor.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE sensors 
		SET name = $3, sensor_type = $4, metadata = $5, tags = $6, updated_at = NOW()
		WHERE device_id = $1 AND id = $2
	`
	result, err := s.db.ExecContext(ctx, query, sensor.DeviceID, sensor.ID, sensor.Name, sensor.SensorType, metadata, pq.Array(sensor.Tags))
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("%w: tag conflict", ErrAlreadyExists)
			}
		}
		return fmt.Errorf("failed to update sensor: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: sensor %s/%s", ErrNotFound, sensor.DeviceID, sensor.ID)
	}

	return nil
}

// DeleteSensor deletes a sensor by device ID and sensor ID
func (s *Storer) DeleteSensor(ctx context.Context, deviceID, sensorID string) error {
	ll := s.logCtx(ctx, "sensor")
	ll.Debug().Str("device_id", deviceID).Str("sensor_id", sensorID).Msg("deleting sensor")
	query := `DELETE FROM sensors WHERE device_id = $1 AND id = $2`
	result, err := s.db.ExecContext(ctx, query, deviceID, sensorID)
	if err != nil {
		return fmt.Errorf("failed to delete sensor: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: sensor %s/%s", ErrNotFound, deviceID, sensorID)
	}

	return nil
}

// ListSensors retrieves all sensors
func (s *Storer) ListSensors(ctx context.Context) ([]*api.BaseSensor, error) {
	ll := s.logCtx(ctx, "sensor")
	ll.Debug().Msg("listing all sensors")
	query := `
		SELECT id, device_id, name, sensor_type, metadata, tags
		FROM sensors 
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query sensors: %w", err)
	}
	defer rows.Close()

	return s.scanSensors(rows)
}

// ListSensorsByDeviceID retrieves all sensors for a device
func (s *Storer) ListSensorsByDeviceID(ctx context.Context, deviceID string) ([]*api.BaseSensor, error) {
	ll := s.logCtx(ctx, "sensor")
	ll.Debug().Str("device_id", deviceID).Msg("listing sensors by device")
	query := `
		SELECT id, device_id, name, sensor_type, metadata, tags
		FROM sensors 
		WHERE device_id = $1
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query sensors by device: %w", err)
	}
	defer rows.Close()

	return s.scanSensors(rows)
}

// GetSensorByTag retrieves a sensor with a specific tag
func (s *Storer) GetSensorByTag(ctx context.Context, tag string) (*api.BaseSensor, error) {
	ll := s.logCtx(ctx, "sensor")
	ll.Debug().Str("tag", tag).Msg("getting sensor by tag")
	query := `
		SELECT id, device_id, name, sensor_type, metadata, tags
		FROM sensors 
		WHERE $1 = ANY(tags)
		LIMIT 1
	`

	var sensor api.BaseSensor
	var metadataJSON []byte
	var tags []string

	err := s.db.QueryRowContext(ctx, query, tag).Scan(
		&sensor.ID, &sensor.DeviceID, &sensor.Name, &sensor.SensorType, &metadataJSON, pq.Array(&tags),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: sensor with tag %s", ErrNotFound, tag)
		}
		return nil, fmt.Errorf("failed to get sensor by tag: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &sensor.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	sensor.Tags = tags
	return &sensor, nil
}

// ListSensorsByTagPrefix retrieves sensors with tags matching a prefix
func (s *Storer) ListSensorsByTagPrefix(ctx context.Context, prefix string) ([]*api.BaseSensor, error) {
	ll := s.logCtx(ctx, "sensor")
	ll.Debug().Str("prefix", prefix).Msg("listing sensors by tag prefix")
	query := `
		SELECT DISTINCT id, device_id, name, sensor_type, metadata, tags
		FROM sensors, unnest(tags) AS tag
		WHERE tag LIKE $1
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query, prefix+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query sensors by tag prefix: %w", err)
	}
	defer rows.Close()

	return s.scanSensors(rows)
}

// scanSensors is a helper to scan sensor rows
func (s *Storer) scanSensors(rows *sql.Rows) ([]*api.BaseSensor, error) {
	var sensors []*api.BaseSensor
	for rows.Next() {
		var sensor api.BaseSensor
		var metadataJSON []byte
		var tags []string

		err := rows.Scan(&sensor.ID, &sensor.DeviceID, &sensor.Name, &sensor.SensorType, &metadataJSON, pq.Array(&tags))
		if err != nil {
			return nil, fmt.Errorf("failed to scan sensor: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &sensor.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		sensor.Tags = tags
		sensors = append(sensors, &sensor)
	}

	return sensors, rows.Err()
}

// Actuator operations

// CreateActuator creates a new actuator
func (s *Storer) CreateActuator(ctx context.Context, actuator *api.BaseActuator) error {
	ll := s.logCtx(ctx, "actuator")
	ll.Debug().Str("device_id", actuator.DeviceID).Str("actuator_id", actuator.ID).Str("actuator_type", string(actuator.ActuatorType)).Msg("creating actuator")
	metadata, err := json.Marshal(actuator.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Generate default tag if not provided
	if len(actuator.Tags) == 0 {
		actuator.Tags = []string{fmt.Sprintf("device.%s.actuator.%s", actuator.DeviceID, actuator.ID)}
	}

	query := `
		INSERT INTO actuators (id, device_id, name, actuator_type, metadata, tags, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, NOW(), NOW())
	`
	_, err = s.db.ExecContext(ctx, query, actuator.ID, actuator.DeviceID, actuator.Name, actuator.ActuatorType, metadata, pq.Array(actuator.Tags))
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("%w: actuator %s/%s", ErrAlreadyExists, actuator.DeviceID, actuator.ID)
			}
		}
		return fmt.Errorf("failed to create actuator: %w", err)
	}

	return nil
}

// GetActuator retrieves an actuator by device ID and actuator ID
func (s *Storer) GetActuator(ctx context.Context, deviceID, actuatorID string) (*api.BaseActuator, error) {
	ll := s.logCtx(ctx, "actuator")
	ll.Debug().Str("device_id", deviceID).Str("actuator_id", actuatorID).Msg("getting actuator")
	query := `
		SELECT id, device_id, name, actuator_type, metadata, tags
		FROM actuators 
		WHERE device_id = $1 AND id = $2
	`

	var actuator api.BaseActuator
	var metadataJSON []byte
	var tags []string

	err := s.db.QueryRowContext(ctx, query, deviceID, actuatorID).Scan(
		&actuator.ID, &actuator.DeviceID, &actuator.Name, &actuator.ActuatorType, &metadataJSON, pq.Array(&tags),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: actuator %s/%s", ErrNotFound, deviceID, actuatorID)
		}
		return nil, fmt.Errorf("failed to get actuator: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &actuator.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	actuator.Tags = tags
	return &actuator, nil
}

// UpdateActuator updates an existing actuator
func (s *Storer) UpdateActuator(ctx context.Context, actuator *api.BaseActuator) error {
	ll := s.logCtx(ctx, "actuator")
	ll.Debug().Str("device_id", actuator.DeviceID).Str("actuator_id", actuator.ID).Msg("updating actuator")
	metadata, err := json.Marshal(actuator.Metadata)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	query := `
		UPDATE actuators 
		SET name = $3, actuator_type = $4, metadata = $5, tags = $6, updated_at = NOW()
		WHERE device_id = $1 AND id = $2
	`
	result, err := s.db.ExecContext(ctx, query, actuator.DeviceID, actuator.ID, actuator.Name, actuator.ActuatorType, metadata, pq.Array(actuator.Tags))
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok {
			if pqErr.Code == "23505" { // unique_violation
				return fmt.Errorf("%w: tag conflict", ErrAlreadyExists)
			}
		}
		return fmt.Errorf("failed to update actuator: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: actuator %s/%s", ErrNotFound, actuator.DeviceID, actuator.ID)
	}

	return nil
}

// DeleteActuator deletes an actuator by device ID and actuator ID
func (s *Storer) DeleteActuator(ctx context.Context, deviceID, actuatorID string) error {
	ll := s.logCtx(ctx, "actuator")
	ll.Debug().Str("device_id", deviceID).Str("actuator_id", actuatorID).Msg("deleting actuator")
	query := `DELETE FROM actuators WHERE device_id = $1 AND id = $2`
	result, err := s.db.ExecContext(ctx, query, deviceID, actuatorID)
	if err != nil {
		return fmt.Errorf("failed to delete actuator: %w", err)
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}
	if rows == 0 {
		return fmt.Errorf("%w: actuator %s/%s", ErrNotFound, deviceID, actuatorID)
	}

	return nil
}

// ListActuators retrieves all actuators
func (s *Storer) ListActuators(ctx context.Context) ([]*api.BaseActuator, error) {
	ll := s.logCtx(ctx, "actuator")
	ll.Debug().Msg("listing all actuators")
	query := `
		SELECT id, device_id, name, actuator_type, metadata, tags
		FROM actuators 
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to query actuators: %w", err)
	}
	defer rows.Close()

	return s.scanActuators(rows)
}

// ListActuatorsByDeviceID retrieves all actuators for a device
func (s *Storer) ListActuatorsByDeviceID(ctx context.Context, deviceID string) ([]*api.BaseActuator, error) {
	ll := s.logCtx(ctx, "actuator")
	ll.Debug().Str("device_id", deviceID).Msg("listing actuators by device")
	query := `
		SELECT id, device_id, name, actuator_type, metadata, tags
		FROM actuators 
		WHERE device_id = $1
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query, deviceID)
	if err != nil {
		return nil, fmt.Errorf("failed to query actuators by device: %w", err)
	}
	defer rows.Close()

	return s.scanActuators(rows)
}

// GetActuatorByTag retrieves an actuator with a specific tag
func (s *Storer) GetActuatorByTag(ctx context.Context, tag string) (*api.BaseActuator, error) {
	ll := s.logCtx(ctx, "actuator")
	ll.Debug().Str("tag", tag).Msg("getting actuator by tag")
	query := `
		SELECT id, device_id, name, actuator_type, metadata, tags
		FROM actuators 
		WHERE $1 = ANY(tags)
		LIMIT 1
	`

	var actuator api.BaseActuator
	var metadataJSON []byte
	var tags []string

	err := s.db.QueryRowContext(ctx, query, tag).Scan(
		&actuator.ID, &actuator.DeviceID, &actuator.Name, &actuator.ActuatorType, &metadataJSON, pq.Array(&tags),
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("%w: actuator with tag %s", ErrNotFound, tag)
		}
		return nil, fmt.Errorf("failed to get actuator by tag: %w", err)
	}

	if len(metadataJSON) > 0 {
		if err := json.Unmarshal(metadataJSON, &actuator.Metadata); err != nil {
			return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
		}
	}

	actuator.Tags = tags
	return &actuator, nil
}

// ListActuatorsByTagPrefix retrieves actuators with tags matching a prefix
func (s *Storer) ListActuatorsByTagPrefix(ctx context.Context, prefix string) ([]*api.BaseActuator, error) {
	ll := s.logCtx(ctx, "actuator")
	ll.Debug().Str("prefix", prefix).Msg("listing actuators by tag prefix")
	query := `
		SELECT DISTINCT id, device_id, name, actuator_type, metadata, tags
		FROM actuators, unnest(tags) AS tag
		WHERE tag LIKE $1
		ORDER BY name
	`

	rows, err := s.db.QueryContext(ctx, query, prefix+"%")
	if err != nil {
		return nil, fmt.Errorf("failed to query actuators by tag prefix: %w", err)
	}
	defer rows.Close()

	return s.scanActuators(rows)
}

// scanActuators is a helper to scan actuator rows
func (s *Storer) scanActuators(rows *sql.Rows) ([]*api.BaseActuator, error) {
	var actuators []*api.BaseActuator
	for rows.Next() {
		var actuator api.BaseActuator
		var metadataJSON []byte
		var tags []string

		err := rows.Scan(&actuator.ID, &actuator.DeviceID, &actuator.Name, &actuator.ActuatorType, &metadataJSON, pq.Array(&tags))
		if err != nil {
			return nil, fmt.Errorf("failed to scan actuator: %w", err)
		}

		if len(metadataJSON) > 0 {
			if err := json.Unmarshal(metadataJSON, &actuator.Metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
		}

		actuator.Tags = tags
		actuators = append(actuators, &actuator)
	}

	return actuators, rows.Err()
}
