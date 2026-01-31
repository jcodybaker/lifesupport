# 🐠 Life Support - Hydroponic Aquarium Management System

A full-stack web application for managing a hydroponic aquarium system with goldfish. Monitor sensors, control devices, view camera feeds, and manage system alerts through a modern web interface.

## 🏗️ Architecture

### Frontend
- **Framework**: Svelte with Vite
- **Features**: Responsive dashboard, real-time monitoring, admin authentication
- **Mode**: Anonymous read-only + authenticated admin mode

### Backend
- **Language**: Go (Golang)
- **Framework**: Gin Web Framework
- **Authentication**: JWT-based authentication
- **API**: RESTful API with CORS support

### Databases
- **PostgreSQL**: Transactional data (devices, sensors, cameras, alerts, users)
- **ClickHouse**: Time-series sensor data (optimized for queries)

### Integration
- **Shelly Devices**: Smart device control (pumps, lights, valves)
- **Future**: Temporal workflow orchestration

## 📋 Features

### Monitoring
- 🌡️ Temperature sensors
- ⚗️ pH level monitoring
- 🌊 Water flow tracking
- ⚖️ System weight measurement
- 📏 Distance/level sensors
- 📹 Live camera feeds

### Control (Admin Mode)
- ⚙️ Device control (pumps, lights, valves)
- 🔔 Alert management
- ⚡ Real-time status updates
- 📊 Historical sensor data visualization

### Security
- 🔐 JWT authentication
- 🔒 Role-based access (read-only vs admin)
- 🛡️ CORS protection

## 🚀 Quick Start

### Prerequisites
- Docker & Docker Compose
- Go 1.22+ (for local development)
- Node.js 20+ (for local development)

### Using Docker Compose (Recommended)

1. **Clone the repository**
   ```bash
   git clone <your-repo>
   cd lifesupport
   ```

2. **Configure environment**
   ```bash
   # Backend
   cp backend/.env.example backend/.env
   # Edit backend/.env and set your JWT secret and admin password

   # Frontend
   cp frontend/.env.example frontend/.env
   ```

3. **Start all services**
   ```bash
   docker-compose up -d
   ```

4. **Access the application**
   - Frontend: http://localhost:5173
   - Backend API: http://localhost:8080
   - PostgreSQL: localhost:5432
   - ClickHouse: localhost:8123 (HTTP), localhost:9000 (Native)

### Local Development

#### Backend

1. **Install dependencies**
   ```bash
   cd backend
   go mod download
   ```

2. **Set up databases**
   ```bash
   # Start only the databases
   docker-compose up -d postgres clickhouse
   
   # Initialize PostgreSQL
   psql -h localhost -U lifesupport -d lifesupport -f init.sql
   ```

3. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your settings
   ```

4. **Run the server**
   ```bash
   go run cmd/server/main.go
   ```

#### Frontend

1. **Install dependencies**
   ```bash
   cd frontend
   npm install
   ```

2. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env if needed
   ```

3. **Run development server**
   ```bash
   npm run dev
   ```

## 📁 Project Structure

```
lifesupport/
├── backend/
│   ├── cmd/
│   │   └── server/          # Main application entry point
│   ├── internal/
│   │   ├── api/             # HTTP handlers and routes
│   │   ├── auth/            # Authentication & JWT
│   │   ├── database/        # PostgreSQL & ClickHouse clients
│   │   ├── models/          # Data models
│   │   └── services/        # Business logic (future)
│   ├── go.mod
│   ├── Dockerfile
│   └── .env.example
├── frontend/
│   ├── src/
│   │   ├── components/      # Svelte components
│   │   │   ├── Dashboard.svelte
│   │   │   ├── Login.svelte
│   │   │   ├── SystemStatus.svelte
│   │   │   ├── SensorGrid.svelte
│   │   │   ├── DeviceGrid.svelte
│   │   │   ├── CameraGrid.svelte
│   │   │   └── AlertPanel.svelte
│   │   ├── api.js           # API client
│   │   ├── App.svelte       # Root component
│   │   └── main.js
│   ├── package.json
│   ├── Dockerfile
│   └── .env.example
└── docker-compose.yml
```

## 🔌 API Endpoints

### Public (Read-Only)
- `GET /api/status` - System status
- `GET /api/devices` - List all devices
- `GET /api/sensors` - List all sensors
- `GET /api/sensors/:id/readings?hours=24` - Sensor readings
- `GET /api/cameras` - List all cameras
- `GET /api/alerts` - List all alerts

### Authentication
- `POST /api/login` - Login (returns JWT token)

### Admin Only (Requires JWT)
- `POST /api/admin/devices/:id/control` - Control device
- `PUT /api/admin/devices/:id` - Update device
- `POST /api/admin/devices` - Create device
- `DELETE /api/admin/devices/:id` - Delete device
- `PUT /api/admin/sensors/:id` - Update sensor
- `POST /api/admin/sensors` - Create sensor
- `DELETE /api/admin/sensors/:id` - Delete sensor
- `PUT /api/admin/cameras/:id` - Update camera
- `POST /api/admin/cameras` - Create camera
- `DELETE /api/admin/cameras/:id` - Delete camera
- `PUT /api/admin/alerts/:id/acknowledge` - Acknowledge alert
- `DELETE /api/admin/alerts/:id` - Delete alert

## 🔐 Default Credentials

**Username**: `admin`  
**Password**: `admin123`

⚠️ **Change these immediately in production!**

## 🛠️ Configuration

### Backend Environment Variables
```env
PORT=8080
POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=lifesupport
POSTGRES_PASSWORD=your_password
POSTGRES_DB=lifesupport
CLICKHOUSE_HOST=localhost
CLICKHOUSE_PORT=9000
CLICKHOUSE_USER=default
CLICKHOUSE_PASSWORD=
CLICKHOUSE_DB=sensors
JWT_SECRET=your_jwt_secret_here
CORS_ORIGIN=http://localhost:5173
```

### Frontend Environment Variables
```env
VITE_API_URL=http://localhost:8080/api
```

## 📊 Database Schema

### PostgreSQL Tables
- `devices` - Controllable devices (pumps, lights, valves)
- `sensors` - Sensor configurations
- `cameras` - Camera configurations
- `alerts` - System alerts and notifications
- `users` - Admin users

### ClickHouse Tables
- `sensor_readings` - Time-series sensor data (90-day TTL)

## 🔄 Future Enhancements

- [ ] Temporal workflow integration for automated actions
- [ ] Shelly device API integration
- [ ] Real-time WebSocket updates
- [ ] Sensor data visualization with charts
- [ ] Email/SMS alert notifications
- [ ] Historical data export
- [ ] Mobile app
- [ ] Multi-user management
- [ ] Automated feeding schedules
- [ ] Water quality predictions with ML

## 🤝 Contributing

This is a personal project for managing a hydroponic aquarium system. Feel free to fork and adapt for your own use.

## 📄 License

MIT License - See LICENSE file for details

## 🐟 About

Built with ❤️ for goldfish welfare and aquaponics enthusiasts.
