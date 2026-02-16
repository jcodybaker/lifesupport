# LifeSupport

A full-stack life support system management application with a Svelte frontend and Go API backend, using PostgreSQL for data persistence.

Manages hierarchical systems of devices, sensors, and actuators for aquaponics, hydroponics, and other life support applications.

## Project Structure

```
lifesupport/
├── backend/              # Go API server
│   ├── main.go          # Server entry point
│   ├── pkg/
│   │   ├── api/         # HTTP handlers and routing
│   │   │   ├── handlers.go
│   │   │   ├── router.go
│   │   │   ├── handlers_test.go
│   │   │   └── README.md
│   │   ├── device/      # Device, sensor, and actuator types
│   │   └── storer/      # PostgreSQL persistence layer
│   ├── API.md           # Complete API documentation
│   └── test-api.sh      # Shell script for manual API testing
├── frontend/            # Svelte web application
└── docker-compose.yml
```

## Prerequisites

- Go 1.21 or higher
- Node.js 18 or higher
- Docker and Docker Compose (for PostgreSQL)
- PostgreSQL (if not using Docker)

## Quick Start

### 1. Start Services

Using Docker Compose (starts PostgreSQL, Temporal, Elasticsearch, and Temporal Web UI):
```bash
docker-compose up -d
```

This starts:
- **PostgreSQL** (port 5432) - Application database
- **Temporal PostgreSQL** (port 5433) - Temporal metadata storage
- **Elasticsearch** (port 9200) - Temporal visibility
- **Temporal Server** (port 7233) - Workflow engine
- **Temporal Web UI** (port 8088) - Workflow debugging interface at http://localhost:8088

Or use your own PostgreSQL and Temporal instances and update connection strings accordingly.

### 2. Start the Backend

The backend now uses Cobra CLI and supports two modes:

**HTTP API Server:**
```bash
cd backend

# Using Make (recommended)
make build
make run-http

# Or directly
./lifesupport-backend http

# With custom port
./lifesupport-backend http --port 3000
```

The API will be available at `http://localhost:8080`

**Temporal Worker:**
```bash
cd backend

# Make sure docker-compose is running first (Temporal needs to be available)

# Using Make
make run-worker

# Or directly
./lifesupport-backend worker

# With custom options
./lifesupport-backend worker \
  --temporal-host localhost:7233 \
  --task-queue lifesupport-tasks
```

For more details on the Temporal worker, see [backend/TEMPORAL.md](backend/TEMPORAL.md).

### 3. Start the Frontend

In a new terminal:
```bash
cd frontend
npm install
npm run dev
```

The frontend will be available at `http://localhost:5173`

## API Documentation

For complete API documentation, see [backend/API.md](backend/API.md).

### Main Endpoints

**Systems**
- `POST /api/systems` - Create a system with hierarchy
- `GET /api/systems/{id}` - Get system with all subsystems and devices
- `PUT /api/systems/{id}` - Update system
- `DELETE /api/systems/{id}` - Delete system

**Subsystems**
- `POST /api/subsystems` - Create subsystem
- `GET /api/subsystems/{id}` - Get subsystem with devices
- `PUT /api/subsystems/{id}` - Update subsystem
- `DELETE /api/subsystems/{id}` - Delete subsystem

**Devices**
- `POST /api/devices` - Create device
- `GET /api/devices/{id}` - Get device
- `PUT /api/devices/{id}` - Update device
- `DELETE /api/devices/{id}` - Delete device

**Sensor Readings**
- `POST /api/sensor-readings` - Store sensor reading
- `GET /api/sensor-readings` - Query readings with filters
- `GET /api/sensor-readings/{sensorId}/latest` - Get latest reading

**Actuator States**
- `POST /api/actuator-states` - Store actuator state
- `GET /api/actuator-states` - Query states with filters
- `GET /api/actuator-states/{actuatorId}/latest` - Get latest state

**Maintenance**
- `POST /api/maintenance/cleanup-readings` - Delete old sensor readings
- `POST /api/maintenance/cleanup-states` - Delete old actuator states

## Database Schema

The application uses the following PostgreSQL tables:
- `systems` - Top-level system definitions
- `subsystems` - Subsystem hierarchy with parent-child relationships
- `devices` - Individual devices within subsystems
- `sensor_readings` - Time-series sensor measurement data
- `actuator_states` - Time-series actuator state data

The schema is automatically created on server startup.

## Environment Variables

### Backend

Create `backend/.env`:
```
DATABASE_URL=host=localhost port=5432 user=postgres password=postgres dbname=lifesupport sslmode=disable
PORT=8080
```

## Development

- Backend: Go server with gorilla/mux for routing, lib/pq for PostgreSQL
- Frontend: Svelte with Vite for fast development
- Database: PostgreSQL 16
- Architecture: Hierarchical device management with time-series data storage
- Testing: Comprehensive Go tests for all API handlers

### Running Tests

Backend API tests:
```bash
cd backend

# Setup test database (first time only)
make setup-test-db

# Run all tests
make test

# Run with verbose output
make test-verbose

# Run API tests only
make test-api

# Run with coverage
make test-cover

# Generate HTML coverage report
make coverage
```

See all available commands:
```bash
make help
```

## Features

- ✅ Hierarchical system management (Systems → Subsystems → Devices)
- ✅ Time-series sensor reading storage and retrieval
- ✅ Actuator state tracking
- ✅ Flexible querying with filters (device, sensor type, time range, etc.)
- ✅ RESTful CRUD API
- ✅ CORS enabled
- ✅ Automatic database initialization
- ✅ Data retention/cleanup utilities

## Device Types

**Sensor Types**: temperature, ph, flow_rate, power, water_depth, humidity, light_level, conductivity, dissolved_oxygen

**Actuator Types**: relay, peristaltic_pump, dimmable_light, servo, valve

**Subsystem Types**: aquarium, hydroponics, reservoir, filtration, lighting, nutrient_dosing, water_exchange, environmental

## License

MIT
