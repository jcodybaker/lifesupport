#!/bin/bash
# Sample API calls for testing the Life Support System API

BASE_URL="http://localhost:8080/api"

echo "=== Creating a System with Hierarchy ==="
curl -X POST "$BASE_URL/systems" \
  -H "Content-Type: application/json" \
  -d '{
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
        ]
      }
    ]
  }'
echo -e "\n\n"

echo "=== Getting System ==="
curl -X GET "$BASE_URL/systems/sys-001"
echo -e "\n\n"

echo "=== Creating a Sensor Reading ==="
curl -X POST "$BASE_URL/sensor-readings" \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "dev-001",
    "sensor_id": "sensor-temp-01",
    "sensor_name": "Temperature Sensor",
    "sensor_type": "temperature",
    "reading": {
      "value": 25.5,
      "unit": "°C",
      "timestamp": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",
      "valid": true
    }
  }'
echo -e "\n\n"

echo "=== Creating Another Sensor Reading ==="
curl -X POST "$BASE_URL/sensor-readings" \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "dev-001",
    "sensor_id": "sensor-ph-01",
    "sensor_name": "pH Sensor",
    "sensor_type": "ph",
    "reading": {
      "value": 7.2,
      "unit": "pH",
      "timestamp": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'",
      "valid": true
    }
  }'
echo -e "\n\n"

echo "=== Getting All Sensor Readings for Device ==="
curl -X GET "$BASE_URL/sensor-readings?device_id=dev-001&limit=10"
echo -e "\n\n"

echo "=== Getting Latest Temperature Reading ==="
curl -X GET "$BASE_URL/sensor-readings/sensor-temp-01/latest"
echo -e "\n\n"

echo "=== Creating an Actuator State ==="
curl -X POST "$BASE_URL/actuator-states" \
  -H "Content-Type: application/json" \
  -d '{
    "device_id": "dev-001",
    "actuator_id": "actuator-light-01",
    "actuator_name": "LED Light",
    "actuator_type": "dimmable_light",
    "state": {
      "active": true,
      "parameters": {
        "brightness": 75.0
      },
      "timestamp": "'$(date -u +"%Y-%m-%dT%H:%M:%SZ")'"
    }
  }'
echo -e "\n\n"

echo "=== Getting Latest Actuator State ==="
curl -X GET "$BASE_URL/actuator-states/actuator-light-01/latest"
echo -e "\n\n"

echo "=== Creating a Device in Existing Subsystem ==="
curl -X POST "$BASE_URL/devices" \
  -H "Content-Type: application/json" \
  -d '{
    "device": {
      "id": "dev-002",
      "driver": "station",
      "name": "pH Controller",
      "description": "Automated pH adjustment system"
    },
    "subsystem_id": "sub-001"
  }'
echo -e "\n\n"

echo "=== Getting Subsystem with Devices ==="
curl -X GET "$BASE_URL/subsystems/sub-001"
echo -e "\n\n"

echo "=== Querying Temperature Readings ==="
curl -X GET "$BASE_URL/sensor-readings?sensor_type=temperature&limit=5"
echo -e "\n\n"

echo "=== Updating System ==="
curl -X PUT "$BASE_URL/systems/sys-001" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Updated Aquaponics System",
    "description": "Updated description"
  }'
echo -e "\n\n"

echo "=== Getting Updated System ==="
curl -X GET "$BASE_URL/systems/sys-001"
echo -e "\n\n"

echo "All tests completed!"
