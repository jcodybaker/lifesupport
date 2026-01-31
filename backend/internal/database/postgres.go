package database

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

type PostgresDB struct {
	DB *sql.DB
}

func NewPostgresDB() (*PostgresDB, error) {
	connStr := fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		os.Getenv("POSTGRES_HOST"),
		os.Getenv("POSTGRES_PORT"),
		os.Getenv("POSTGRES_USER"),
		os.Getenv("POSTGRES_PASSWORD"),
		os.Getenv("POSTGRES_DB"),
	)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres connection: %w", err)
	}

	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping postgres: %w", err)
	}

	log.Println("Connected to PostgreSQL")
	return &PostgresDB{DB: db}, nil
}

func (p *PostgresDB) Close() error {
	return p.DB.Close()
}

func (p *PostgresDB) InitSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS devices (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(50) NOT NULL,
		shelly_id VARCHAR(255) NOT NULL,
		status VARCHAR(50) DEFAULT 'off',
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		enabled BOOLEAN DEFAULT true
	);

	CREATE TABLE IF NOT EXISTS sensors (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		type VARCHAR(50) NOT NULL,
		unit VARCHAR(20) NOT NULL,
		location VARCHAR(255),
		last_value DECIMAL(10, 2),
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		enabled BOOLEAN DEFAULT true
	);

	CREATE TABLE IF NOT EXISTS cameras (
		id SERIAL PRIMARY KEY,
		name VARCHAR(255) NOT NULL,
		url VARCHAR(512) NOT NULL,
		location VARCHAR(255),
		enabled BOOLEAN DEFAULT true,
		last_updated TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS alerts (
		id SERIAL PRIMARY KEY,
		type VARCHAR(50) NOT NULL,
		message TEXT NOT NULL,
		source VARCHAR(255),
		acknowledged BOOLEAN DEFAULT false,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
		resolved_at TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS users (
		id SERIAL PRIMARY KEY,
		username VARCHAR(255) UNIQUE NOT NULL,
		password_hash VARCHAR(255) NOT NULL,
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);

	CREATE INDEX IF NOT EXISTS idx_alerts_acknowledged ON alerts(acknowledged);
	CREATE INDEX IF NOT EXISTS idx_devices_enabled ON devices(enabled);
	CREATE INDEX IF NOT EXISTS idx_sensors_enabled ON sensors(enabled);
	`

	_, err := p.DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to initialize schema: %w", err)
	}

	log.Println("PostgreSQL schema initialized")
	return nil
}
