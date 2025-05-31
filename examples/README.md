# Function System Examples

This directory contains practical examples demonstrating how to use the Mycelium function system with **NATS Service API**. Each example focuses on a specific aspect of the system and can be run independently.

## What's New: NATS Service API Integration

The function system now uses the **NATS Service API** (micro package) for client-service communication, providing:

- **Service Discovery**: Automatic discovery of available services using `$SRV.PING`
- **Service Information**: Detailed metadata and endpoint information via `$SRV.INFO`
- **Statistics & Monitoring**: Real-time service statistics through `$SRV.STATS` 
- **Health Checks**: Built-in service health monitoring and status reporting
- **Load Balancing**: Automatic load balancing across service instances
- **Versioning**: Service versioning and compatibility checking

## Prerequisites

- Go 1.21 or later
- NATS server (for runtime service examples)
- CloudEvents Go SDK

## Quick Start

### Install Dependencies

```bash
# From the root of the mycelium project
go mod tidy
```

### Start NATS Server (for runtime examples)

```bash
# Install NATS server if not already installed
go install github.com/nats-io/nats-server/v2@latest

# Start NATS server
nats-server
```

## Examples Overview

### 1. Basic Function (`basic/`)

**Purpose**: Demonstrates how to implement a simple function that processes CloudEvents.

**What it shows**:
- Basic function interface implementation
- CloudEvent handling and transformation
- Creating response events
- Error handling

**Run it**:
```bash
cd examples/basic
go run function.go
```

**Expected Output**:
```
Processing event: ID=example-001, Type=com.example.data, Source=https://example.com/source
Function returned 1 events:
Event 1: ID=response-example-001, Type=com.example.response, Source=basic-function
  Data: map[message:Processed by my-basic-function original_id:example-001 ...]
```

### 2. Simple Logger (`simple-logger/`)

**Purpose**: Shows how to use and implement custom loggers.

**What it shows**:
- Using the built-in SimpleLogger
- Structured logging with fields
- Creating custom logger implementations
- Logger interface compliance

**Run it**:
```bash
cd examples/simple-logger
go run logger.go
```

**Expected Output**:
```
=== Simple Logger Example ===
INFO: Application started version=1.0.0 env=development
INFO: Processing request method=POST path=/api/functions user_id=12345
ERROR: Database connection failed host=localhost:5432 database=functions timeout=30s
...
```

### 3. Memory Registry (`memory-registry/`)

**Purpose**: Demonstrates function metadata storage and retrieval using the in-memory registry.

**What it shows**:
- Function metadata management
- CRUD operations (Create, Read, Update, Delete)
- Registry interface usage
- Function versioning

**Run it**:
```bash
cd examples/memory-registry
go run registry.go
```

**Expected Output**:
```
=== Memory Registry Example ===

Storing functions in registry:
✓ Stored function: echo-function v1.0.0
✓ Stored function: transform-function v2.1.0
✓ Stored function: notify-function v1.5.2
...
```

### 4. Runtime Service with NATS Service API (`runtime-service/`) ⭐

**Purpose**: Shows complete runtime service setup with NATS Service API integration.

**What it shows**:
- **NATS Service API**: Service registration and discovery
- **Service Metadata**: Automatic endpoint documentation
- **Statistics**: Real-time performance monitoring
- **Error Handling**: Structured error responses
- End-to-end function invocation

**Run it**:
```bash
# Make sure NATS server is running first
nats-server &

cd examples/runtime-service
go run main.go
```

**Expected Output**:
```
=== NATS Service API Runtime Service Example ===

1. Setting up runtime service with NATS Service API
✓ Runtime service started successfully

2. Discovering services using NATS Service API
✓ Discovered service: {"name":"example-function-runtime","endpoints":[...],...}
✓ Service info: {"name":"example-function-runtime","endpoints":[...],...}

3. Testing function invocation with NATS Service API
✓ Success: received 1 response events in 1.016334ms

4. Testing service statistics
✓ Service stats: {"endpoints":[{"num_requests":1,"average_processing_time":383166}],...}
```

### 5. NATS Service CLI Tool (`nats-service-cli/`) 🆕

**Purpose**: Interactive CLI tool for discovering and monitoring NATS services.

**What it shows**:
- Service discovery using NATS Service API
- Service information retrieval
- Real-time statistics monitoring
- Service health checking

