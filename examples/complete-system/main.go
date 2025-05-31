package main

import (
	"context"
	"fmt"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
)

// This example demonstrates all the core components working together
// It's a comprehensive example moved from internal/function/example.go

// ExampleFunction is a simple function implementation for demonstration
type ExampleFunction struct {
	name string
}

// Execute implements the Function interface
func (f *ExampleFunction) Execute(ctx context.Context, event *ce.Event) ([]*ce.Event, error) {
	fmt.Printf("🔧 Processing event: ID=%s, Type=%s, Source=%s\n",
		event.ID(), event.Type(), event.Source())

	// Create a response event
	response := ce.NewEvent()
	response.SetID("response-" + event.ID())
	response.SetSource("example-function")
	response.SetType("com.example.response")
	response.SetDataContentType("application/json")
	response.SetData("application/json", map[string]string{
		"message":       fmt.Sprintf("Processed event %s by function %s", event.ID(), f.name),
		"original_type": event.Type(),
		"processed_at":  time.Now().Format(time.RFC3339),
	})

	return []*ce.Event{&response}, nil
}

// ExamplePlugin demonstrates plugin wrapper functionality
type ExamplePlugin struct {
	meta FunctionMeta
	fn   Function
}

func (p *ExamplePlugin) Name() string       { return p.meta.Name }
func (p *ExamplePlugin) Version() string    { return p.meta.Version }
func (p *ExamplePlugin) Type() string       { return p.meta.Type }
func (p *ExamplePlugin) Function() Function { return p.fn }

// SimpleMetricsCollector demonstrates metrics collection
type SimpleMetricsCollector struct{}

func (m *SimpleMetricsCollector) RecordFunctionInvocation(functionName string, duration time.Duration, status string) {
	fmt.Printf("📊 METRIC: Function %s executed in %v with status %s\n", functionName, duration, status)
}

func (m *SimpleMetricsCollector) RecordFunctionError(functionName string, errorType string) {
	fmt.Printf("📊 METRIC: Function %s error: %s\n", functionName, errorType)
}

func (m *SimpleMetricsCollector) RecordFunctionMemoryUsage(functionName string, memoryBytes int64) {
	fmt.Printf("📊 METRIC: Function %s memory usage: %d bytes\n", functionName, memoryBytes)
}

// SimpleLogger demonstrates structured logging
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string, fields ...Field) {
	fmt.Printf("📝 INFO: %s", msg)
	for _, field := range fields {
		fmt.Printf(" %s=%v", field.Key, field.Value)
	}
	fmt.Println()
}

func (l *SimpleLogger) Error(msg string, fields ...Field) {
	fmt.Printf("📝 ERROR: %s", msg)
	for _, field := range fields {
		fmt.Printf(" %s=%v", field.Key, field.Value)
	}
	fmt.Println()
}

func (l *SimpleLogger) WithFields(fields ...Field) Logger {
	return l // Simplified for demo
}

// MemoryRegistry demonstrates in-memory function storage
type MemoryRegistry struct {
	functions map[string]registryEntry
}

type registryEntry struct {
	meta   FunctionMeta
	binary []byte
}

func (r *MemoryRegistry) StoreFunction(meta FunctionMeta, binary []byte) error {
	if r.functions == nil {
		r.functions = make(map[string]registryEntry)
	}
	r.functions[meta.Name] = registryEntry{meta: meta, binary: binary}
	fmt.Printf("💾 Stored function: %s v%s (%d bytes)\n", meta.Name, meta.Version, len(binary))
	return nil
}

func (r *MemoryRegistry) GetFunction(name string) (FunctionMeta, []byte, error) {
	entry, exists := r.functions[name]
	if !exists {
		return FunctionMeta{}, nil, fmt.Errorf("function %s not found", name)
	}
	fmt.Printf("💾 Retrieved function: %s v%s\n", entry.meta.Name, entry.meta.Version)
	return entry.meta, entry.binary, nil
}

