package function

import (
	"context"
	"fmt"
	"sync"

	"github.com/cloudevents/sdk-go/v2/event"
	"github.com/hashicorp/go-plugin"
)

// Registry manages function plugins
type Registry struct {
	mu       sync.RWMutex
	plugins  map[string]*plugin.Client
	functions map[string]Function
}

// NewRegistry creates a new function registry
func NewRegistry() *Registry {
	return &Registry{
		plugins:   make(map[string]*plugin.Client),
		functions: make(map[string]Function),
	}
}

// RegisterPlugin registers a new function plugin
func (r *Registry) RegisterPlugin(name string, client *plugin.Client) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Get the plugin instance
	raw, err := client.Client()
	if err != nil {
		return err
	}

	function, ok := raw.(Function)
	if !ok {
		return fmt.Errorf("plugin does not implement Function interface")
	}

	r.plugins[name] = client
	r.functions[name] = function
	return nil
}

// UnregisterPlugin removes a function plugin
func (r *Registry) UnregisterPlugin(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if client, exists := r.plugins[name]; exists {
		client.Kill()
		delete(r.plugins, name)
		delete(r.functions, name)
	}
}

// ExecuteFunction executes a function by name
func (r *Registry) ExecuteFunction(ctx context.Context, name string, event *event.Event) (FunctionResult, error) {
	r.mu.RLock()
	function, exists := r.functions[name]
	r.mu.RUnlock()

	if !exists {
		return FunctionResult{}, fmt.Errorf("function %s not found", name)
	}

	return function.Execute(ctx, event)
}

// Close closes all registered plugins
func (r *Registry) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, client := range r.plugins {
		client.Kill()
	}
	r.plugins = make(map[string]*plugin.Client)
	r.functions = make(map[string]Function)
} 