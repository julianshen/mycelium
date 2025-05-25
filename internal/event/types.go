package event

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// Actor represents the entity that triggered the event
type Actor struct {
	Type string `json:"type"` // e.g., "user", "system", "service"
	ID   string `json:"id"`   // Identifier of the actor
}

// Context contains trace and correlation information
type Context struct {
	RequestID string `json:"request_id,omitempty"` // Request-scoped ID (for tracing)
	TraceID   string `json:"trace_id,omitempty"`   // Distributed trace correlation ID
}

// NATSMeta contains metadata from NATS JetStream delivery
type NATSMeta struct {
	Stream     string `json:"stream"`      // JetStream stream name
	Sequence   int64  `json:"sequence"`    // Sequence number in stream
	ReceivedAt string `json:"received_at"` // Timestamp when received by consumer
}

// Event represents a system event following the v1.3 specification
type Event struct {
	EventID      string    `json:"event_id"`          // Unique event identifier (UUID v4)
	EventType    string    `json:"event_type"`        // Semantic name (e.g., "user.created")
	EventVersion string    `json:"event_version"`     // Schema version (e.g., "1.3.0")
	Namespace    string    `json:"namespace"`         // Logical group or tenant
	ObjectType   string    `json:"object_type"`       // Entity type (e.g., "Order", "User")
	ObjectID     string    `json:"object_id"`         // Unique ID of the entity
	Timestamp    time.Time `json:"timestamp"`         // UTC timestamp of the event
	Actor        Actor     `json:"actor"`             // Entity that triggered the event
	Context      *Context  `json:"context,omitempty"` // Trace and correlation info
	Payload      struct {
		Before interface{} `json:"before,omitempty"` // Previous state
		After  interface{} `json:"after,omitempty"`  // New state or action result
	} `json:"payload"` // State diff or action result
	NATSMeta *NATSMeta `json:"nats_meta,omitempty"` // Metadata from NATS JetStream delivery
}

// NewEvent creates a new Event with required fields
func NewEvent(eventType, namespace, objectType, objectID string, actor Actor) *Event {
	return &Event{
		EventID:      generateUUID(), // You'll need to implement this
		EventType:    eventType,
		EventVersion: "1.3.0",
		Namespace:    namespace,
		ObjectType:   objectType,
		ObjectID:     objectID,
		Timestamp:    time.Now().UTC(),
		Actor:        actor,
	}
}

// SetPayload sets the before and after states of the event
func (e *Event) SetPayload(before, after interface{}) {
	e.Payload.Before = before
	e.Payload.After = after
}

// SetContext sets the context information for the event
func (e *Event) SetContext(requestID, traceID string) {
	e.Context = &Context{
		RequestID: requestID,
		TraceID:   traceID,
	}
}

// SetNATSMeta sets the NATS metadata for the event
func (e *Event) SetNATSMeta(stream string, sequence int64) {
	e.NATSMeta = &NATSMeta{
		Stream:     stream,
		Sequence:   sequence,
		ReceivedAt: time.Now().UTC().Format(time.RFC3339),
	}
}

// generateUUID generates a new UUID v4
func generateUUID() string {
	return uuid.New().String()
}

// UnmarshalJSON implements custom JSON unmarshaling for Event
func (e *Event) UnmarshalJSON(data []byte) error {
	// Create a temporary struct to handle the JSON unmarshaling
	type Alias Event
	aux := &struct {
		Timestamp string `json:"timestamp"`
		*Alias
	}{
		Alias: (*Alias)(e),
	}

	if err := json.Unmarshal(data, &aux); err != nil {
		return fmt.Errorf("failed to unmarshal event: %w", err)
	}

	// Parse the timestamp
	if aux.Timestamp != "" {
		t, err := time.Parse(time.RFC3339, aux.Timestamp)
		if err != nil {
			return fmt.Errorf("failed to parse timestamp: %w", err)
		}
		e.Timestamp = t
	}

	return nil
}
