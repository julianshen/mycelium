package main

import (
	"context"
	"fmt"

	ce "github.com/cloudevents/sdk-go/v2"
)

// BasicFunction demonstrates a simple function implementation
type BasicFunction struct {
	name string
}

// NewBasicFunction creates a new basic function
func NewBasicFunction(name string) *BasicFunction {
	return &BasicFunction{name: name}
}

// Execute implements the Function interface
func (f *BasicFunction) Execute(ctx context.Context, event *ce.Event) ([]*ce.Event, error) {
	// Log the incoming event
	fmt.Printf("Processing event: ID=%s, Type=%s, Source=%s\n",
		event.ID(), event.Type(), event.Source())

	// Extract data from the event
	var eventData map[string]interface{}
	if err := event.DataAs(&eventData); err != nil {
		// Handle events without data gracefully
		eventData = map[string]interface{}{}
	}

	// Create a response event
	response := ce.NewEvent()
	response.SetID("response-" + event.ID())
	response.SetSource("basic-function")
	response.SetType("com.example.response")
	response.SetDataContentType("application/json")

	// Create response data
	responseData := map[string]interface{}{
		"message":         fmt.Sprintf("Processed by %s", f.name),
		"original_id":     event.ID(),
		"original_type":   event.Type(),
		"original_source": event.Source(),
		"processed_data":  eventData,
		"timestamp":       "now",
	}

	response.SetData("application/json", responseData)

	return []*ce.Event{&response}, nil
}

// main demonstrates how to use the basic function
func main() {
	// Create a function instance
	function := NewBasicFunction("my-basic-function")

	// Create a sample CloudEvent
	event := ce.NewEvent()
	event.SetID("example-001")
	event.SetSource("https://example.com/source")
	event.SetType("com.example.data")
	event.SetDataContentType("application/json")
	event.SetData("application/json", map[string]string{
		"user":     "alice",
		"action":   "create",
		"resource": "document",
	})

	// Execute the function
	ctx := context.Background()
	results, err := function.Execute(ctx, &event)
	if err != nil {
		fmt.Printf("Error executing function: %v\n", err)
		return
	}

	// Display results
	fmt.Printf("Function returned %d events:\n", len(results))
	for i, result := range results {
		fmt.Printf("Event %d: ID=%s, Type=%s, Source=%s\n",
			i+1, result.ID(), result.Type(), result.Source())

		var resultData map[string]interface{}
		if err := result.DataAs(&resultData); err == nil {
			fmt.Printf("  Data: %+v\n", resultData)
		}
	}
}
