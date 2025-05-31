package function

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestExampleFunction tests the basic function implementation
func TestExampleFunction(t *testing.T) {
	function := &ExampleFunction{name: "test-function"}

	// Create a mock CloudEvent
	event := ce.NewEvent()
	event.SetID("test-123")
	event.SetSource("test-source")
	event.SetType("com.example.test")
	event.SetDataContentType("application/json")
	event.SetData("application/json", map[string]string{
		"message": "hello world",
		"user":    "testuser",
	})

	// Execute the function
	ctx := context.Background()
	results, err := function.Execute(ctx, &event)

	// Verify results
	require.NoError(t, err)
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "response-test-123", result.ID())
	assert.Equal(t, "example-function", result.Source())
	assert.Equal(t, "com.example.response", result.Type())
	assert.Equal(t, "application/json", result.DataContentType())

	// Verify response data
	var responseData map[string]string
	err = result.DataAs(&responseData)
	require.NoError(t, err)
	assert.Contains(t, responseData["message"], "test-function")
	assert.Equal(t, "com.example.test", responseData["original_type"])
}

// TestExamplePlugin tests the plugin wrapper
func TestExamplePlugin(t *testing.T) {
	meta := FunctionMeta{
		Name:    "test-plugin",
		Type:    "builtin",
		Version: "1.0.0",
	}

	function := &ExampleFunction{name: meta.Name}
	plugin := &ExamplePlugin{
		meta: meta,
		fn:   function,
	}

	// Test plugin metadata
	assert.Equal(t, "test-plugin", plugin.Name())
	assert.Equal(t, "builtin", plugin.Type())
	assert.Equal(t, "1.0.0", plugin.Version())
	assert.Equal(t, function, plugin.Function())
}

// TestMemoryRegistry tests the in-memory registry implementation
func TestMemoryRegistry(t *testing.T) {
	registry := &MemoryRegistry{
		functions: make(map[string]registryEntry),
	}

	meta := FunctionMeta{
		Name:    "test-function",
		Type:    "builtin",
		Version: "1.0.0",
		Config:  map[string]string{"env": "test"},
	}
	binary := []byte("mock binary data")

	// Test storing function
	err := registry.StoreFunction(meta, binary)
	require.NoError(t, err)

	// Test retrieving function
	retrievedMeta, retrievedBinary, err := registry.GetFunction("test-function")
	require.NoError(t, err)
	assert.Equal(t, meta, retrievedMeta)
	assert.Equal(t, binary, retrievedBinary)

	// Test listing functions
	functions, err := registry.ListFunctions()
	require.NoError(t, err)
	require.Len(t, functions, 1)
	assert.Equal(t, meta, functions[0])

	// Test function not found
	_, _, err = registry.GetFunction("nonexistent")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test deleting function
	err = registry.DeleteFunction("test-function")
	require.NoError(t, err)

	// Verify deletion
	functions, err = registry.ListFunctions()
	require.NoError(t, err)
	assert.Len(t, functions, 0)
}

// TestSimpleMetricsCollector tests the metrics collection
func TestSimpleMetricsCollector(t *testing.T) {
	metrics := &SimpleMetricsCollector{}

	// These methods should not panic
	metrics.RecordFunctionInvocation("test-function", time.Millisecond*100, "success")
	metrics.RecordFunctionError("test-function", "execution_error")
	metrics.RecordFunctionMemoryUsage("test-function", 1024*1024) // 1MB
}

// TestSimpleLogger tests the logging implementation
func TestSimpleLogger(t *testing.T) {
	logger := &SimpleLogger{}

	// These methods should not panic
	logger.Info("Test info message", Field{Key: "function", Value: "test"})
	logger.Error("Test error message", Field{Key: "error", Value: "test error"})

	// Test WithFields
	fieldsLogger := logger.WithFields(Field{Key: "service", Value: "function-runtime"})
	assert.NotNil(t, fieldsLogger)
}

