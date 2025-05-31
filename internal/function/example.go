package function

import (
	"context"
	"fmt"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
)

// These are minimal implementations needed for the test suite.
// For detailed examples and comprehensive demonstrations, see the examples/ directory.

// ExampleFunction is a minimal function implementation for testing
type ExampleFunction struct {
	name string
}

// Execute implements the Function interface
func (f *ExampleFunction) Execute(ctx context.Context, event *ce.Event) ([]*ce.Event, error) {
	response := ce.NewEvent()
	response.SetID("response-" + event.ID())
	response.SetSource("example-function")
	response.SetType("com.example.response")
	response.SetDataContentType("application/json")
	response.SetData("application/json", map[string]string{
		"message":       fmt.Sprintf("Processed event %s by function %s", event.ID(), f.name),
		"original_type": event.Type(),
	})
	return []*ce.Event{&response}, nil
}

// ExamplePlugin is a minimal plugin implementation for testing
type ExamplePlugin struct {
	meta FunctionMeta
	fn   Function
}

func (p *ExamplePlugin) Name() string       { return p.meta.Name }
func (p *ExamplePlugin) Version() string    { return p.meta.Version }
func (p *ExamplePlugin) Type() string       { return p.meta.Type }
func (p *ExamplePlugin) Function() Function { return p.fn }

// SimpleMetricsCollector is a minimal metrics collector for testing
type SimpleMetricsCollector struct{}

func (m *SimpleMetricsCollector) RecordFunctionInvocation(functionName string, duration time.Duration, status string) {
	fmt.Printf("METRIC: Function %s executed in %v with status %s\n", functionName, duration, status)
}

func (m *SimpleMetricsCollector) RecordFunctionError(functionName string, errorType string) {
	fmt.Printf("METRIC: Function %s error: %s\n", functionName, errorType)
}

func (m *SimpleMetricsCollector) RecordFunctionMemoryUsage(functionName string, memoryBytes int64) {
	fmt.Printf("METRIC: Function %s memory usage: %d bytes\n", functionName, memoryBytes)
}

// SimpleLogger is a minimal logger implementation for testing
type SimpleLogger struct{}

func (l *SimpleLogger) Info(msg string, fields ...Field) {
	fmt.Printf("INFO: %s", msg)
	for _, field := range fields {
		fmt.Printf(" %s=%v", field.Key, field.Value)
	}
	fmt.Println()
}

func (l *SimpleLogger) Error(msg string, fields ...Field) {
	fmt.Printf("ERROR: %s", msg)
	for _, field := range fields {
		fmt.Printf(" %s=%v", field.Key, field.Value)
	}
	fmt.Println()
}

func (l *SimpleLogger) WithFields(fields ...Field) Logger {
	return l
}

// registryEntry represents a stored function
type registryEntry struct {
	meta   FunctionMeta
	binary []byte
}

// MemoryRegistry is a minimal in-memory registry implementation for testing
type MemoryRegistry struct {
	functions map[string]registryEntry
}

func (r *MemoryRegistry) StoreFunction(meta FunctionMeta, binary []byte) error {
	if r.functions == nil {
		r.functions = make(map[string]registryEntry)
	}
	r.functions[meta.Name] = registryEntry{meta: meta, binary: binary}
	return nil
}

func (r *MemoryRegistry) GetFunction(name string) (FunctionMeta, []byte, error) {
	entry, exists := r.functions[name]
	if !exists {
		return FunctionMeta{}, nil, fmt.Errorf("function %s not found", name)
	}
	return entry.meta, entry.binary, nil
}

func (r *MemoryRegistry) ListFunctions() ([]FunctionMeta, error) {
	functions := make([]FunctionMeta, 0, len(r.functions))
	for _, entry := range r.functions {
		functions = append(functions, entry.meta)
	}
	return functions, nil
}

func (r *MemoryRegistry) DeleteFunction(name string) error {
	delete(r.functions, name)
	return nil
}

// CreateExampleRuntimeService creates a runtime service for testing.
// For detailed examples, see examples/ directory.
func CreateExampleRuntimeService(natsURL string) (*RuntimeService, error) {
	registry := &MemoryRegistry{functions: make(map[string]registryEntry)}

	// Register minimal example function
	err := registry.StoreFunction(FunctionMeta{
		Name:    "example",
		Type:    "builtin",
		Version: "1.0.0",
	}, []byte{})
	if err != nil {
		return nil, fmt.Errorf("failed to store example function: %w", err)
	}

	return NewRuntimeService(RuntimeServiceConfig{
		NATSURL:     natsURL,
		ServiceName: "function-runtime",
		Version:     "1.0.0",
		Description: "Example function runtime service",
		Registry:    registry,
		Metrics:     &SimpleMetricsCollector{},
		Logger:      &SimpleLogger{},
	})
}

// CreateProductionRuntimeService creates a runtime service with NATS registry for testing.
// For detailed examples, see examples/ directory.
func CreateProductionRuntimeService(natsURL string) (*RuntimeService, error) {
	nc, err := nats.Connect(natsURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	registry, err := NewNATSRegistry(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create NATS registry: %w", err)
	}

	// Pre-register example function
	err = registry.StoreFunction(FunctionMeta{
		Name:    "example",
		Type:    "builtin",
		Version: "1.0.0",
	}, []byte{})
	if err != nil {
		return nil, fmt.Errorf("failed to store example function: %w", err)
	}

	return NewRuntimeService(RuntimeServiceConfig{
		NATSURL:     natsURL,
		ServiceName: "production-function-runtime",
		Version:     "1.0.0",
		Description: "Production function runtime service with NATS registry",
		Registry:    registry,
		Metrics:     &SimpleMetricsCollector{},
		Logger:      &SimpleLogger{},
	})
}
