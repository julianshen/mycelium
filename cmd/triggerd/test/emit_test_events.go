package main

import (
	"context"
	"flag"
	"log"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
)

func main() {
	// Parse command line flags
	natsURL := flag.String("nats-url", "nats://localhost:4222", "NATS server URL")
	streamName := flag.String("stream", "config-stream", "NATS stream name")
	flag.Parse()

	// Connect to NATS
	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Create JetStream context
	js, err := nc.JetStream()
	if err != nil {
		log.Fatalf("Failed to create JetStream context: %v", err)
	}

	// Create stream if it doesn't exist
	stream, err := js.StreamInfo(*streamName)
	if err != nil {
		_, err = js.AddStream(&nats.StreamConfig{
			Name:     *streamName,
			Subjects: []string{"events.>"},
		})
		if err != nil {
			log.Fatalf("Failed to create stream: %v", err)
		}
	} else {
		log.Printf("Using existing stream: %s", stream.Config.Name)
	}

	// Emit test events
	ctx := context.Background()
	emitConfigUpdate(ctx, js)
	emitUserRoleChange(ctx, js)
	emitResourceUsage(ctx, js)
	emitSecurityAlert(ctx, js)
}

func emitConfigUpdate(ctx context.Context, js nats.JetStreamContext) {
	ce := cloudevents.NewEvent()
	ce.SetID("app-config")
	ce.SetSource("mycelium/test")
	ce.SetType("config.updated")
	ce.SetExtension("actor_type", "user")
	ce.SetExtension("actor_id", "test-user")
	ce.SetExtension("context_request_id", "test-req-1")
	ce.SetExtension("context_trace_id", "test-trace-1")
	if err := ce.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"before": map[string]interface{}{
			"critical": false,
			"value":    "old-value",
		},
		"after": map[string]interface{}{
			"critical": true,
			"value":    "new-value",
		},
	}); err != nil {
		log.Printf("Failed to set config CloudEvent data: %v", err)
		return
	}

	data, err := ce.MarshalJSON()
	if err != nil {
		log.Printf("Failed to marshal config CloudEvent: %v", err)
		return
	}

	_, err = js.Publish("events.config.updated", data)
	if err != nil {
		log.Printf("Failed to publish config CloudEvent: %v", err)
		return
	}

	log.Println("Emitted config update CloudEvent")
}

func emitUserRoleChange(ctx context.Context, js nats.JetStreamContext) {
	ce := cloudevents.NewEvent()
	ce.SetID("test-user")
	ce.SetSource("mycelium/test")
	ce.SetType("user.updated")
	ce.SetExtension("actor_type", "admin")
	ce.SetExtension("actor_id", "admin-user")
	ce.SetExtension("context_request_id", "test-req-2")
	ce.SetExtension("context_trace_id", "test-trace-2")
	if err := ce.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"before": map[string]interface{}{
			"role": "user",
			"name": "Test User",
		},
		"after": map[string]interface{}{
			"role": "admin",
			"name": "Test User",
		},
	}); err != nil {
		log.Printf("Failed to set user CloudEvent data: %v", err)
		return
	}

	data, err := ce.MarshalJSON()
	if err != nil {
		log.Printf("Failed to marshal user CloudEvent: %v", err)
		return
	}

	_, err = js.Publish("events.user.updated", data)
	if err != nil {
		log.Printf("Failed to publish user CloudEvent: %v", err)
		return
	}

	log.Println("Emitted user role change CloudEvent")
}

func emitResourceUsage(ctx context.Context, js nats.JetStreamContext) {
	ce := cloudevents.NewEvent()
	ce.SetID("server-1")
	ce.SetSource("mycelium/test")
	ce.SetType("resource.updated")
	ce.SetExtension("actor_type", "system")
	ce.SetExtension("actor_id", "monitor")
	ce.SetExtension("context_request_id", "test-req-3")
	ce.SetExtension("context_trace_id", "test-trace-3")
	if err := ce.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"before": map[string]interface{}{
			"usage": 75.5,
			"type":  "cpu",
		},
		"after": map[string]interface{}{
			"usage": 95.2,
			"type":  "cpu",
		},
	}); err != nil {
		log.Printf("Failed to set resource CloudEvent data: %v", err)
		return
	}

	data, err := ce.MarshalJSON()
	if err != nil {
		log.Printf("Failed to marshal resource CloudEvent: %v", err)
		return
	}

	_, err = js.Publish("events.resource.updated", data)
	if err != nil {
		log.Printf("Failed to publish resource CloudEvent: %v", err)
		return
	}

	log.Println("Emitted resource usage CloudEvent")
}

func emitSecurityAlert(ctx context.Context, js nats.JetStreamContext) {
	ce := cloudevents.NewEvent()
	ce.SetID("alert-1")
	ce.SetSource("mycelium/test")
	ce.SetType("security.alert")
	ce.SetExtension("actor_type", "system")
	ce.SetExtension("actor_id", "security-scanner")
	ce.SetExtension("context_request_id", "test-req-4")
	ce.SetExtension("context_trace_id", "test-trace-4")
	if err := ce.SetData(cloudevents.ApplicationJSON, map[string]interface{}{
		"before": map[string]interface{}{
			"severity": "low",
			"status":   "investigating",
		},
		"after": map[string]interface{}{
			"severity":    "high",
			"status":      "active",
			"source_ip":   "192.168.1.100",
			"attack_type": "brute_force",
		},
	}); err != nil {
		log.Printf("Failed to set security CloudEvent data: %v", err)
		return
	}

	data, err := ce.MarshalJSON()
	if err != nil {
		log.Printf("Failed to marshal security CloudEvent: %v", err)
		return
	}

	_, err = js.Publish("events.security.alert", data)
	if err != nil {
		log.Printf("Failed to publish security CloudEvent: %v", err)
		return
	}

	log.Println("Emitted security alert CloudEvent")
}
