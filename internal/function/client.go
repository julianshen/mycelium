package function

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// Client is used to call functions
type Client struct {
	nc *nats.Conn
	js jetstream.JetStream
}

// NewClient creates a new function client
func NewClient(nc *nats.Conn, js jetstream.JetStream) *Client {
	return &Client{
		nc: nc,
		js: js,
	}
}

// CallFunction calls a function by name with the given event
func (c *Client) CallFunction(ctx context.Context, name string, event *ce.Event) (FunctionResult, error) {
	// Create request
	request := struct {
		Name  string      `json:"name"`
		Event *ce.Event   `json:"event"`
	}{
		Name:  name,
		Event: event,
	}

	data, err := json.Marshal(request)
	if err != nil {
		return FunctionResult{}, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create a reply subject
	reply := fmt.Sprintf("function.reply.%s", nats.NewInbox())

	// Create a subscription for the reply
	sub, err := c.nc.SubscribeSync(reply)
	if err != nil {
		return FunctionResult{}, fmt.Errorf("failed to subscribe to reply: %w", err)
	}
	defer sub.Unsubscribe()

	// Publish the request
	if err := c.nc.PublishRequest("function.call", reply, data); err != nil {
		return FunctionResult{}, fmt.Errorf("failed to publish request: %w", err)
	}

	// Wait for reply with timeout
	msg, err := sub.NextMsgWithContext(ctx)
	if err != nil {
		return FunctionResult{}, fmt.Errorf("failed to receive reply: %w", err)
	}

	var result FunctionResult
	if err := json.Unmarshal(msg.Data, &result); err != nil {
		return FunctionResult{}, fmt.Errorf("failed to unmarshal reply: %w", err)
	}

	if result.Error != "" {
		return result, fmt.Errorf("function error: %s", result.Error)
	}

	return result, nil
}

// CallFunctionWithTimeout calls a function with a timeout
func (c *Client) CallFunctionWithTimeout(name string, event *ce.Event, timeout time.Duration) (FunctionResult, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return c.CallFunction(ctx, name, event)
} 