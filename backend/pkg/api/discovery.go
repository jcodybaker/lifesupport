package api

// DiscoveryOptions configures device discovery behavior
type DiscoveryOptions struct {
	// Add any options needed for device discovery, e.g. timeouts, concurrency limits, etc.
}

// DiscoveryResult contains the results of device discovery
type DiscoveryResult struct {
	DiscoveredTags []string `json:"discovered_tags"`
}
