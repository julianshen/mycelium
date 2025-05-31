package function

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
)

// Client represents a function client that communicates with NATS Service API
type Client struct {
	nc       *nats.Conn
	registry Registry
	timeout  time.Duration
}

// ClientConfig holds the configuration for the client
type ClientConfig struct {
	NATSURL  string
	Registry Registry
	Timeout  time.Duration
}

// NewClient creates a new function client
func NewClient(cfg ClientConfig) (*Client, error) {
	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 30 * time.Second
	}

	return &Client{
		nc:       nc,
		registry: cfg.Registry,
		timeout:  cfg.Timeout,
	}, nil
}

// InvokeFunction invokes a function with the given event using NATS Service API
func (c *Client) InvokeFunction(ctx context.Context, name string, event *ce.Event) ([]*ce.Event, error) {
	// Create request
	req := struct {
		FunctionName string    `json:"functionName"`
		Event        *ce.Event `json:"event"`
	}{
		FunctionName: name,
		Event:        event,
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Use NATS Service API endpoint subject for function invocation
	// The service listens on "function.invoke" as defined in the service
	responseMsg, err := c.nc.RequestWithContext(ctx, "function.invoke", reqData)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	// Parse response
	var resp struct {
		Events    []*ce.Event `json:"events,omitempty"`
		Error     string      `json:"error,omitempty"`
		ErrorType string      `json:"errorType,omitempty"`
	}

	if err := json.Unmarshal(responseMsg.Data, &resp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.Error != "" {
		return nil, fmt.Errorf("function error (%s): %s", resp.ErrorType, resp.Error)
	}

	return resp.Events, nil
}

// Close closes the client
func (c *Client) Close() {
	c.nc.Close()
}
