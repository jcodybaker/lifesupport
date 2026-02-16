package httpapi

import (
	"context"
	"testing"

	"lifesupport/backend/pkg/storer"
)

// Test helpers

func setupTestDB(t *testing.T) *storer.Storer {
	t.Helper()

	// Use a test database connection
	connStr := "host=localhost port=5432 user=postgres password=postgres dbname=lifesupport_test sslmode=disable"
	store, err := storer.New(connStr)
	if err != nil {
		t.Skipf("Skipping test: cannot connect to test database: %v", err)
		return nil
	}

	// Initialize schema
	ctx := context.Background()
	if err := store.InitSchema(ctx); err != nil {
		t.Fatalf("Failed to initialize schema: %v", err)
	}

	// Clean up existing test data
	cleanupTestData(t, store)

	return store
}

func cleanupTestData(t *testing.T, store *storer.Storer) {
	t.Helper()
	ctx := context.Background()

	// Delete test devices
	_ = store.DeleteDevice(ctx, "test-dev-001")
	_ = store.DeleteDevice(ctx, "test-dev-002")
}

func teardownTestDB(t *testing.T, store *storer.Storer) {
	t.Helper()
	if store != nil {
		cleanupTestData(t, store)
		store.Close()
	}
}
