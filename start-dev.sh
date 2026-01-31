#!/bin/bash

# Life Support System - Development Startup Script

echo "🐠 Starting Life Support System..."
echo ""

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker first."
    exit 1
fi

# Start databases
echo "📊 Starting databases (PostgreSQL & ClickHouse)..."
docker-compose up -d postgres clickhouse

# Wait for databases to be healthy
echo "⏳ Waiting for databases to be ready..."
sleep 5

# Check database health
until docker-compose exec -T postgres pg_isready -U lifesupport > /dev/null 2>&1; do
    echo "   Waiting for PostgreSQL..."
    sleep 2
done
echo "✅ PostgreSQL is ready"

until docker-compose exec -T clickhouse wget --spider -q localhost:8123/ping > /dev/null 2>&1; do
    echo "   Waiting for ClickHouse..."
    sleep 2
done
echo "✅ ClickHouse is ready"

echo ""
echo "🎉 Databases are ready!"
echo ""
echo "To start the backend:"
echo "  cd backend && go run cmd/server/main.go"
echo ""
echo "To start the frontend:"
echo "  cd frontend && npm run dev"
echo ""
echo "Or start everything with Docker:"
echo "  docker-compose up -d"
echo ""
