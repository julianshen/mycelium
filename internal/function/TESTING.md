# Function System Testing Guide

This document describes the comprehensive test suite for the function system, including unit tests, integration tests, and performance benchmarks.

## Test Structure

### Unit Tests (`function_test.go`)

**🔬 Core Component Tests**
- `TestExampleFunction` - Tests basic function implementation with mock CloudEvents
- `TestExamplePlugin` - Tests plugin wrapper functionality
- `TestMemoryRegistry` - Tests in-memory registry CRUD operations
- `TestSimpleMetricsCollector` - Tests metrics collection interface
- `TestSimpleLogger` - Tests structured logging interface
- `TestRuntimeServiceLoadPlugin` - Tests plugin loading mechanisms

**🎯 Flow and Data Tests**
- `TestFunctionInvocationFlow` - Tests complete NATS message flow simulation
- `TestErrorHandling` - Tests error propagation and response structures
- `TestConcurrentFunctionExecution` - Tests thread safety with 10 concurrent executions
- `TestEventDataTypes` - Tests different CloudEvent data types (JSON, text, binary)
- `TestCloudEventCompliance` - Tests CloudEvents specification compliance

### Integration Tests (`integration_test.go`)

**🚀 End-to-End Tests** (Require NATS server)
- `TestCompleteWorkflow` - Tests production runtime service with NATS registry
- `TestMultipleClients` - Tests 3 clients with 5 invocations each (15 total)
- `TestErrorPropagation` - Tests error handling through complete system
- `TestContextCancellation` - Tests context timeout behavior
- `TestHighThroughput` - Tests 100 requests with 10 max concurrency

**📊 Performance Benchmarks**
- `BenchmarkFunctionExecution` - Measures function execution performance

## Test Categories

### 🟢 **Unit Tests** (No Dependencies)
These tests run without external dependencies and focus on individual components:

```bash
go test -v -run "Test.*Function|Test.*Plugin|Test.*Registry|Test.*Metrics|Test.*Logger|Test.*InvocationFlow|Test.*ErrorHandling|Test.*Concurrent|Test.*DataTypes|Test.*Compliance" ./internal/function
```

**Expected Results:**
- ✅ All tests should pass
- ⏱️ Run time: < 1 second
- 📊 Coverage: Core interfaces and business logic

### 🔵 **Integration Tests** (Require NATS)
These tests require a running NATS server on `localhost:4222`:

```bash
# Start NATS server first
nats-server

# Run integration tests
go test -v -run "Test.*Integration|Test.*Workflow|Test.*Clients|Test.*Propagation|Test.*Cancellation|Test.*Throughput" ./internal/function
```

**Expected Results:**
- ✅ All tests should pass with NATS running
- ⚠️ Tests are skipped if NATS is unavailable
- ⏱️ Run time: 1-2 seconds
- 📊 Coverage: Full system integration and NATS messaging

## Mock Event Examples

### Basic CloudEvent
```go
event := ce.NewEvent()
event.SetID("test-123")
event.SetSource("test-source")
event.SetType("com.example.test")
event.SetDataContentType("application/json")
event.SetData("application/json", map[string]string{
    "message": "hello world",
    "user":    "testuser",
})
```

### Complex CloudEvent with Extensions
```go
event := ce.NewEvent()
event.SetID("compliance-test-001")
event.SetSource("https://example.com/compliance-test")
event.SetSpecVersion("1.0")
event.SetType("com.example.compliance")
event.SetDataContentType("application/json")
event.SetDataSchema("https://example.com/schema")
event.SetSubject("compliance/testing")
event.SetTime(time.Now())

// Add extensions
event.SetExtension("custom-ext-1", "value1")
event.SetExtension("custom-ext-2", 42)

// Set complex data
event.SetData("application/json", map[string]interface{}{
    "compliance": true,
    "version":    "1.0",
    "metadata": map[string]interface{}{
        "processed": time.Now().Unix(),
        "tags":      []string{"test", "compliance"},
    },
})
```

### Different Data Types
```go
// JSON Object
event.SetData("application/json", map[string]interface{}{
    "name":  "test",
    "value": 42,
    "array": []string{"a", "b", "c"},
})

// Plain Text
event.SetData("text/plain", "simple text content")

// Binary Data
event.SetData("application/octet-stream", []byte{0x48, 0x65, 0x6c, 0x6c, 0x6f})
```

