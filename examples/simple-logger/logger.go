package main

import (
	"fmt"
	"mycelium/internal/function"
)

// DemoLogger shows how to implement and use the Logger interface
func main() {
	fmt.Println("=== Simple Logger Example ===")

	// Create a simple logger instance
	logger := &function.SimpleLogger{}

	// Log info messages with fields
	logger.Info("Application started",
		function.Field{Key: "version", Value: "1.0.0"},
		function.Field{Key: "env", Value: "development"})

	logger.Info("Processing request",
		function.Field{Key: "method", Value: "POST"},
		function.Field{Key: "path", Value: "/api/functions"},
		function.Field{Key: "user_id", Value: 12345})

	// Log error messages
	logger.Error("Database connection failed",
		function.Field{Key: "host", Value: "localhost:5432"},
		function.Field{Key: "database", Value: "functions"},
		function.Field{Key: "timeout", Value: "30s"})

	// Use WithFields to create a logger with context
	funcLogger := logger.WithFields(
		function.Field{Key: "component", Value: "function-executor"},
		function.Field{Key: "function_name", Value: "data-processor"})

	funcLogger.Info("Function execution started")
	funcLogger.Info("Processing 1000 records")
	funcLogger.Error("Failed to process record",
		function.Field{Key: "record_id", Value: "rec_123"},
		function.Field{Key: "error", Value: "validation failed"})

	fmt.Println("\n=== Custom Logger Implementation ===")

	// Example of a custom logger that could implement the interface
	customLogger := &CustomLogger{prefix: "[CUSTOM]"}

	customLogger.Info("This is a custom logger",
		function.Field{Key: "feature", Value: "custom logging"})
	customLogger.Error("Custom error message",
		function.Field{Key: "code", Value: 500})
}

// CustomLogger demonstrates a custom implementation of the Logger interface
type CustomLogger struct {
	prefix string
}

func (l *CustomLogger) Info(msg string, fields ...function.Field) {
	fmt.Printf("%s INFO: %s", l.prefix, msg)
	for _, field := range fields {
		fmt.Printf(" [%s=%v]", field.Key, field.Value)
	}
	fmt.Println()
}

func (l *CustomLogger) Error(msg string, fields ...function.Field) {
	fmt.Printf("%s ERROR: %s", l.prefix, msg)
	for _, field := range fields {
		fmt.Printf(" [%s=%v]", field.Key, field.Value)
	}
	fmt.Println()
}

func (l *CustomLogger) WithFields(fields ...function.Field) function.Logger {
	// For this simple example, just return self
	// In a real implementation, you might create a new logger with the fields
	return l
}
