-- Initialize PostgreSQL database

-- Create database
CREATE DATABASE lifesupport;

-- Connect to the database
\c lifesupport;

-- Insert sample admin user (password: admin123)
-- You should change this in production!
INSERT INTO users (username, password_hash)
VALUES ('admin', '$2a$10$rqIvK8wPMvJ7.fBfKj.zDuJ3qPNqRQqXYv8yLWGJ7cZH8YH9DGqP6');

-- Insert sample devices
INSERT INTO devices (name, type, shelly_id, status, enabled) VALUES
('Main Water Pump', 'pump', 'shellyplug-123456', 'off', true),
('Grow Light', 'light', 'shellyplug-234567', 'off', true),
('Air Pump', 'pump', 'shellyplug-345678', 'off', true),
('Drain Valve', 'valve', 'shellyplug-456789', 'off', true);

-- Insert sample sensors
INSERT INTO sensors (name, type, unit, location, enabled) VALUES
('Water Temperature', 'temperature', '°C', 'Main Tank', true),
('pH Level', 'ph', 'pH', 'Main Tank', true),
('Water Flow', 'flow', 'L/min', 'Pump Output', true),
('System Weight', 'weight', 'kg', 'Tank Platform', true),
('Water Level', 'distance', 'cm', 'Main Tank', true);

-- Insert sample cameras
INSERT INTO cameras (name, url, location, enabled) VALUES
('Tank Overview', 'http://camera1.local/stream', 'Main Tank', true),
('Fish Close-up', 'http://camera2.local/stream', 'Viewing Window', true);

-- Insert sample alerts
INSERT INTO alerts (type, message, source, acknowledged) VALUES
('warning', 'Water temperature slightly elevated', 'Water Temperature Sensor', false),
('info', 'System startup complete', 'System', true);