// TestRuntimeServiceLoadPlugin tests the plugin loading functionality
func TestRuntimeServiceLoadPlugin(t *testing.T) {
	cfg := RuntimeServiceConfig{
		NATSURL:     "nats://localhost:4222",
		ServiceName: "test-function-runtime",
		Version:     "1.0.0",
		Description: "Test function runtime service",
		Registry: &MemoryRegistry{
			functions: make(map[string]registryEntry),
		},
		Metrics: &SimpleMetricsCollector{},
		Logger:  &SimpleLogger{},
	}

	// Mock NATS connection (this would fail in real test without NATS server)
	// For unit testing, we'll test the loadPlugin method directly
	rs := &RuntimeService{
		registry: cfg.Registry,
		plugins:  make(map[string]Plugin),
		metrics:  cfg.Metrics,
		logger:   cfg.Logger,
	}

	// Test loading built-in function
	meta := FunctionMeta{
		Name:    "example",
		Type:    "builtin",
		Version: "1.0.0",
	}

	plugin, err := rs.loadPlugin(meta, []byte{})
	require.NoError(t, err)
	assert.Equal(t, "example", plugin.Name())
	assert.Equal(t, "builtin", plugin.Type())
	assert.Equal(t, "1.0.0", plugin.Version())

	// Test loading unknown built-in function
	unknownMeta := FunctionMeta{
		Name:    "unknown",
		Type:    "builtin",
		Version: "1.0.0",
	}

	_, err = rs.loadPlugin(unknownMeta, []byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not found")

	// Test unsupported plugin type
	unsupportedMeta := FunctionMeta{
		Name:    "test",
		Type:    "unsupported",
		Version: "1.0.0",
	}

	_, err = rs.loadPlugin(unsupportedMeta, []byte{})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported plugin type")
}

// MockNATSTest tests with embedded NATS server (requires NATS server running)
func TestIntegrationWithNATS(t *testing.T) {
	// Skip this test if NATS is not available
	nc, err := nats.Connect("nats://localhost:4222")
	if err != nil {
		t.Skip("NATS server not available, skipping integration test")
		return
	}
	defer nc.Close()

	// Test with example runtime service
	service, err := CreateExampleRuntimeService("nats://localhost:4222")
	require.NoError(t, err)

	// Start the service
	err = service.Start()
	require.NoError(t, err)
	defer service.Stop()

	// Wait a bit for the service to start
	time.Sleep(100 * time.Millisecond)

	// Create a client
	clientCfg := ClientConfig{
		NATSURL: "nats://localhost:4222",
		Registry: &MemoryRegistry{
			functions: map[string]registryEntry{
				"example": {
					meta: FunctionMeta{
						Name:    "example",
						Type:    "builtin",
						Version: "1.0.0",
					},
					binary: []byte{},
				},
			},
		},
		Timeout: 5 * time.Second,
	}

	client, err := NewClient(clientCfg)
	require.NoError(t, err)
	defer client.Close()

	// Create a test event
	event := ce.NewEvent()
	event.SetID("integration-test-123")
	event.SetSource("integration-test")
	event.SetType("com.example.integration")
	event.SetDataContentType("application/json")
	event.SetData("application/json", map[string]string{
		"test": "integration",
	})

	// Invoke the function
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	results, err := client.InvokeFunction(ctx, "example", &event)
	require.NoError(t, err)
	require.Len(t, results, 1)

	result := results[0]
	assert.Equal(t, "response-integration-test-123", result.ID())
	assert.Equal(t, "example-function", result.Source())
	assert.Equal(t, "com.example.response", result.Type())
}

// TestFunctionInvocationFlow tests the complete function invocation flow
func TestFunctionInvocationFlow(t *testing.T) {
	// Create mock CloudEvent
	event := ce.NewEvent()
	event.SetID("flow-test-456")
	event.SetSource("flow-test")
	event.SetType("com.example.flow")
	event.SetDataContentType("application/json")
	event.SetData("application/json", map[string]interface{}{
		"action": "process",
		"data":   "test data",
	})

	// Test the request structure that would be sent over NATS
	req := struct {
		FunctionName string    `json:"functionName"`
		Event        *ce.Event `json:"event"`
	}{
		FunctionName: "example",
		Event:        &event,
	}

	// Marshal to JSON (simulating NATS message)
	reqData, err := json.Marshal(req)
	require.NoError(t, err)

	// Unmarshal back (simulating receiving NATS message)
	var receivedReq struct {
		FunctionName string    `json:"functionName"`
		Event        *ce.Event `json:"event"`
	}

	err = json.Unmarshal(reqData, &receivedReq)
	require.NoError(t, err)

	assert.Equal(t, "example", receivedReq.FunctionName)
	assert.Equal(t, event.ID(), receivedReq.Event.ID())
	assert.Equal(t, event.Source(), receivedReq.Event.Source())
	assert.Equal(t, event.Type(), receivedReq.Event.Type())

	// Execute function
	function := &ExampleFunction{name: "example"}
	results, err := function.Execute(context.Background(), receivedReq.Event)
	require.NoError(t, err)

	// Test response structure
	response := struct {
		Events []*ce.Event `json:"events"`
	}{
		Events: results,
	}

	responseData, err := json.Marshal(response)
	require.NoError(t, err)

	// Verify response can be unmarshaled
	var receivedResp struct {
		Events []*ce.Event `json:"events"`
	}

	err = json.Unmarshal(responseData, &receivedResp)
	require.NoError(t, err)
	require.Len(t, receivedResp.Events, 1)

	result := receivedResp.Events[0]
	assert.Equal(t, "response-flow-test-456", result.ID())
}

// TestErrorHandling tests error scenarios
func TestErrorHandling(t *testing.T) {
	// Test function that returns an error
	errorFunc := &ErrorFunction{}

	event := ce.NewEvent()
	event.SetID("error-test")
	event.SetSource("test")
	event.SetType("com.example.error")

	_, err := errorFunc.Execute(context.Background(), &event)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "simulated error")

	// Test error response structure
	errorResponse := struct {
		Error     string `json:"error"`
		ErrorType string `json:"errorType"`
	}{
		Error:     err.Error(),
		ErrorType: "execution_error",
	}

	errorData, err := json.Marshal(errorResponse)
	require.NoError(t, err)

	var receivedError struct {
		Error     string `json:"error"`
		ErrorType string `json:"errorType"`
	}

	err = json.Unmarshal(errorData, &receivedError)
	require.NoError(t, err)
	assert.Contains(t, receivedError.Error, "simulated error")
	assert.Equal(t, "execution_error", receivedError.ErrorType)
}

