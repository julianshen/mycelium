#!/bin/bash

echo "Function System Test Runner"
echo "=========================="

# Check if NATS is running
if ! nc -z localhost 4222 2>/dev/null; then
    echo "⚠️  NATS server not running on localhost:4222"
    echo "   Integration tests will be skipped"
    echo "   To run integration tests, start NATS with: nats-server"
    echo ""
fi

echo "Running unit tests..."
go test -v -run "Test.*Function|Test.*Plugin|Test.*Registry|Test.*Metrics|Test.*Logger|Test.*InvocationFlow|Test.*ErrorHandling|Test.*Concurrent|Test.*DataTypes|Test.*Compliance" ./internal/function

echo ""
echo "Running integration tests (requires NATS server)..."
go test -v -run "Test.*Integration|Test.*Workflow|Test.*Clients|Test.*Propagation|Test.*Cancellation|Test.*Throughput|Test.*BasicClientServer" ./internal/function

echo ""
echo "Running benchmarks..."
go test -bench=. -benchmem ./internal/function

echo ""
echo "Test coverage..."
go test -cover ./internal/function 