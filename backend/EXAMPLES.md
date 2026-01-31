# API Request Examples

Collection of example JSON payloads for testing the Life Support API.

## Create System (Complete Hierarchy)

```json
{
  "id": "sys-001",
  "name": "Aquaponics System",
  "description": "Main aquaponics life support system",
  "subsystems": [
    {
      "id": "sub-001",
      "name": "Fish Tank",
      "description": "Main fish tank subsystem",
      "type": "aquarium",
      "metadata": {
        "capacity": "500L",
        "species": "tilapia"
      },
      "devices": [
        {
          "id": "dev-001",
          "driver": "shelly",
          "name": "Water Monitor",
          "description": "Temperature and pH monitoring device",
          "metadata": {
            "location": "tank-center",
            "version": "1.0"
          }
        }
      ],
      "subsystems": [
        {
          "id": "sub-002",
          "name": "Filtration",
          "type": "filtration",
          "devices": [
            {
              "id": "dev-002",
              "driver": "station",
              "name": "Filter Pump",
              "metadata": {
                "flow_rate": "1000L/h"
              }
            }
          ]
        }
      ]
    },
    {
      "id": "sub-003",
      "name": "Grow Beds",
      "type": "hydroponics",
      "metadata": {
        "count": "4",
        "total_capacity": "400L"
      }
    }
  ]
}
```

## Create Subsystem

```json
{
  "subsystem": {
    "id": "sub-004",
    "name": "LED Lighting System",
    "description": "Automated grow lights",
    "type": "lighting",
    "metadata": {
      "power": "200W",
      "spectrum": "full"
    }
  },
  "system_id": "sys-001"
}
```

## Create Device

```json
{
  "device": {
    "id": "dev-003",
    "driver": "shelly",
    "name": "pH Controller",
    "description": "Automated pH adjustment",
    "metadata": {
      "version": "2.1",
      "location": "main-tank"
    }
  },
  "subsystem_id": "sub-001"
}
```

## Store Sensor Reading - Temperature

```json
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

## Store Sensor Reading - pH

```json
{
  "device_id": "dev-001",
  "sensor_id": "sensor-ph-01",
  "sensor_name": "pH Sensor",
  "sensor_type": "ph",
  "reading": {
    "value": 7.2,
    "unit": "pH",
    "timestamp": "2026-01-31T10:30:00Z",
    "valid": true
  }
}
```

## Store Sensor Reading - Flow Rate

```json
{
  "device_id": "dev-002",
  "sensor_id": "sensor-flow-01",
  "sensor_name": "Flow Rate Sensor",
  "sensor_type": "flow_rate",
  "reading": {
    "value": 15.5,
    "unit": "L/min",
    "timestamp": "2026-01-31T10:30:00Z",
    "valid": true
  }
}
```

## Store Sensor Reading - Dissolved Oxygen

```json
{
  "device_id": "dev-001",
  "sensor_id": "sensor-do-01",
  "sensor_name": "Dissolved Oxygen Sensor",
  "sensor_type": "dissolved_oxygen",
  "reading": {
    "value": 8.5,
    "unit": "mg/L",
    "timestamp": "2026-01-31T10:30:00Z",
    "valid": true
  }
}
```

## Store Sensor Reading - Invalid/Error

```json
{
  "device_id": "dev-001",
  "sensor_id": "sensor-temp-01",
  "sensor_name": "Temperature Sensor",
  "sensor_type": "temperature",
  "reading": {
    "value": 0,
    "unit": "°C",
    "timestamp": "2026-01-31T10:30:00Z",
    "valid": false,
    "error": "Sensor connection timeout"
  }
}
```

## Store Actuator State - Relay (On)

```json
{
  "device_id": "dev-002",
  "actuator_id": "actuator-pump-01",
  "actuator_name": "Water Pump",
  "actuator_type": "relay",
  "state": {
    "active": true,
    "timestamp": "2026-01-31T10:30:00Z"
  }
}
```

## Store Actuator State - Dimmable Light

```json
{
  "device_id": "dev-003",
  "actuator_id": "actuator-light-01",
  "actuator_name": "LED Grow Light",
  "actuator_type": "dimmable_light",
  "state": {
    "active": true,
    "parameters": {
      "brightness": 75.0,
      "red": 80.0,
      "blue": 90.0
    },
    "timestamp": "2026-01-31T10:30:00Z"
  }
}
```

## Store Actuator State - Peristaltic Pump

```json
{
  "device_id": "dev-003",
  "actuator_id": "actuator-dosing-01",
  "actuator_name": "pH Up Dosing Pump",
  "actuator_type": "peristaltic_pump",
  "state": {
    "active": true,
    "parameters": {
      "flow_rate": 2.5,
      "volume_dispensed": 10.0
    },
    "timestamp": "2026-01-31T10:30:00Z"
  }
}
```

## Store Actuator State - Servo

```json
{
  "device_id": "dev-004",
  "actuator_id": "actuator-valve-01",
  "actuator_name": "Inlet Valve",
  "actuator_type": "servo",
  "state": {
    "active": true,
    "parameters": {
      "position": 45.0,
      "angle": 90.0
    },
    "timestamp": "2026-01-31T10:30:00Z"
  }
}
```

## Update System

```json
{
  "name": "Updated Aquaponics System",
  "description": "Updated with new sensors"
}
```

## Update Subsystem

```json
{
  "name": "Updated Fish Tank",
  "description": "Increased capacity",
  "type": "aquarium",
  "metadata": {
    "capacity": "750L",
    "species": "tilapia"
  }
}
```

## Update Device

```json
{
  "driver": "shelly",
  "name": "Advanced Water Monitor",
  "description": "Upgraded monitoring device",
  "metadata": {
    "location": "tank-center",
    "version": "2.0",
    "firmware": "3.1.4"
  }
}
```

## Cleanup Old Data

```json
{
  "days_old": 30
}
```

---

## Query Parameter Examples

### Get Sensor Readings

```
GET /api/sensor-readings?device_id=dev-001&limit=100
GET /api/sensor-readings?sensor_id=sensor-temp-01
GET /api/sensor-readings?sensor_type=temperature&limit=50
GET /api/sensor-readings?start_time=2026-01-30T00:00:00Z&end_time=2026-01-31T23:59:59Z
GET /api/sensor-readings?device_id=dev-001&sensor_type=ph&start_time=2026-01-31T00:00:00Z&limit=20
```

### Get Actuator States

```
GET /api/actuator-states?device_id=dev-002&limit=100
GET /api/actuator-states?actuator_id=actuator-pump-01
GET /api/actuator-states?actuator_type=relay&limit=50
GET /api/actuator-states?start_time=2026-01-30T00:00:00Z&end_time=2026-01-31T23:59:59Z
GET /api/actuator-states?device_id=dev-003&actuator_type=dimmable_light&limit=20
```
