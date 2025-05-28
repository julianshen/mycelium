package function

import (
	"context"
	"net/rpc"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/hashicorp/go-plugin"
)

// FunctionResult represents the result returned from a function
type FunctionResult struct {
	Event *event.Event `json:"event"`
	Error string      `json:"error,omitempty"`
}

// Function is the interface that all function plugins must implement
type Function interface {
	Execute(ctx context.Context, event *event.Event) (FunctionResult, error)
}

// FunctionPlugin is the plugin implementation for functions
type FunctionPlugin struct {
	Impl Function
}

func (p *FunctionPlugin) Server(*plugin.MuxBroker) (interface{}, error) {
	return &FunctionServer{Impl: p.Impl}, nil
}

func (p *FunctionPlugin) Client(b *plugin.MuxBroker, c *rpc.Client) (interface{}, error) {
	return &FunctionClient{client: c}, nil
}

// FunctionServer is the RPC server for functions
type FunctionServer struct {
	Impl Function
}

// Execute implements the RPC call for function execution
func (s *FunctionServer) Execute(ctx context.Context, event *event.Event, result *FunctionResult) error {
	res, err := s.Impl.Execute(ctx, event)
	if err != nil {
		return err
	}
	*result = res
	return nil
}

// FunctionClient is the RPC client for functions
type FunctionClient struct {
	client *rpc.Client
}

// Execute calls the remote function implementation
func (c *FunctionClient) Execute(ctx context.Context, event *event.Event) (FunctionResult, error) {
	var result FunctionResult
	err := c.client.Call("Plugin.Execute", event, &result)
	return result, err
} 