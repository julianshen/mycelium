package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"mycelium/internal/event"
	"mycelium/internal/trigger"

	"github.com/nats-io/nats.go"
)

func main() {
	// Parse command line flags
	natsURL := flag.String("nats-url", "nats://localhost:4222", "NATS server URL")
	streamName := flag.String("stream", "config-stream", "NATS stream name")
	subject := flag.String("subject", "config.>", "NATS subject to subscribe to")
	queueGroup := flag.String("queue-group", "trigger-processors", "NATS queue group name")
	durableName := flag.String("durable", "trigger-consumer", "NATS durable consumer name")
	flag.Parse()

	// Connect to NATS
	nc, err := nats.Connect(*natsURL)
	if err != nil {
		log.Fatalf("Failed to connect to NATS: %v", err)
	}
	defer nc.Close()

	// Create NATS store for triggers
	store, err := trigger.NewNATSStore(nc, *streamName)
	if err != nil {
		log.Fatalf("Failed to create trigger store: %v", err)
	}
	defer store.Close()

	// Load triggers
	ctx := context.Background()
	if err := store.LoadAll(ctx); err != nil {
		log.Fatalf("Failed to load triggers: %v", err)
	}

	// Start watching for trigger changes
	go store.Watch(ctx)

	// Create event handler
	handler := func(e *event.Event) error {
		matchedTriggers, err := trigger.FindMatchingTriggers(store, e)
		if err != nil {
			log.Printf("Error finding matching triggers: %v", err)
			return err
		}

		if len(matchedTriggers) > 0 {
			log.Printf("Event %s matched %d triggers:", e.EventID, len(matchedTriggers))
			for _, t := range matchedTriggers {
				log.Printf("  - Trigger: %s", t.Name)
				log.Printf("    Action: %s", t.Action)
				// Here you would execute the actual action
				// For now, we just print the action
			}
		}
		return nil
	}

	// Create watcher configuration
	config := event.WatcherConfig{
		URL:           *natsURL,
		StreamName:    *streamName,
		Subject:       *subject,
		QueueGroup:    *queueGroup,
		DurableName:   *durableName,
		AckWait:       30 * time.Second,
		MaxDeliveries: 5,
	}

	// Create the watcher
	watcher, err := event.NewWatcher(config, handler)
	if err != nil {
		log.Fatalf("Failed to create watcher: %v", err)
	}

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Handle OS signals
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start watching for events
	if err := watcher.Start(ctx); err != nil {
		log.Fatalf("Failed to start watcher: %v", err)
	}

	log.Printf("Trigger daemon started. Watching for events...")
	log.Printf("Press Ctrl+C to stop")

	// Wait for signal
	<-sigChan
	log.Printf("Shutting down...")
}
