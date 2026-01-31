# Storer Package

The `storer` package provides PostgreSQL-based persistence for device data in the life support system. It handles storage and retrieval of systems, subsystems, devices, sensor readings, and actuator states.

## Features

- **Hierarchical data storage**: Systems contain subsystems, subsystems contain devices
- **Sensor reading history**: Store and query sensor measurements over time
- **Actuator state tracking**: Record actuator state changes
- **Flexible querying**: Filter readings and states by time range, device, sensor type, etc.
- **Data retention**: Clean up old readings with built-in deletion methods
- **Transaction support**: All operations use the provided context for cancellation and timeouts

## Database Schema

The package creates the following tables:

- `systems`: Top-level system definitions
- `subsystems`: Subsystem hierarchy with parent-child relationships
- `devices`: Individual devices within subsystems
- `sensor_readings`: Time-series sensor measurement data
- `actuator_states`: Time-series actuator state data

## Usage

### Initialize

```go
import "lifesupport/backend/pkg/storer"

// Connect to database
store, err := storer.New("postgres://user:pass@localhost:5432/dbname?sslmode=disable")
if err != nil {
    log.Fatal(err)
}
defer store.Close()

// Create tables
ctx := context.Background()
if err := store.InitSchema(ctx); err != nil {
    log.Fatal(err)
}
```

### Store a System

```go
sys := &device.System{
    ID:          "sys-001",
    Name:        "Aquaponics System",
    Description: "Main system",
    CreatedAt:   time.Now(),
    UpdatedAt:   time.Now(),
    Subsystems: []*device.Subsystem{
        {
            ID:   "sub-001",
            Name: "Fish Tank",
            Type: device.SubsystemTypeAquarium,
            Devices: []*device.Device{
                {
                    ID:     "dev-001",
                    Driver: device.DriverShelly,
                    Name:   "Water Monitor",
                },
            },
        },
    },
}

err := store.CreateSystem(ctx, sys)
```

### Store Sensor Readings

```go
reading := &device.SensorReading{
    Value:     25.5,
    Unit:      device.UnitCelsius,
    Timestamp: time.Now(),
    Valid:     true,
}

err := store.StoreSensorReading(ctx, 
    "dev-001",              // device ID
    "sensor-temp-01",       // sensor ID
    "Temperature Sensor",   // sensor name
    device.SensorTypeTemperature,
    reading,
)
```

### Query Sensor Readings

```go
// Get readings with filters
deviceID := "dev-001"
startTime := time.Now().Add(-24 * time.Hour)

filters := storer.SensorReadingFilters{
    DeviceID:  &deviceID,
    StartTime: &startTime,
    Limit:     100,
}

readings, err := store.GetSensorReadings(ctx, filters)

// Get latest reading for a sensor
latest, err := store.GetLatestSensorReading(ctx, "sensor-temp-01")
```

### Store Actuator States

```go
state := &device.ActuatorState{
    Active: true,
    Parameters: map[string]float64{
        "brightness": 75.0,
    },
    Timestamp: time.Now(),
}

err := store.StoreActuatorState(ctx,
    "dev-001",                          // device ID
    "actuator-light-01",                // actuator ID
    "LED Light",                        // actuator name
    device.ActuatorTypeDimmableLight,
    state,
)
```

### Query Actuator States

```go
// Get states with filters
actuatorID := "actuator-light-01"
filters := storer.ActuatorStateFilters{
    ActuatorID: &actuatorID,
    Limit:      50,
}

states, err := store.GetActuatorStates(ctx, filters)

// Get latest state
latest, err := store.GetLatestActuatorState(ctx, "actuator-light-01")
```

### Retrieve Hierarchical Data

```go
// Get entire system with all subsystems and devices
sys, err := store.GetSystem(ctx, "sys-001")

// Get single subsystem with devices
subsystem, err := store.GetSubsystem(ctx, "sub-001")

// Get single device
device, err := store.GetDevice(ctx, "dev-001")
```

### Update Operations

```go
// Update system
sys.Name = "Updated System Name"
sys.UpdatedAt = time.Now()
err := store.UpdateSystem(ctx, sys)

// Update subsystem
subsystem.Description = "New description"
err := store.UpdateSubsystem(ctx, subsystem)

// Update device
device.Name = "Updated Device Name"
err := store.UpdateDevice(ctx, device)
```

### Delete Operations

```go
// Delete system (cascades to subsystems and devices)
err := store.DeleteSystem(ctx, "sys-001")

// Delete subsystem (cascades to devices)
err := store.DeleteSubsystem(ctx, "sub-001")

// Delete device (cascades to readings and states)
err := store.DeleteDevice(ctx, "dev-001")
```

### Data Retention

```go
// Delete sensor readings older than 30 days
cutoff := time.Now().Add(-30 * 24 * time.Hour)
deletedCount, err := store.DeleteOldSensorReadings(ctx, cutoff)

// Delete actuator states older than 30 days
deletedCount, err := store.DeleteOldActuatorStates(ctx, cutoff)
```

## Filter Types

### SensorReadingFilters

All fields are optional:

- `DeviceID *string`: Filter by device ID
- `SensorID *string`: Filter by sensor ID
- `SensorType *device.SensorType`: Filter by sensor type
- `StartTime *time.Time`: Get readings after this time
- `EndTime *time.Time`: Get readings before this time
- `Limit int`: Maximum number of results

### ActuatorStateFilters

All fields are optional:

- `DeviceID *string`: Filter by device ID
- `ActuatorID *string`: Filter by actuator ID
- `ActuatorType *device.ActuatorType`: Filter by actuator type
- `StartTime *time.Time`: Get states after this time
- `EndTime *time.Time`: Get states before this time
- `Limit int`: Maximum number of results

## Notes

- Sensors and Actuators themselves are not persisted (they are interfaces). Only their readings/states are stored.
- The Device struct returned from `GetDevice()` will have empty Sensors and Actuators slices. These should be reconstructed by the application layer based on device configuration.
- All timestamps are stored in UTC.
- JSON fields (metadata, parameters) support arbitrary key-value data.
- Cascading deletes ensure referential integrity.
- Use contexts with timeouts for long-running queries.
