package trigger

import (
	"context"
	
	"gopkg.in/yaml.v3"
)

type Trigger struct {
	ID         string   `json:"id" yaml:"id"`
	Name       string   `json:"name" yaml:"name"`
	Namespaces []string `json:"namespaces,omitempty" yaml:"namespaces,omitempty"` // List of namespace patterns to match, "*" means all namespaces
	ObjectType string   `json:"object_type" yaml:"object_type"`
	EventType  string   `json:"event_type" yaml:"event_type"`
	// Criteria is an expression that is evaluated against the event.
	// It uses the expr language (https://github.com/expr-lang/expr) and must evaluate to a boolean.
	// Example: event.event_type == "user.created" && event.payload.after.role == "admin"
	Criteria    string `json:"criteria" yaml:"criteria"`
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	Enabled     bool   `json:"enabled" yaml:"enabled"`
	Action      string `json:"action" yaml:"action"`
}

// ToYAML marshals the trigger to YAML
func (t *Trigger) ToYAML() ([]byte, error) {
	return yaml.Marshal(t)
}

// FromYAML unmarshals the trigger from YAML
func (t *Trigger) FromYAML(data []byte) error {
	return yaml.Unmarshal(data, t)
}

// TriggerStore defines the interface for a trigger store
type TriggerStore interface {
	// LoadAll loads all triggers from the store
	LoadAll(ctx context.Context) error

	// Watch starts watching for changes to triggers
	Watch(ctx context.Context)

	// GetTriggers returns all triggers for a namespace
	GetTriggers(namespace string) []*Trigger

	// GetAllTriggers returns all triggers from all namespaces
	GetAllTriggers() []*Trigger

	// SaveTrigger saves a trigger to the store
	SaveTrigger(ctx context.Context, namespace, name string, trigger *Trigger) error

	// DeleteTrigger deletes a trigger from the store
	DeleteTrigger(ctx context.Context, namespace, name string) error

	// Close closes the store
	Close() error
}
