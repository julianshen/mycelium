package function

import (
	"context"
	"fmt"
	"net/rpc"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// PluginManager manages function plugins
type PluginManager struct {
	plugins map[string]Plugin
	client  *plugin.Client
}

// NewPluginManager creates a new plugin manager
func NewPluginManager() *PluginManager {
	return &PluginManager{
		plugins: make(map[string]Plugin),
	}
}

// LoadPlugin loads a function plugin
func (pm *PluginManager) LoadPlugin(meta FunctionMeta, binary []byte) (Plugin, error) {
	// Create a temporary directory for the plugin
	dir, err := os.MkdirTemp("", "function-plugin-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp directory: %w", err)
	}
	defer os.RemoveAll(dir)

	// Write the plugin binary
	pluginPath := filepath.Join(dir, "plugin")
	if err := os.WriteFile(pluginPath, binary, 0755); err != nil {
		return nil, fmt.Errorf("failed to write plugin binary: %w", err)
	}

	// Create the plugin client
	client := plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig: plugin.HandshakeConfig{
			ProtocolVersion:  1,
			MagicCookieKey:   "FUNCTION_PLUGIN",
			MagicCookieValue: "function",
		},
		Plugins: map[string]plugin.Plugin{
			"function": &FunctionPlugin{},
		},
		Cmd:              exec.Command(pluginPath),
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		GRPCDialOptions: []grpc.DialOption{
			grpc.WithInsecure(),
		},
	})

	// Connect to the plugin
	rpcClient, err := client.Client()
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to connect to plugin: %w", err)
	}

	// Get the plugin instance
	raw, err := rpcClient.Dispense("function")
	if err != nil {
		client.Kill()
		return nil, fmt.Errorf("failed to dispense plugin: %w", err)
	}

	// Create the plugin wrapper
	p := &pluginWrapper{
		meta:   meta,
		client: client,
		plugin: raw.(Function),
	}

	return p, nil
}

// pluginWrapper wraps a function plugin
type pluginWrapper struct {
	meta   FunctionMeta
	client *plugin.Client
	plugin Function
}

// Name returns the name of the plugin
func (p *pluginWrapper) Name() string {
	return p.meta.Name
}

// Version returns the version of the plugin
func (p *pluginWrapper) Version() string {
	return p.meta.Version
}

// Type returns the type of the plugin
func (p *pluginWrapper) Type() string {
	return p.meta.Type
}

// Function returns the function implementation
func (p *pluginWrapper) Function() Function {
	return p.plugin
}

// FunctionPlugin is the plugin implementation
type FunctionPlugin struct {
	plugin.NetRPCUnsupportedPlugin
	Impl Function
}

// GRPCServer implements the plugin.GRPCPlugin interface
func (p *FunctionPlugin) GRPCServer(broker *plugin.GRPCBroker, s *grpc.Server) error {
	// TODO: Implement gRPC server
	return nil
}

// GRPCClient implements the plugin.GRPCPlugin interface
func (p *FunctionPlugin) GRPCClient(ctx context.Context, broker *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	// TODO: Implement gRPC client
	return nil, nil
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
	events, err := s.Impl.Execute(ctx, event)
	if err != nil {
		result.Error = err.Error()
		return nil
	}

	// For now, return the first event if available
	// TODO: Handle multiple events properly
	if len(events) > 0 {
		result.Event = events[0]
	}

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
