# API Package

HTTP handlers and routing for the Life Support System API.

## Structure

- `handlers.go` - HTTP handler functions for all API endpoints
- `router.go` - Router setup and middleware
- `handlers_test.go` - Comprehensive test suite

## Testing

### Prerequisites

You need a PostgreSQL test database. Create it with:

```bash
./setup-test-db.sh
```

Or manually:

```bash
createdb lifesupport_test
```

### Running Tests

Run all tests:
```bash
go test ./pkg/api -v
```

Run specific test:
```bash
go test ./pkg/api -v -run TestCreateSystem
```

Run with coverage:
```bash
go test ./pkg/api -cover
go test ./pkg/api -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Test Database

Tests use `lifesupport_test` database by default. Configure with environment variable if needed:

```bash
export TEST_DATABASE_URL="host=localhost port=5432 user=postgres password=postgres dbname=lifesupport_test sslmode=disable"
```

The test suite:
- Creates necessary schema automatically
- Cleans up test data after each test
- Skips tests if database is unavailable
- Tests all CRUD operations
- Tests query filtering
- Tests error handling
- Tests invalid input validation

## Usage

```go
import "lifesupport/backend/pkg/api"

// Create handler
store := storer.New(connString)
handler := api.NewHandler(store)

// Setup router
router := handler.SetupRouter()

// Start server
http.ListenAndServe(":8080", router)
```

## Handler Methods

All handlers follow standard http.HandlerFunc signature:

### Systems
- `CreateSystem(w, r)` - POST /api/systems
- `GetSystem(w, r)` - GET /api/systems/{id}
- `UpdateSystem(w, r)` - PUT /api/systems/{id}
- `DeleteSystem(w, r)` - DELETE /api/systems/{id}

### Subsystems
- `CreateSubsystem(w, r)` - POST /api/subsystems
- `GetSubsystem(w, r)` - GET /api/subsystems/{id}
- `UpdateSubsystem(w, r)` - PUT /api/subsystems/{id}
- `DeleteSubsystem(w, r)` - DELETE /api/subsystems/{id}

### Devices
- `CreateDevice(w, r)` - POST /api/devices
- `GetDevice(w, r)` - GET /api/devices/{id}
- `UpdateDevice(w, r)` - PUT /api/devices/{id}
- `DeleteDevice(w, r)` - DELETE /api/devices/{id}

### Sensor Readings
- `StoreSensorReading(w, r)` - POST /api/sensor-readings
- `GetSensorReadings(w, r)` - GET /api/sensor-readings (with query filters)
- `GetLatestSensorReading(w, r)` - GET /api/sensor-readings/{sensorId}/latest

### Actuator States
- `StoreActuatorState(w, r)` - POST /api/actuator-states
- `GetActuatorStates(w, r)` - GET /api/actuator-states (with query filters)
- `GetLatestActuatorState(w, r)` - GET /api/actuator-states/{actuatorId}/latest

### Maintenance
- `CleanupOldReadings(w, r)` - POST /api/maintenance/cleanup-readings
- `CleanupOldStates(w, r)` - POST /api/maintenance/cleanup-states

## Middleware

### CORS Middleware

Enables cross-origin requests:
- Allows all origins (`*`)
- Methods: GET, POST, PUT, DELETE, OPTIONS
- Headers: Content-Type

## Error Handling

All handlers return appropriate HTTP status codes:
- `200 OK` - Successful GET/PUT
- `201 Created` - Successful POST
- `204 No Content` - Successful DELETE
- `400 Bad Request` - Invalid input
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Database/server errors

Error responses are plain text with descriptive messages.
