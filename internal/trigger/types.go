package trigger

import (
	"context"
	"strings"

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

// namespaceIndex maintains an index of triggers by namespace pattern
type namespaceIndex struct {
	// exact matches: namespace -> []triggerID
	exactMatches map[string][]string
	// pattern matches: pattern -> []triggerID
	patternMatches map[string][]string
	// all triggers by ID
	triggers map[string]*Trigger
}

func newNamespaceIndex() *namespaceIndex {
	return &namespaceIndex{
		exactMatches:   make(map[string][]string),
		patternMatches: make(map[string][]string),
		triggers:       make(map[string]*Trigger),
	}
}

func (idx *namespaceIndex) addTrigger(trigger *Trigger) {
	idx.triggers[trigger.ID] = trigger

	// If no namespaces specified, add to pattern matches with "*"
	if len(trigger.Namespaces) == 0 {
		idx.patternMatches["*"] = append(idx.patternMatches["*"], trigger.ID)
		return
	}

	// Add to appropriate index based on pattern type
	for _, pattern := range trigger.Namespaces {
		if strings.Contains(pattern, "*") {
			idx.patternMatches[pattern] = append(idx.patternMatches[pattern], trigger.ID)
		} else {
			idx.exactMatches[pattern] = append(idx.exactMatches[pattern], trigger.ID)
		}
	}
}

func (idx *namespaceIndex) removeTrigger(triggerID string) {
	// Check if trigger exists
	if _, exists := idx.triggers[triggerID]; !exists {
		return
	}

	// Remove from triggers map
	delete(idx.triggers, triggerID)

	// Remove from exact matches
	for namespace, ids := range idx.exactMatches {
		newIds := make([]string, 0, len(ids))
		for _, id := range ids {
			if id != triggerID {
				newIds = append(newIds, id)
			}
		}
		if len(newIds) == 0 {
			delete(idx.exactMatches, namespace)
		} else {
			idx.exactMatches[namespace] = newIds
		}
	}

	// Remove from pattern matches
	for pattern, ids := range idx.patternMatches {
		newIds := make([]string, 0, len(ids))
		for _, id := range ids {
			if id != triggerID {
				newIds = append(newIds, id)
			}
		}
		if len(newIds) == 0 {
			delete(idx.patternMatches, pattern)
		} else {
			idx.patternMatches[pattern] = newIds
		}
	}
}

func (idx *namespaceIndex) getTriggers(namespace string) []*Trigger {
	var triggerIDs []string

	// Get exact matches
	if ids, exists := idx.exactMatches[namespace]; exists {
		triggerIDs = append(triggerIDs, ids...)
	}

	// Get pattern matches
	for pattern, ids := range idx.patternMatches {
		if pattern == "*" || isNamespaceMatch(&Trigger{Namespaces: []string{pattern}}, namespace) {
			triggerIDs = append(triggerIDs, ids...)
		}
	}

	// Convert IDs to triggers
	triggers := make([]*Trigger, 0, len(triggerIDs))
	seen := make(map[string]bool)
	for _, id := range triggerIDs {
		if !seen[id] {
			if trigger, exists := idx.triggers[id]; exists {
				triggers = append(triggers, trigger)
				seen[id] = true
			}
		}
	}

	return triggers
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
