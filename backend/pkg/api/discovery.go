package api

import "time"

// DiscoveryOptions configures device discovery behavior
type DiscoveryOptions struct {
	// Add any options needed for device discovery, e.g. timeouts, concurrency limits, etc.
}

type StatusOptions struct {
	NewerThan *time.Time // Only return status if it's newer than this timestamp
}

// DiscoveryResult contains the results of device discovery
type DiscoveryResult struct {
	DiscoveredTags []string `json:"discovered_tags"`
}