func (r *MemoryRegistry) ListFunctions() ([]FunctionMeta, error) {
	functions := make([]FunctionMeta, 0, len(r.functions))
	for _, entry := range r.functions {
		functions = append(functions, entry.meta)
	}
	fmt.Printf("💾 Listed %d functions from registry\n", len(functions))
	return functions, nil
}

func (r *MemoryRegistry) DeleteFunction(name string) error {
	delete(r.functions, name)
	fmt.Printf("💾 Deleted function: %s\n", name)
	return nil
}

// Required interfaces for this example
type Function interface {
	Execute(ctx context.Context, event *ce.Event) ([]*ce.Event, error)
}

type Plugin interface {
	Name() string
	Version() string
	Type() string
	Function() Function
}

type FunctionMeta struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Version string            `json:"version"`
	Config  map[string]string `json:"config,omitempty"`
}

type Registry interface {
	StoreFunction(meta FunctionMeta, binary []byte) error
	GetFunction(name string) (FunctionMeta, []byte, error)
	ListFunctions() ([]FunctionMeta, error)
	DeleteFunction(name string) error
}

type MetricsCollector interface {
	RecordFunctionInvocation(functionName string, duration time.Duration, status string)
	RecordFunctionError(functionName string, errorType string)
	RecordFunctionMemoryUsage(functionName string, memoryBytes int64)
}

type Logger interface {
	Info(msg string, fields ...Field)
	Error(msg string, fields ...Field)
	WithFields(fields ...Field) Logger
}

type Field struct {
	Key   string
	Value interface{}
}

// Demonstration functions
func main() {
	fmt.Println("🚀 === Complete Function System Example ===")

	// 1. Demonstrate individual components
	demonstrateComponents()

	// 2. Demonstrate complete workflow
	demonstrateWorkflow()
}

func demonstrateComponents() {
	fmt.Println("\n📦 === Component Demonstrations ===")

	// 1. Function execution
	fmt.Println("\n1. Function Execution:")
	function := &ExampleFunction{name: "demo-function"}

	event := ce.NewEvent()
	event.SetID("demo-001")
	event.SetSource("https://example.com/demo")
	event.SetType("com.example.demo")
	event.SetDataContentType("application/json")
	event.SetData("application/json", map[string]string{
		"action": "process",
		"data":   "example data",
	})

	results, err := function.Execute(context.Background(), &event)
	if err != nil {
		fmt.Printf("❌ Error: %v\n", err)
	} else {
		fmt.Printf("✅ Function returned %d events\n", len(results))
	}

	// 2. Plugin demonstration
	fmt.Println("\n2. Plugin System:")
	plugin := &ExamplePlugin{
		meta: FunctionMeta{
			Name:    "demo-plugin",
			Type:    "builtin",
			Version: "1.0.0",
			Config: map[string]string{
				"description": "Demo plugin for system showcase",
				"timeout":     "30s",
			},
		},
		fn: function,
	}

	fmt.Printf("Plugin: %s v%s (%s)\n", plugin.Name(), plugin.Version(), plugin.Type())

	// 3. Registry operations
	fmt.Println("\n3. Registry Operations:")
	registry := &MemoryRegistry{}

	// Store some functions
	functions := []FunctionMeta{
		{Name: "func1", Type: "builtin", Version: "1.0.0"},
		{Name: "func2", Type: "builtin", Version: "2.1.0"},
		{Name: "func3", Type: "external", Version: "1.5.0"},
	}

	for _, meta := range functions {
		registry.StoreFunction(meta, []byte(fmt.Sprintf("binary-for-%s", meta.Name)))
	}

	// List and retrieve
	allFuncs, _ := registry.ListFunctions()
	fmt.Printf("Registry contains %d functions\n", len(allFuncs))

	// 4. Metrics collection
	fmt.Println("\n4. Metrics Collection:")
	metrics := &SimpleMetricsCollector{}
	metrics.RecordFunctionInvocation("demo-function", 15*time.Millisecond, "success")
	metrics.RecordFunctionMemoryUsage("demo-function", 2048)

	// 5. Logging
	fmt.Println("\n5. Structured Logging:")
	logger := &SimpleLogger{}
	logger.Info("System demonstration started",
		Field{Key: "version", Value: "1.0.0"},
		Field{Key: "mode", Value: "demo"})
	logger.Error("Simulated error for demo",
		Field{Key: "error_code", Value: 404},
		Field{Key: "component", Value: "demo"})
}

