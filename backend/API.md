# Life Support System API Documentation

RESTful API for managing the life support system, including systems, subsystems, devices, sensor readings, and actuator states.

Base URL: `http://localhost:8080/api`

## Systems

### Create System
```http
POST /api/systems
Content-Type: application/json

{
  "id": "sys-001",
  "name": "Aquaponics System",
  "description": "Main life support system",
  "subsystems": [
    {
      "id": "sub-001",
      "name": "Fish Tank",
      "type": "aquarium",
      "devices": [
        {
          "id": "dev-001",
          "driver": "shelly",
          "name": "Water Monitor",
          "metadata": {
            "location": "tank-center"
          }
        }
      ]
    }
  ]
}
```

Response: `201 Created`

### Get System
```http
GET /api/systems/{id}
```

Response: `200 OK` with system data including all subsystems and devices

### Update System
```http
PUT /api/systems/{id}
Content-Type: application/json

{
  "name": "Updated System Name",
  "description": "Updated description"
}
```

Response: `200 OK`

### Delete System
```http
DELETE /api/systems/{id}
```

Response: `204 No Content`

---

## Subsystems

### Create Subsystem
```http
POST /api/subsystems
Content-Type: application/json

{
  "subsystem": {
    "id": "sub-002",
    "name": "Grow Bed",
    "description": "Hydroponic grow bed",
    "type": "hydroponics",
    "metadata": {
      "capacity": "100L"
    }
  },
  "system_id": "sys-001"
}
```

Response: `201 Created`

### Get Subsystem
```http
GET /api/subsystems/{id}
```

Response: `200 OK` with subsystem data including devices and child subsystems

### Update Subsystem
```http
PUT /api/subsystems/{id}
Content-Type: application/json

{
  "name": "Updated Subsystem",
  "description": "Updated description",
  "type": "hydroponics",
  "metadata": {
    "capacity": "150L"
  }
}
```

Response: `200 OK`

### Delete Subsystem
```http
DELETE /api/subsystems/{id}
```

Response: `204 No Content`

---

## Devices

### Create Device
```http
POST /api/devices
Content-Type: application/json

{
  "device": {
    "id": "dev-002",
    "driver": "station",
    "name": "pH Monitor",
    "description": "Monitors pH levels",
    "metadata": {
      "version": "2.0"
    }
  },
  "subsystem_id": "sub-001"
}
```

Response: `201 Created`

### Get Device
```http
GET /api/devices/{id}
```

Response: `200 OK`

### Update Device
```http
PUT /api/devices/{id}
Content-Type: application/json

{
  "driver": "shelly",
  "name": "Updated Device Name",
  "description": "Updated description",
  "metadata": {
    "version": "2.1"
  }
}
```

Response: `200 OK`

### Delete Device
```http
DELETE /api/devices/{id}
```

Response: `204 No Content`

---

## Sensor Readings

### Store Sensor Reading
```http
POST /api/sensor-readings
Content-Type: application/json

{
  "device_id": "dev-001",
  "sensor_id": "sensor-temp-01",
  "sensor_name": "Temperature Sensor",
  "sensor_type": "temperature",
  "reading": {
    "value": 25.5,
    "unit": "°C",
    "timestamp": "2026-01-31T10:30:00Z",
    "valid": true
  }
}
```

Response: `201 Created`

### Get Sensor Readings
```http
GET /api/sensor-readings?device_id=dev-001&limit=100
GET /api/sensor-readings?sensor_id=sensor-temp-01&start_time=2026-01-30T00:00:00Z&end_time=2026-01-31T23:59:59Z
GET /api/sensor-readings?sensor_type=temperature&limit=50
```

**Query Parameters:**
- `device_id` (optional): Filter by device ID
- `sensor_id` (optional): Filter by sensor ID
- `sensor_type` (optional): Filter by sensor type (temperature, ph, flow_rate, etc.)
- `start_time` (optional): RFC3339 timestamp, readings after this time
- `end_time` (optional): RFC3339 timestamp, readings before this time
- `limit` (optional): Maximum number of results

