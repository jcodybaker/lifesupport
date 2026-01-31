# 🐠 Life Support System - Setup Checklist

## ✅ Pre-Deployment Checklist

### Security (Critical!)
- [ ] Change default admin password
  ```bash
  cd backend
  go run cmd/create_admin/main.go admin <your-secure-password>
  ```
- [ ] Update JWT secret in `backend/.env`
  ```
  JWT_SECRET=<generate-a-random-secure-string>
  ```
- [ ] Remove or secure default credentials from `init.sql`

### Configuration
- [ ] Update `backend/.env` with your database credentials
- [ ] Update `frontend/.env` with your API URL
- [ ] Configure CORS_ORIGIN in `backend/.env` to match your frontend URL
- [ ] Set proper PostgreSQL password in `docker-compose.yml`

### Hardware Integration
- [ ] Add your actual Shelly device IDs to the database
  ```sql
  UPDATE devices SET shelly_id = 'your-actual-shelly-id' WHERE id = 1;
  ```
- [ ] Configure your camera URLs
  ```sql
  UPDATE cameras SET url = 'http://your-camera-url/stream' WHERE id = 1;
  ```
- [ ] Update sensor configurations to match your hardware
- [ ] Implement sensor data collection (write to ClickHouse)
- [ ] Integrate Shelly API calls in `internal/services/shelly.go`

### Testing
- [ ] Test login with admin credentials
- [ ] Verify database connections
- [ ] Test device control endpoints
- [ ] Verify sensor data queries
- [ ] Check camera feeds
- [ ] Test alert creation and acknowledgment
- [ ] Verify CORS is working from frontend

### Deployment
- [ ] Build Docker images
  ```bash
  docker-compose build
  ```
- [ ] Test full stack with Docker Compose
  ```bash
  docker-compose up
  ```
- [ ] Verify all services are healthy
  ```bash
  docker-compose ps
  ```
- [ ] Check application logs
  ```bash
  docker-compose logs -f
  ```

### Monitoring
- [ ] Set up database backups
- [ ] Configure log rotation
- [ ] Set up system monitoring
- [ ] Configure alert notifications (email/SMS)
- [ ] Test disaster recovery procedures

## 🔧 Hardware Setup Tasks

### Sensors
- [ ] Install temperature sensor
- [ ] Install pH sensor
- [ ] Install flow sensor
- [ ] Install weight sensor
- [ ] Install distance/level sensor
- [ ] Calibrate all sensors
- [ ] Test sensor readings

### Devices
- [ ] Connect pumps to Shelly devices
- [ ] Connect lights to Shelly devices
- [ ] Connect valves to Shelly devices
- [ ] Test device on/off control
- [ ] Verify device status reporting

### Cameras
- [ ] Install cameras at viewing points
- [ ] Configure camera streaming
- [ ] Test camera feeds
- [ ] Set up network connectivity

## 📊 Data Setup

### Initial Configuration
- [ ] Create all sensor entries in database
- [ ] Create all device entries in database
- [ ] Create all camera entries in database
- [ ] Set up initial alert rules
- [ ] Configure notification preferences

### Testing Data
- [ ] Insert test sensor readings
  ```sql
  INSERT INTO sensor_readings (sensor_id, timestamp, value)
  VALUES (1, now(), 23.5);
  ```
- [ ] Create test alerts
- [ ] Verify data appears in UI

## 🚀 Production Readiness

### Performance
- [ ] Test with multiple concurrent users
- [ ] Verify database query performance
- [ ] Check ClickHouse data retention (90 days)
- [ ] Monitor memory usage
- [ ] Monitor CPU usage

### Security
- [ ] Enable HTTPS/TLS
- [ ] Set up firewall rules
- [ ] Restrict database access
- [ ] Enable database SSL connections
- [ ] Review CORS configuration
- [ ] Set up rate limiting (if needed)

### Backup & Recovery
- [ ] Set up automated PostgreSQL backups
- [ ] Set up automated ClickHouse backups
- [ ] Test restore procedures
- [ ] Document recovery process

### Documentation
- [ ] Document your specific hardware setup
- [ ] Create maintenance procedures
- [ ] Document troubleshooting steps
- [ ] Create emergency contact list

## 🎯 Next Development Phase

### Temporal Integration (Future)
- [ ] Design workflow for automated feeding
- [ ] Design workflow for water changes
- [ ] Design workflow for light schedules
- [ ] Design workflow for alert escalation
- [ ] Implement Temporal worker
- [ ] Test workflows in isolation

### Feature Enhancements (Future)
- [ ] Add data visualization charts
- [ ] Implement WebSocket for real-time updates
- [ ] Add email notifications
- [ ] Add SMS notifications
- [ ] Create mobile app or PWA
- [ ] Add historical data export
- [ ] Implement data analytics

## 📝 Notes

- Keep this checklist updated as you progress
- Mark items complete as you finish them
- Add your own specific items as needed
- Review regularly to ensure nothing is missed

## 🆘 Support

If you encounter issues:
1. Check `QUICK_REFERENCE.md` for common tasks
2. Review `PROJECT_SUMMARY.md` for architecture details
3. Check logs: `docker-compose logs -f`
4. Verify database connectivity
5. Test API endpoints with curl
6. Check browser console for frontend errors

---

**Last Updated**: Initial setup - adjust as you customize the system
