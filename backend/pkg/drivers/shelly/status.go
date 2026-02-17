package shelly

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"lifesupport/backend/pkg/api"
	"lifesupport/backend/pkg/drivers"

	"github.com/Masterminds/squirrel"
)

func (d *Driver) GetLastStatus(ctx context.Context, opt api.StatusOptions, resource drivers.Statuser) (*api.SensorReading, error) {
	// Query to find the latest event for this resource
	// We filter by src (device ID) and check that params contains the resource ID key
	q := squirrel.Select("timestamp", "params").
		From("rabbitmq.shelly_events").
		Where(squirrel.Eq{"src": resource.GetDeviceID()}).
		Where("JSONHas(params::String, ?, ?)", resource.GetID(), "output").
		OrderBy("timestamp DESC").
		Limit(1)

	if opt.NewerThan != nil {
		q = q.Where(squirrel.Gt{"timestamp": *opt.NewerThan})
	}

	query, args, err := q.
		PlaceholderFormat(squirrel.Question).
		ToSql()
	if err != nil {
		return nil, fmt.Errorf("failed to build query: %w", err)
	}

	rows, err := d.clickhouseConn.Query(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to query latest event: %w", err)
	}
	defer rows.Close()

	if !rows.Next() {
		return nil, fmt.Errorf("no events found for device %s resource %s: %w", resource.GetDeviceID(), resource.GetID(), drivers.ErrNoData)
	}

	var timestamp time.Time
	var paramsJSON string

	if err := rows.Scan(&timestamp, &paramsJSON); err != nil {
		return nil, fmt.Errorf("failed to scan row: %w", err)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("row iteration error: %w", err)
	}

	// Parse the params JSON to extract the resource-specific data
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(paramsJSON), &params); err != nil {
		return nil, fmt.Errorf("failed to parse params JSON: %w", err)
	}

	// Extract the resource-specific data (e.g., "switch:2" object)
	resourceData, ok := params[resource.GetID()]
	if !ok {
		return nil, fmt.Errorf("resource %s not found in params", resource.GetID())
	}

	resourceMap, ok := resourceData.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("resource data is not a JSON object")
	}

	// Extract the output value (boolean in the example)
	// This assumes the value we care about is in the "output" field
	output, ok := resourceMap["output"]
	if !ok {
		return nil, fmt.Errorf("output field not found in resource data")
	}

	// Convert to float64 for SensorReading
	var value float64
	switch v := output.(type) {
	case bool:
		if v {
			value = 1.0
		} else {
			value = 0.0
		}
	case float64:
		value = v
	case int:
		value = float64(v)
	default:
		return nil, fmt.Errorf("unsupported output type: %T", output)
	}

	return &api.SensorReading{
		Value:     value,
		Unit:      "",
		Timestamp: timestamp,
		Valid:     true,
	}, nil
}
