package database

import (
	"context"
	"crypto/tls"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/ClickHouse/clickhouse-go/v2/lib/driver"
)

type ClickHouseDB struct {
	Conn driver.Conn
}

func NewClickHouseDB() (*ClickHouseDB, error) {
	conn, err := clickhouse.Open(&clickhouse.Options{
		Addr: []string{fmt.Sprintf("%s:%s", os.Getenv("CLICKHOUSE_HOST"), os.Getenv("CLICKHOUSE_PORT"))},
		Auth: clickhouse.Auth{
			Database: os.Getenv("CLICKHOUSE_DB"),
			Username: os.Getenv("CLICKHOUSE_USER"),
			Password: os.Getenv("CLICKHOUSE_PASSWORD"),
		},
		TLS: &tls.Config{
			InsecureSkipVerify: true,
		},
		DialTimeout: 5 * time.Second,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to connect to clickhouse: %w", err)
	}

	if err := conn.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("failed to ping clickhouse: %w", err)
	}

	log.Println("Connected to ClickHouse")
	return &ClickHouseDB{Conn: conn}, nil
}

func (c *ClickHouseDB) Close() error {
	return c.Conn.Close()
}

func (c *ClickHouseDB) InitSchema() error {
	ctx := context.Background()

	// Create sensor_readings table
	query := `
	CREATE TABLE IF NOT EXISTS sensor_readings (
		sensor_id Int32,
		timestamp DateTime,
		value Float64
	) ENGINE = MergeTree()
	ORDER BY (sensor_id, timestamp)
	TTL timestamp + INTERVAL 90 DAY
	`

	if err := c.Conn.Exec(ctx, query); err != nil {
		return fmt.Errorf("failed to create sensor_readings table: %w", err)
	}

	log.Println("ClickHouse schema initialized")
	return nil
}

func (c *ClickHouseDB) InsertReading(ctx context.Context, sensorID int, timestamp time.Time, value float64) error {
	query := `INSERT INTO sensor_readings (sensor_id, timestamp, value) VALUES (?, ?, ?)`
	return c.Conn.Exec(ctx, query, sensorID, timestamp, value)
}

func (c *ClickHouseDB) GetReadings(ctx context.Context, sensorID int, start, end time.Time) ([]struct {
	Timestamp time.Time
	Value     float64
}, error) {
	query := `
		SELECT timestamp, value 
		FROM sensor_readings 
		WHERE sensor_id = ? AND timestamp BETWEEN ? AND ?
		ORDER BY timestamp ASC
	`

	rows, err := c.Conn.Query(ctx, query, sensorID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var readings []struct {
		Timestamp time.Time
		Value     float64
	}

	for rows.Next() {
		var r struct {
			Timestamp time.Time
			Value     float64
		}
		if err := rows.Scan(&r.Timestamp, &r.Value); err != nil {
			return nil, err
		}
		readings = append(readings, r)
	}

	return readings, rows.Err()
}