**Run it**:
```bash
# Discover all services
cd examples/nats-service-cli
go run main.go discover

# Get detailed service information
go run main.go info example-function-runtime

# Get service statistics
go run main.go stats example-function-runtime
```

### 6. Complete System (`complete-system/`) ⭐

**Purpose**: Comprehensive demonstration of all system components working together.

**What it shows**:
- All core interfaces and implementations
- Complete workflow with real-world scenarios
- Component integration patterns
- Metrics collection and logging in action
- Production-like event processing

**Run it**:
```bash
cd examples/complete-system
go run main.go
```

**Expected Output**:
```
🚀 === Complete Function System Example ===

📦 === Component Demonstrations ===

1. Function Execution:
🔧 Processing event: ID=demo-001, Type=com.example.demo, Source=https://example.com/demo
✅ Function returned 1 events

2. Plugin System:
Plugin: demo-plugin v1.0.0 (builtin)

3. Registry Operations:
💾 Stored function: func1 v1.0.0 (14 bytes)
💾 Stored function: func2 v2.1.0 (14 bytes)
💾 Stored function: func3 v1.5.0 (14 bytes)
💾 Listed 3 functions from registry
Registry contains 3 functions

4. Metrics Collection:
📊 METRIC: Function demo-function executed in 15ms with status success
📊 METRIC: Function demo-function memory usage: 2048 bytes

5. Structured Logging:
📝 INFO: System demonstration started version=1.0.0 mode=demo
📝 ERROR: Simulated error for demo error_code=404 component=demo

🔄 === Complete Workflow Demonstration ===
📝 INFO: Setting up function system components
💾 Stored function: workflow-demo v1.0.0 (20 bytes)
📝 INFO: Processing workflow events event_count=3
...
✅ Complete system demonstration finished!
```

## Examples by Complexity

### **Beginner Level**
1. **Basic Function** - Start here to understand the core Function interface
2. **Simple Logger** - Learn about structured logging
3. **Memory Registry** - Understand function metadata management

### **Intermediate Level**
4. **Runtime Service** - See complete NATS Service API integration
5. **NATS Service CLI** - Learn service discovery and monitoring

### **Advanced Level**
6. **Complete System** - Comprehensive workflow with all components integrated

## NATS Service API Features

### Service Discovery

```bash
# Discover all available services
curl -X POST http://localhost:8222/v1/subs/request -d '{"subject":"$SRV.PING"}'

# Or use the CLI tool
go run examples/nats-service-cli/main.go discover
```

### Service Information

```bash
# Get detailed service information
go run examples/nats-service-cli/main.go info example-function-runtime
```

### Service Statistics

```bash
# Get real-time service statistics
go run examples/nats-service-cli/main.go stats example-function-runtime
```

## Common Use Cases

### Creating a Service with NATS Service API

```go
// Create runtime service with NATS Service API
service, err := function.NewRuntimeService(function.RuntimeServiceConfig{
    NATSURL:     "nats://localhost:4222",
    ServiceName: "my-function-service",
    Version:     "1.0.0",
    Description: "My serverless function service",
    Registry:    registry,
    Metrics:     &function.SimpleMetricsCollector{},
    Logger:      &function.SimpleLogger{},
})

// Service automatically registers with NATS and provides:
// - $SRV.PING response for discovery
// - $SRV.INFO.my-function-service for detailed information  
// - $SRV.STATS.my-function-service for statistics
// - function.invoke endpoint for function execution
```

### Discovering Services

```go
// Discover all available services
nc, _ := nats.Connect("nats://localhost:4222")
resp, err := nc.Request("$SRV.PING", nil, 2*time.Second)
// Parse JSON response to get service list
```

### Monitoring Service Health

```go
// Get service statistics
resp, err := nc.Request("$SRV.STATS.my-function-service", nil, 2*time.Second)
// Parse JSON to get request counts, error rates, processing times
```

## Integration with Tests

These examples use the same components as the test suite. You can run the full test suite to verify everything works:

```bash
# From project root
cd internal/function
./run_tests.sh
```

## Next Steps

- Start with **Runtime Service** to see NATS Service API in action
- Try **NATS Service CLI** for interactive service discovery
- Explore **Complete System** for comprehensive overview
- Read the NATS Service API documentation at https://docs.nats.io/using-nats/developer/services

## Contributing

To add new examples:

1. Create a new directory under `examples/`
2. Add a `README.md` explaining the example
3. Include a runnable `main.go` or similar
4. Update this main `README.md` with the new example
5. Test the example with various scenarios 