// TestConcurrentFunctionExecution tests concurrent function calls
func TestConcurrentFunctionExecution(t *testing.T) {
	function := &ExampleFunction{name: "concurrent-test"}

	// Create multiple events
	numEvents := 10
	events := make([]*ce.Event, numEvents)
	for i := 0; i < numEvents; i++ {
		event := ce.NewEvent()
		event.SetID(fmt.Sprintf("concurrent-%d", i))
		event.SetSource("concurrent-test")
		event.SetType("com.example.concurrent")
		events[i] = &event
	}

	// Execute functions concurrently
	results := make(chan []*ce.Event, numEvents)
	errors := make(chan error, numEvents)

	for i := 0; i < numEvents; i++ {
		go func(event *ce.Event) {
			result, err := function.Execute(context.Background(), event)
			if err != nil {
				errors <- err
			} else {
				results <- result
			}
		}(events[i])
	}

	// Collect results
	successCount := 0
	errorCount := 0
	timeout := time.After(5 * time.Second)

	for i := 0; i < numEvents; i++ {
		select {
		case <-results:
			successCount++
		case <-errors:
			errorCount++
		case <-timeout:
			t.Fatal("Test timed out")
		}
	}

	assert.Equal(t, numEvents, successCount)
	assert.Equal(t, 0, errorCount)
}

// ErrorFunction is a test function that always returns an error
type ErrorFunction struct{}

func (f *ErrorFunction) Execute(ctx context.Context, event *ce.Event) ([]*ce.Event, error) {
	return nil, fmt.Errorf("simulated error for testing")
}

// BenchmarkFunctionExecution benchmarks function execution performance
func BenchmarkFunctionExecution(b *testing.B) {
	function := &ExampleFunction{name: "benchmark"}

	event := ce.NewEvent()
	event.SetID("benchmark-test")
	event.SetSource("benchmark")
	event.SetType("com.example.benchmark")
	event.SetDataContentType("application/json")
	event.SetData("application/json", map[string]string{"test": "data"})

	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := function.Execute(ctx, &event)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestRuntimeService(t *testing.T) {
	registry := &MemoryRegistry{}
	metrics := &SimpleMetricsCollector{}
	logger := &SimpleLogger{}

	config := RuntimeServiceConfig{
		NATSURL:     nats.DefaultURL,
		ServiceName: "test-function-runtime",
		Version:     "1.0.0",
		Description: "Test function runtime service",
		Registry:    registry,
		Metrics:     metrics,
		Logger:      logger,
	}

	service, err := NewRuntimeService(config)
	if err != nil {
		t.Fatalf("Failed to create runtime service: %v", err)
	}
	defer service.Stop()

	if err := service.Start(); err != nil {
		t.Fatalf("Failed to start service: %v", err)
	}
}