Response: `200 OK` with array of readings

### Get Latest Sensor Reading
```http
GET /api/sensor-readings/{sensorId}/latest
```

Response: `200 OK` with the most recent reading for the sensor

---

## Actuator States

### Store Actuator State
```http
POST /api/actuator-states
Content-Type: application/json

{
  "device_id": "dev-001",
  "actuator_id": "actuator-light-01",
  "actuator_name": "LED Light",
  "actuator_type": "dimmable_light",
  "state": {
    "active": true,
    "parameters": {
      "brightness": 75.0
    },
    "timestamp": "2026-01-31T10:30:00Z"
  }
}
```

Response: `201 Created`

### Get Actuator States
```http
GET /api/actuator-states?device_id=dev-001&limit=100
GET /api/actuator-states?actuator_id=actuator-light-01&start_time=2026-01-30T00:00:00Z
GET /api/actuator-states?actuator_type=relay&limit=50
```

**Query Parameters:**
- `device_id` (optional): Filter by device ID
- `actuator_id` (optional): Filter by actuator ID
- `actuator_type` (optional): Filter by actuator type (relay, peristaltic_pump, etc.)
- `start_time` (optional): RFC3339 timestamp, states after this time
- `end_time` (optional): RFC3339 timestamp, states before this time
- `limit` (optional): Maximum number of results

Response: `200 OK` with array of states

### Get Latest Actuator State
```http
GET /api/actuator-states/{actuatorId}/latest
```

Response: `200 OK` with the most recent state for the actuator

---

## Maintenance

### Cleanup Old Sensor Readings
```http
POST /api/maintenance/cleanup-readings
Content-Type: application/json

{
  "days_old": 30
}
```

Deletes sensor readings older than the specified number of days.

Response: `200 OK`
```json
{
  "deleted": 1523,
  "cutoff": "2025-12-31T10:30:00Z"
}
```

### Cleanup Old Actuator States
```http
POST /api/maintenance/cleanup-states
Content-Type: application/json

{
  "days_old": 30
}
```

Deletes actuator states older than the specified number of days.

Response: `200 OK`
```json
{
  "deleted": 892,
  "cutoff": "2025-12-31T10:30:00Z"
}
```

---

## Data Types

### Sensor Types
- `temperature`
- `ph`
- `flow_rate`
- `power`
- `water_depth`
- `actuator_status`
- `humidity`
- `light_level`
- `conductivity`
- `dissolved_oxygen`

### Actuator Types
- `relay`
- `peristaltic_pump`
- `dimmable_light`
- `servo`
- `valve`

### Subsystem Types
- `aquarium`
- `hydroponics`
- `reservoir`
- `filtration`
- `lighting`
- `nutrient_dosing`
- `water_exchange`
- `environmental`

### Driver Types
- `shelly`
- `station`

---

## Error Responses

All error responses follow this format:

```json
HTTP/1.1 400 Bad Request
Content-Type: text/plain

Invalid request body: json: cannot unmarshal...
```

Common status codes:
- `400 Bad Request`: Invalid input data
- `404 Not Found`: Resource not found
- `500 Internal Server Error`: Database or server error

---

## CORS

The API supports CORS with the following headers:
- `Access-Control-Allow-Origin: *`
- `Access-Control-Allow-Methods: GET, POST, PUT, DELETE, OPTIONS`
- `Access-Control-Allow-Headers: Content-Type`

---

## Running the Server

```bash
# With default settings (localhost:5432)
go run main.go

# With custom database URL
export DATABASE_URL="postgres://user:pass@localhost:5432/lifesupport?sslmode=disable"
go run main.go

# Custom port
export PORT=3000
go run main.go
```

The server will automatically initialize the database schema on startup.
