package event

import (
	"context"
	"fmt"
	"log"
	"time"

	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
)

// WatcherConfig holds the configuration for the NATS event watcher
type WatcherConfig struct {
	URL           string        // NATS server URL
	StreamName    string        // JetStream stream name
	Subject       string        // Subject to subscribe to
	QueueGroup    string        // Queue group name (optional)
	DurableName   string        // Durable consumer name
	AckWait       time.Duration // How long to wait for ACK
	MaxDeliveries int           // Maximum number of delivery attempts
}

// EventHandler is a function type that processes events
type EventHandler func(*cloudevents.Event) error

// Watcher represents a NATS event watcher
type Watcher struct {
	conn    *nats.Conn
	js      nats.JetStreamContext
	sub     *nats.Subscription
	config  WatcherConfig
	handler EventHandler
}

// NewWatcher creates a new NATS event watcher
func NewWatcher(config WatcherConfig, handler EventHandler) (*Watcher, error) {
	// Connect to NATS
	nc, err := nats.Connect(config.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	// Create JetStream Context
	js, err := nc.JetStream()
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create JetStream context: %w", err)
	}

	return &Watcher{
		conn:    nc,
		js:      js,
		config:  config,
		handler: handler,
	}, nil
}

// Start begins watching for events
func (w *Watcher) Start(ctx context.Context) error {
	// Create consumer configuration
	consumerConfig := &nats.ConsumerConfig{
		Durable:       w.config.DurableName,
		AckPolicy:     nats.AckExplicitPolicy,
		DeliverPolicy: nats.DeliverNewPolicy,
		AckWait:       w.config.AckWait,
		MaxDeliver:    w.config.MaxDeliveries,
	}

	// Create or update the consumer
	_, err := w.js.AddConsumer(w.config.StreamName, consumerConfig)
	if err != nil {
		return fmt.Errorf("failed to create consumer: %w", err)
	}

	// Subscribe to the subject
	var sub *nats.Subscription
	if w.config.QueueGroup != "" {
		sub, err = w.js.QueueSubscribe(w.config.Subject, w.config.QueueGroup, w.handleMessage)
	} else {
		sub, err = w.js.Subscribe(w.config.Subject, w.handleMessage)
	}
	if err != nil {
		return fmt.Errorf("failed to subscribe: %w", err)
	}

	w.sub = sub

	// Wait for context cancellation
	go func() {
		<-ctx.Done()
		w.Stop()
	}()

	return nil
}

// Stop stops watching for events
func (w *Watcher) Stop() {
	if w.sub != nil {
		if err := w.sub.Unsubscribe(); err != nil {
			log.Printf("Error unsubscribing: %v", err)
		}
	}
	if w.conn != nil {
		w.conn.Close()
	}
}

// handleMessage processes incoming NATS messages
func (w *Watcher) handleMessage(msg *nats.Msg) {
	// Parse the CloudEvent
	ce := cloudevents.NewEvent()
	if err := ce.UnmarshalJSON(msg.Data); err != nil {
		log.Printf("Error unmarshaling CloudEvent: %v", err)
		if err := msg.Nak(); err != nil {
			log.Printf("Error sending NAK: %v", err)
		}
		return
	}

	// Optionally extract NATS metadata using the NATS extension if needed
	// Optionally extract Actor and Context from extensions if needed

	if err := w.handler(&ce); err != nil {
		log.Printf("Error processing CloudEvent: %v", err)
		if err := msg.Nak(); err != nil {
			log.Printf("Error sending NAK: %v", err)
		}
		return
	}

	if err := msg.Ack(); err != nil {
		log.Printf("Error sending ACK: %v", err)
	}
}