func demonstrateWorkflow() {
	fmt.Println("\n🔄 === Complete Workflow Demonstration ===")

	// Setup all components
	logger := &SimpleLogger{}
	metrics := &SimpleMetricsCollector{}
	registry := &MemoryRegistry{}

	logger.Info("Setting up function system components")

	// Create and register a function
	function := &ExampleFunction{name: "workflow-demo"}
	meta := FunctionMeta{
		Name:    "workflow-demo",
		Type:    "builtin",
		Version: "1.0.0",
		Config: map[string]string{
			"description": "Workflow demonstration function",
			"timeout":     "30s",
			"memory":      "256MB",
		},
	}

	registry.StoreFunction(meta, []byte("workflow-demo-binary"))

	// Simulate multiple events processing
	testEvents := []struct {
		id        string
		source    string
		eventType string
		data      map[string]interface{}
	}{
		{
			id:        "workflow-001",
			source:    "https://example.com/orders",
			eventType: "com.example.order.created",
			data: map[string]interface{}{
				"order_id": "ORD-12345",
				"customer": "Alice Johnson",
				"amount":   99.99,
			},
		},
		{
			id:        "workflow-002",
			source:    "https://example.com/inventory",
			eventType: "com.example.inventory.updated",
			data: map[string]interface{}{
				"product_id": "PROD-789",
				"quantity":   50,
				"location":   "warehouse-1",
			},
		},
		{
			id:        "workflow-003",
			source:    "https://example.com/users",
			eventType: "com.example.user.login",
			data: map[string]interface{}{
				"user_id":    "user-456",
				"ip_address": "192.168.1.100",
				"timestamp":  time.Now().Unix(),
			},
		},
	}

	logger.Info("Processing workflow events", Field{Key: "event_count", Value: len(testEvents)})

	for i, testEvent := range testEvents {
		startTime := time.Now()

		// Create CloudEvent
		event := ce.NewEvent()
		event.SetID(testEvent.id)
		event.SetSource(testEvent.source)
		event.SetType(testEvent.eventType)
		event.SetDataContentType("application/json")
		event.SetData("application/json", testEvent.data)

		// Process event
		logger.Info("Processing event",
			Field{Key: "event_id", Value: event.ID()},
			Field{Key: "event_type", Value: event.Type()})

		results, err := function.Execute(context.Background(), &event)
		duration := time.Since(startTime)

		if err != nil {
			logger.Error("Function execution failed",
				Field{Key: "event_id", Value: event.ID()},
				Field{Key: "error", Value: err.Error()})
			metrics.RecordFunctionError("workflow-demo", "execution_error")
		} else {
			logger.Info("Function execution completed",
				Field{Key: "event_id", Value: event.ID()},
				Field{Key: "response_count", Value: len(results)},
				Field{Key: "duration_ms", Value: duration.Milliseconds()})
			metrics.RecordFunctionInvocation("workflow-demo", duration, "success")
		}

		// Simulate memory usage
		memoryUsage := int64(1024 * (i + 1) * 2) // Simulate increasing memory usage
		metrics.RecordFunctionMemoryUsage("workflow-demo", memoryUsage)

		// Small delay to make the demo more realistic
		time.Sleep(50 * time.Millisecond)
	}

	// Final statistics
	logger.Info("Workflow demonstration completed",
		Field{Key: "total_events", Value: len(testEvents)},
		Field{Key: "status", Value: "success"})

	fmt.Println("\n✅ Complete system demonstration finished!")
	fmt.Println("   This example shows all core components working together:")
	fmt.Println("   • Function execution with CloudEvents")
	fmt.Println("   • Plugin system integration")
	fmt.Println("   • Registry operations (store/retrieve)")
	fmt.Println("   • Metrics collection and reporting")
	fmt.Println("   • Structured logging throughout")
}
