package main

import (
	"context"
	"fmt"
	"log"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"

	"mycelium/internal/function"
)

func main() {
	fmt.Println("=== NATS Service API Runtime Service Example ===")

	// Connect to NATS
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	fmt.Println("\n1. Setting up runtime service with NATS Service API")

	// Create runtime service with memory registry
	registry := &function.MemoryRegistry{}

	// Store some example functions
	functions := []function.FunctionMeta{
		{Name: "example", Type: "builtin", Version: "1.0.0"},
		{Name: "echo", Type: "builtin", Version: "1.1.0"},
		{Name: "transform", Type: "builtin", Version: "2.0.0"},
	}

	for _, meta := range functions {
		registry.StoreFunction(meta, []byte(fmt.Sprintf("binary-for-%s", meta.Name)))
	}

	// Create and start runtime service
	service, err := function.NewRuntimeService(function.RuntimeServiceConfig{
		NATSURL:     nats.DefaultURL,
		ServiceName: "example-function-runtime",
		Version:     "1.0.0",
		Description: "Example serverless function runtime using NATS Service API",
		Registry:    registry,
		Metrics:     &function.SimpleMetricsCollector{},
		Logger:      &function.SimpleLogger{},
	})
	if err != nil {
		log.Fatalf("Failed to create runtime service: %v", err)
	}

	if err := service.Start(); err != nil {
		log.Fatalf("Failed to start runtime service: %v", err)
	}
	defer service.Stop()

	fmt.Println("✓ Runtime service started successfully")

	// Give the service a moment to fully initialize
	time.Sleep(100 * time.Millisecond)

	fmt.Println("\n2. Discovering services using NATS Service API")

	// Discover services using the service discovery protocol
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// Send a PING to discover available services
	response, err := nc.RequestWithContext(ctx, "$SRV.PING", nil)
	if err != nil {
		log.Printf("Warning: Failed to discover services: %v", err)
	} else {
		fmt.Printf("✓ Discovered service: %s\n", string(response.Data))
	}

	// Get service information
	response, err = nc.RequestWithContext(ctx, "$SRV.INFO.example-function-runtime", nil)
	if err != nil {
		log.Printf("Warning: Failed to get service info: %v", err)
	} else {
		fmt.Printf("✓ Service info: %s\n", string(response.Data))
	}

	fmt.Println("\n3. Testing function invocation with NATS Service API")

	// Create client
	client, err := function.NewClient(function.ClientConfig{
		NATSURL:  nats.DefaultURL,
		Registry: registry,
		Timeout:  5 * time.Second,
	})
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Test events for different scenarios
	testCases := []struct {
		name         string
		functionName string
		eventID      string
		eventType    string
		data         map[string]interface{}
	}{
		{
			name:         "Basic function call",
			functionName: "example",
			eventID:      "test-001",
			eventType:    "com.example.test",
			data: map[string]interface{}{
				"message": "Hello from NATS Service API",
				"version": "1.0.0",
			},
		},
	}

	for i, testCase := range testCases {
		fmt.Printf("\nTest %d: %s\n", i+1, testCase.name)

		// Create CloudEvent
		event := ce.NewEvent()
		event.SetID(testCase.eventID)
		event.SetSource("https://example.com/test")
		event.SetType(testCase.eventType)
		event.SetDataContentType("application/json")
		event.SetData("application/json", testCase.data)

		// Invoke function
		fmt.Printf("  Invoking function '%s' with event ID: %s\n", testCase.functionName, testCase.eventID)

		start := time.Now()
		results, err := client.InvokeFunction(context.Background(), testCase.functionName, &event)
		duration := time.Since(start)

		if err != nil {
			fmt.Printf("  ❌ Error: %v\n", err)
			continue
		}

		fmt.Printf("  ✓ Success: received %d response events in %v\n", len(results), duration)

		// Display response details
		for j, result := range results {
			fmt.Printf("    Response %d: ID=%s, Type=%s, Source=%s\n",
				j+1, result.ID(), result.Type(), result.Source())

			if result.Data() != nil {
				fmt.Printf("      Data: %s\n", string(result.Data()))
			}
		}
	}

	fmt.Println("\n4. Testing service statistics")

	// Get service statistics
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	response, err = nc.RequestWithContext(ctx, "$SRV.STATS.example-function-runtime", nil)
	if err != nil {
		log.Printf("Warning: Failed to get service stats: %v", err)
	} else {
		fmt.Printf("✓ Service stats: %s\n", string(response.Data))
	}

	fmt.Println("\n5. Testing error handling")

	// Test with non-existent function
	event := ce.NewEvent()
	event.SetID("error-test-001")
	event.SetSource("https://example.com/error-test")
	event.SetType("com.example.error")
	event.SetDataContentType("application/json")
	event.SetData("application/json", map[string]string{"test": "error handling"})

	fmt.Println("  Testing non-existent function...")
	_, err = client.InvokeFunction(context.Background(), "non-existent", &event)
	if err != nil {
		fmt.Printf("  ✓ Expected error received: %v\n", err)
	} else {
		fmt.Println("  ❌ Expected an error but didn't get one")
	}

	fmt.Println("\n6. Service discovery and monitoring")

	// List all available services
	ctx, cancel = context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	// The PING command should discover all services
	fmt.Println("  Discovering all available services...")
	response, err = nc.RequestWithContext(ctx, "$SRV.PING", nil)
	if err != nil {
		log.Printf("Warning: Service discovery failed: %v", err)
	} else {
		fmt.Printf("  ✓ Available services discovered\n")
	}

	fmt.Println("\n✅ NATS Service API demonstration completed!")
	fmt.Println("\nKey improvements with NATS Service API:")
	fmt.Println("  • Structured service discovery ($SRV.PING, $SRV.INFO)")
	fmt.Println("  • Built-in service statistics and monitoring ($SRV.STATS)")
	fmt.Println("  • Standardized service metadata and versioning")
	fmt.Println("  • Better error handling and service health checks")
	fmt.Println("  • Simplified client-service communication")

	// Keep the service running for a bit to allow for manual testing
	fmt.Println("\nService will remain running for 5 seconds for manual testing...")
	time.Sleep(5 * time.Second)
}

func init() {
	// Configure logging to be less verbose for the example
	log.SetFlags(log.LstdFlags | log.Lshortfile)
}
