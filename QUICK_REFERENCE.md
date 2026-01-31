# Life Support - Quick Reference

## 🚀 Getting Started

### Option 1: Docker Compose (Easiest)
```bash
# Copy environment files
cp backend/.env.example backend/.env
cp frontend/.env.example frontend/.env

# Start everything
docker-compose up -d

# View logs
docker-compose logs -f

# Stop everything
docker-compose down
```

### Option 2: Local Development
```bash
# Start databases only
./start-dev.sh

# In terminal 1 - Backend
cd backend
go run cmd/server/main.go

# In terminal 2 - Frontend
cd frontend
npm run dev
```

## 🔐 User Management

### Create/Update Admin User
```bash
cd backend
go run cmd/create_admin/main.go <username> <password>
```

### Default Login
- Username: `admin`
- Password: `admin123`

## 📊 Database Access

### PostgreSQL
```bash
# Using Docker
docker-compose exec postgres psql -U lifesupport -d lifesupport

# Local
psql -h localhost -U lifesupport -d lifesupport
```

### ClickHouse
```bash
# Using Docker
docker-compose exec clickhouse clickhouse-client

# HTTP Interface
curl http://localhost:8123/
```

## 🔌 API Testing

### Public Endpoints
```bash
# System status
curl http://localhost:8080/api/status

# Get sensors
curl http://localhost:8080/api/sensors

# Get devices
curl http://localhost:8080/api/devices
```

### Admin Endpoints
```bash
# Login
TOKEN=$(curl -X POST http://localhost:8080/api/login \
  -H "Content-Type: application/json" \
  -d '{"username":"admin","password":"admin123"}' \
  | jq -r '.token')

# Control device
curl -X POST http://localhost:8080/api/admin/devices/1/control \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"action":"toggle"}'
```

## 📝 Common Tasks

### Add a New Sensor
```sql
INSERT INTO sensors (name, type, unit, location, enabled)
VALUES ('New Sensor', 'temperature', '°C', 'Location', true);
```

### Add a New Device
```sql
INSERT INTO devices (name, type, shelly_id, status, enabled)
VALUES ('New Pump', 'pump', 'shelly-id-123', 'off', true);
```

### Create an Alert
```sql
INSERT INTO alerts (type, message, source, acknowledged)
VALUES ('warning', 'Temperature high', 'Temp Sensor', false);
```

### Insert Sensor Reading (ClickHouse)
```sql
INSERT INTO sensor_readings (sensor_id, timestamp, value)
VALUES (1, now(), 23.5);
```

### Query Sensor History
```sql
SELECT timestamp, value 
FROM sensor_readings 
WHERE sensor_id = 1 
  AND timestamp >= now() - INTERVAL 24 HOUR
ORDER BY timestamp DESC;
```

## 🐛 Troubleshooting

### Backend won't start
```bash
# Check database connections
docker-compose ps

# View backend logs
docker-compose logs backend

# Or if running locally
go run cmd/server/main.go
```

### Frontend won't start
```bash
# Reinstall dependencies
cd frontend
rm -rf node_modules package-lock.json
npm install

# Check for port conflicts
lsof -i :5173
```

### Database connection failed
```bash
# Restart databases
docker-compose restart postgres clickhouse

# Check database logs
docker-compose logs postgres
docker-compose logs clickhouse
```

### Reset everything
```bash
# Stop and remove all containers and volumes
docker-compose down -v

# Restart
docker-compose up -d
```

## 📱 URLs

- **Frontend**: http://localhost:5173
- **Backend API**: http://localhost:8080
- **API Health**: http://localhost:8080/health
- **PostgreSQL**: localhost:5432
- **ClickHouse HTTP**: http://localhost:8123
- **ClickHouse Native**: localhost:9000

## 🔧 Configuration Files

- `backend/.env` - Backend configuration
- `frontend/.env` - Frontend configuration
- `docker-compose.yml` - Docker services
- `backend/init.sql` - Database initialization

## 📦 Project Structure

```
lifesupport/
├── backend/              # Go backend
│   ├── cmd/
│   │   ├── server/      # Main server
│   │   └── create_admin/# Admin user utility
│   ├── internal/        # Internal packages
│   └── init.sql         # Database init script
├── frontend/            # Svelte frontend
│   └── src/
│       ├── components/  # UI components
│       └── api.js       # API client
├── docker-compose.yml   # Docker setup
└── start-dev.sh         # Dev startup script
```

## 🎯 Next Steps

1. ✅ Change default admin password
2. ✅ Configure your JWT secret
3. ✅ Add your Shelly device IDs
4. ✅ Set up your sensors
5. ✅ Configure camera URLs
6. 🔄 Implement Shelly API integration
7. 🔄 Add Temporal workflows
8. 🔄 Set up monitoring and alerts
