package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Service handles function execution through NATS
type Service struct {
	nc       *nats.Conn
	js       jetstream.JetStream
	kv       jetstream.KeyValue
	store    jetstream.ObjectStore
	registry *Registry
}

// NewService creates a new function service
func NewService(nc *nats.Conn, js jetstream.JetStream, kv jetstream.KeyValue, store jetstream.ObjectStore) *Service {
	return &Service{
		nc:       nc,
		js:       js,
		kv:       kv,
		store:    store,
		registry: NewRegistry(),
	}
}

// Start starts the function service
func (s *Service) Start(ctx context.Context) error {
	// Subscribe to function calls
	sub, err := s.nc.SubscribeSync("function.call")
	if err != nil {
		return fmt.Errorf("failed to subscribe to function calls: %w", err)
	}
	defer func() {
		if err := sub.Unsubscribe(); err != nil {
			// Log the error but don't return it since this is in a defer
			fmt.Printf("Error unsubscribing: %v\n", err)
		}
	}()

	// Watch for function updates in KV store
	watch, err := s.kv.Watch(ctx, "function.*")
	if err != nil {
		return fmt.Errorf("failed to watch function updates: %w", err)
	}
	defer func() {
		if err := watch.Stop(); err != nil {
			// Log the error but don't return it since this is in a defer
			fmt.Printf("Error stopping watch: %v\n", err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return nil
		case update := <-watch.Updates():
			if update == nil {
				continue
			}
			if err := s.handleFunctionUpdate(ctx, update); err != nil {
				return fmt.Errorf("failed to handle function update: %w", err)
			}
		default:
			msg, err := sub.NextMsgWithContext(ctx)
			if err == nil && msg != nil {
				s.handleFunctionCall(msg)
			}
		}
	}
}

// handleFunctionCall processes incoming function calls
func (s *Service) handleFunctionCall(msg *nats.Msg) {
	var call struct {
		Name  string    `json:"name"`
		Event *ce.Event `json:"event"`
	}

	data := msg.Data
	if err := json.Unmarshal(data, &call); err != nil {
		s.publishError(msg.Reply, fmt.Errorf("invalid function call: %w", err))
		return
	}

	result, err := s.registry.ExecuteFunction(context.Background(), call.Name, call.Event)
	if err != nil {
		s.publishError(msg.Reply, err)
		return
	}

	response, err := json.Marshal(result)
	if err != nil {
		s.publishError(msg.Reply, fmt.Errorf("failed to marshal response: %w", err))
		return
	}

	if err := msg.Respond(response); err != nil {
		s.publishError(msg.Reply, fmt.Errorf("failed to respond: %w", err))
	}
}

// handleFunctionUpdate processes function updates from KV store
func (s *Service) handleFunctionUpdate(ctx context.Context, update jetstream.KeyValueEntry) error {
	// Get function code from object store
	obj, err := s.store.Get(ctx, update.Key())
	if err != nil {
		return fmt.Errorf("failed to get function code: %w", err)
	}
	defer obj.Close()

	// Read function code
	_, err = io.ReadAll(obj)
	if err != nil {
		return fmt.Errorf("failed to read function code: %w", err)
	}

	// TODO: Compile and load function plugin
	// This would involve:
	// 1. Writing the code to a temporary file
	// 2. Compiling it as a plugin
	// 3. Loading it using go-plugin
	// 4. Registering it with the registry

	return nil
}

// publishError publishes an error response
func (s *Service) publishError(reply string, err error) {
	response := FunctionResult{
		Error: err.Error(),
	}
	data, _ := json.Marshal(response)
	if err := s.nc.Publish(reply, data); err != nil {
		fmt.Printf("Error publishing error response: %v\n", err)
	}
}
