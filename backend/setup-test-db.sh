#!/bin/bash
# Setup script for running tests

# Create test database if it doesn't exist
export PGPASSWORD=postgres
psql -h localhost -U postgres -tc "SELECT 1 FROM pg_database WHERE datname = 'lifesupport_test'" | grep -q 1 || \
  psql -h localhost -U postgres -c "CREATE DATABASE lifesupport_test"

echo "Test database ready"
