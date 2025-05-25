package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"

	"mycelium/internal/event"

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
	// Create config update event
	evt := event.NewEvent(
		"config.updated",
		"default",
		"Config",
		"app-config",
		event.Actor{
			Type: "user",
			ID:   "test-user",
		},
	)

	evt.SetContext("test-req-1", "test-trace-1")
	evt.SetPayload(
		map[string]interface{}{
			"critical": false,
			"value":    "old-value",
		},
		map[string]interface{}{
			"critical": true,
			"value":    "new-value",
		},
	)

	// Publish event
	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("Failed to marshal config event: %v", err)
		return
	}

	_, err = js.Publish("events.config.updated", data)
	if err != nil {
		log.Printf("Failed to publish config event: %v", err)
		return
	}

	log.Println("Emitted config update event")
}

func emitUserRoleChange(ctx context.Context, js nats.JetStreamContext) {
	// Create user role change event
	evt := event.NewEvent(
		"user.updated",
		"default",
		"User",
		"test-user",
		event.Actor{
			Type: "admin",
			ID:   "admin-user",
		},
	)

	evt.SetContext("test-req-2", "test-trace-2")
	evt.SetPayload(
		map[string]interface{}{
			"role": "user",
			"name": "Test User",
		},
		map[string]interface{}{
			"role": "admin",
			"name": "Test User",
		},
	)

	// Publish event
	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("Failed to marshal user event: %v", err)
		return
	}

	_, err = js.Publish("events.user.updated", data)
	if err != nil {
		log.Printf("Failed to publish user event: %v", err)
		return
	}

	log.Println("Emitted user role change event")
}

func emitResourceUsage(ctx context.Context, js nats.JetStreamContext) {
	// Create resource usage event
	evt := event.NewEvent(
		"resource.updated",
		"prod",
		"Resource",
		"server-1",
		event.Actor{
			Type: "system",
			ID:   "monitor",
		},
	)

	evt.SetContext("test-req-3", "test-trace-3")
	evt.SetPayload(
		map[string]interface{}{
			"usage": 75.5,
			"type":  "cpu",
		},
		map[string]interface{}{
			"usage": 95.2,
			"type":  "cpu",
		},
	)

	// Publish event
	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("Failed to marshal resource event: %v", err)
		return
	}

	_, err = js.Publish("events.resource.updated", data)
	if err != nil {
		log.Printf("Failed to publish resource event: %v", err)
		return
	}

	log.Println("Emitted resource usage event")
}

func emitSecurityAlert(ctx context.Context, js nats.JetStreamContext) {
	// Create security alert event
	evt := event.NewEvent(
		"security.alert",
		"prod",
		"Security",
		"alert-1",
		event.Actor{
			Type: "system",
			ID:   "security-scanner",
		},
	)

	evt.SetContext("test-req-4", "test-trace-4")
	evt.SetPayload(
		map[string]interface{}{
			"severity": "low",
			"status":   "investigating",
		},
		map[string]interface{}{
			"severity":    "high",
			"status":      "active",
			"source_ip":   "192.168.1.100",
			"attack_type": "brute_force",
		},
	)

	// Publish event
	data, err := json.Marshal(evt)
	if err != nil {
		log.Printf("Failed to marshal security event: %v", err)
		return
	}

	_, err = js.Publish("events.security.alert", data)
	if err != nil {
		log.Printf("Failed to publish security event: %v", err)
		return
	}

	log.Println("Emitted security alert event")
}