## Test Scenarios

### 🎯 **Function Execution Scenarios**

1. **Single Event Processing**
   - Input: CloudEvent with JSON data
   - Expected: Single response event with transformed data
   - Validates: Basic function execution

2. **Multiple Event Types**
   - Input: Various content types (JSON, text, binary)
   - Expected: Consistent response format
   - Validates: Data type handling

3. **Error Conditions**
   - Input: Events that trigger errors
   - Expected: Proper error responses
   - Validates: Error handling and propagation

### 🔄 **Concurrency Scenarios**

1. **Concurrent Function Execution**
   - Input: 10 simultaneous function calls
   - Expected: All succeed with correct responses
   - Validates: Thread safety

2. **Multiple Client Load**
   - Input: 3 clients, 5 calls each (15 total)
   - Expected: All succeed with proper event correlation
   - Validates: NATS queue handling

3. **High Throughput**
   - Input: 100 requests with controlled concurrency
   - Expected: >80% success rate, >10 req/sec throughput
   - Validates: Performance under load

### 🌐 **Integration Scenarios**

1. **End-to-End Workflow**
   - Input: Complete NATS request/reply flow
   - Expected: Proper event processing and response
   - Validates: Full system integration

2. **Error Propagation**
   - Input: Requests for non-existent functions
   - Expected: Proper error responses via NATS
   - Validates: Error handling through messaging layer

3. **Context Management**
   - Input: Requests with tight timeout contexts
   - Expected: Proper timeout handling
   - Validates: Context cancellation support

## Running Tests

### Quick Test (Unit Only)
```bash
go test ./internal/function
```

### Complete Test Suite
```bash
# Start NATS server
nats-server &

# Run all tests
go test -v ./internal/function

# Kill NATS server
killall nats-server
```

### Using Test Runner Script
```bash
./internal/function/run_tests.sh
```

### Performance Testing
```bash
# Run benchmarks
go test -bench=. -benchmem ./internal/function

# With CPU profiling
go test -bench=. -cpuprofile=cpu.prof ./internal/function

# With memory profiling
go test -bench=. -memprofile=mem.prof ./internal/function
```

### Coverage Analysis
```bash
# Generate coverage report
go test -cover ./internal/function

# Detailed coverage HTML report
go test -coverprofile=coverage.out ./internal/function
go tool cover -html=coverage.out -o coverage.html
open coverage.html
```

## Expected Test Results

### Successful Unit Test Run
```
=== RUN   TestExampleFunction
--- PASS: TestExampleFunction (0.00s)
=== RUN   TestExamplePlugin
--- PASS: TestExamplePlugin (0.00s)
=== RUN   TestMemoryRegistry
--- PASS: TestMemoryRegistry (0.00s)
...
PASS
ok      mycelium/internal/function      0.296s
```

### Successful Integration Test Run (with NATS)
```
=== RUN   TestCompleteWorkflow
INFO: Runtime service started queueGroup=function-runtime
METRIC: Function example executed in 12.5µs with status success
INFO: Runtime service stopped
--- PASS: TestCompleteWorkflow (0.23s)
...
PASS
ok      mycelium/internal/function      1.157s
```

### Performance Expectations
- **Function Execution**: < 50µs per call
- **End-to-End Latency**: < 1ms per request
- **Throughput**: > 1000 requests/second
- **Memory Usage**: < 1MB per function instance
- **Success Rate**: > 99% under normal load

## Troubleshooting

### Common Issues

**NATS Connection Failed**
```
t.Skip("NATS server not available, skipping integration test")
```
**Solution**: Start NATS server with `nats-server`

**Context Timeout in Tests**
```
Test timed out waiting for results
```
**Solution**: Check NATS connectivity or increase test timeouts

**Function Not Found Errors**
```
built-in function error-function not found
```
**Solution**: This is expected for error testing scenarios

### Debug Mode

Enable verbose logging in tests:
```go
// In test setup
logger := &SimpleLogger{} // Already verbose by default
```

### Test Data Cleanup

Tests automatically clean up:
- NATS connections are closed
- Runtime services are stopped
- Temporary data is cleared

No manual cleanup required. 