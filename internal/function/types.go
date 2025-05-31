package function

import (
	"context"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
)

// FunctionMeta represents the metadata of a function
type FunctionMeta struct {
	Name    string            `json:"name"`
	Type    string            `json:"type"`
	Version string            `json:"version"`
	Config  map[string]string `json:"config,omitempty"`
}

// FunctionResult represents the result returned from a function
type FunctionResult struct {
	Event *ce.Event `json:"event"`
	Error string    `json:"error,omitempty"`
}

// Function represents the interface that all functions must implement
type Function interface {
	// Execute processes the incoming event and returns zero or more events
	Execute(ctx context.Context, event *ce.Event) ([]*ce.Event, error)
}

// Plugin represents a loaded function plugin
type Plugin interface {
	// Name returns the name of the plugin
	Name() string
	// Version returns the version of the plugin
	Version() string
	// Type returns the type of the plugin
	Type() string
	// Function returns the function implementation
	Function() Function
}

// Registry defines the interface for function storage and retrieval
type Registry interface {
	// StoreFunction stores a function's metadata and binary
	StoreFunction(meta FunctionMeta, binary []byte) error
	// GetFunction retrieves a function's metadata and binary
	GetFunction(name string) (FunctionMeta, []byte, error)
	// ListFunctions returns a list of all available functions
	ListFunctions() ([]FunctionMeta, error)
	// DeleteFunction removes a function
	DeleteFunction(name string) error
}

// MetricsCollector defines the interface for collecting metrics
type MetricsCollector interface {
	// RecordFunctionInvocation records a function invocation
	RecordFunctionInvocation(functionName string, duration time.Duration, status string)
	// RecordFunctionError records a function error
	RecordFunctionError(functionName string, errorType string)
	// RecordFunctionMemoryUsage records function memory usage
	RecordFunctionMemoryUsage(functionName string, memoryBytes int64)
}

// Logger defines the interface for logging
type Logger interface {
	// Info logs an info message
	Info(msg string, fields ...Field)
	// Error logs an error message
	Error(msg string, fields ...Field)
	// WithFields returns a new logger with the given fields
	WithFields(fields ...Field) Logger
}

// Field represents a structured logging field
type Field struct {
	Key   string
	Value interface{}
}
