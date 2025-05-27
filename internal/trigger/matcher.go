package trigger

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
	cloudevents "github.com/cloudevents/sdk-go/v2"
)

var (
	ErrTriggerNotFound = errors.New("no matching trigger found")
)

// isNamespaceMatch checks if the event's namespace matches any of the trigger's namespace patterns
func isNamespaceMatch(trigger *Trigger, eventNamespace string) bool {
	// If Namespaces is empty, match all namespaces (default behavior)
	if len(trigger.Namespaces) == 0 {
		return true
	}

	// Check each namespace pattern
	for _, pattern := range trigger.Namespaces {
		// Convert pattern to regex-like string
		pattern = strings.ReplaceAll(pattern, "*", ".*")
		// Add start and end anchors
		pattern = "^" + pattern + "$"

		// Check if the namespace matches the pattern
		matched, err := regexp.MatchString(pattern, eventNamespace)
		if err == nil && matched {
			return true
		}
	}
	return false
}

// MatchTrigger returns true if the event satisfies the trigger's criteria.
// It supports:
// Expression-based matching using the expr library (preferred)
//
// Expression-based matching evaluates a string expression against the event object.
// The expression must evaluate to a boolean value.
// Example: event.event_type == "user.created" && event.payload.after.role == "admin"
//
// See the event system specification for more details on the expression language.
func MatchTrigger(trigger *Trigger, event *cloudevents.Event) (bool, error) {
	if trigger == nil || !trigger.Enabled {
		return false, nil
	}

	// If criteria is empty, match based on event type and namespace
	if trigger.Criteria == "" {
		return (trigger.EventType == "" || trigger.EventType == event.Type()) &&
			isNamespaceMatch(trigger, event.Source()) &&
			(trigger.ObjectType == "" || trigger.ObjectType == ""), nil
	}

	// If the trigger has a criteria expression, evaluate it
	return evaluateTriggerCriteria(event, trigger.Criteria)
}

// has(obj, "a.b.c") returns true if all keys exist down the path
func has(args ...any) (any, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("has() expects 2 arguments")
	}
	root, ok := args[0].(map[string]interface{})
	if !ok {
		return false, nil
	}
	path, ok := args[1].(string)
	if !ok {
		return false, nil
	}

	parts := strings.Split(path, ".")
	current := root
	for i, part := range parts {
		val, exists := current[part]
		if !exists {
			return false, nil
		}
		if i == len(parts)-1 {
			return true, nil
		}
		// If not final, it must be a nested map
		next, ok := val.(map[string]interface{})
		if !ok {
			return false, nil
		}
		current = next
	}
	return true, nil
}

// Extract extensions
func extractExtensions(event *cloudevents.Event) (string, string, string, string) {
	actorType, _ := event.Extensions()["actor_type"].(string)
	actorID, _ := event.Extensions()["actor_id"].(string)
	contextRequestID, _ := event.Extensions()["context_request_id"].(string)
	contextTraceID, _ := event.Extensions()["context_trace_id"].(string)
	return actorType, actorID, contextRequestID, contextTraceID
}

// Extract payload from Data
func extractPayload(event *cloudevents.Event) (map[string]interface{}, error) {
	var payload map[string]interface{}
	if err := event.DataAs(&payload); err != nil {
		payload = map[string]interface{}{}
	}
	return payload, nil
}

// EvaluateTriggerCriteria safely evaluates a criteria string against the given event
func evaluateTriggerCriteria(event *cloudevents.Event, criteria string) (bool, error) {
	// If criteria is empty, match based on event type and namespace
	if criteria == "" {
		// For empty criteria, we'll just return true since we don't have trigger information here
		// The actual matching based on event type and namespace is done in the MatchTrigger function
		return true, nil
	}

	// Extract extensions
	actorType, actorID, contextRequestID, contextTraceID := extractExtensions(event)

	// Extract payload from Data
	payload, err := extractPayload(event)
	if err != nil {
		return false, fmt.Errorf("failed to extract payload: %w", err)
	}

	// Only include 'before' and 'after' if present
	payloadMap := map[string]interface{}{}
	if before, ok := payload["before"]; ok {
		payloadMap["before"] = before
	}
	if after, ok := payload["after"]; ok {
		payloadMap["after"] = after
	}

	// Create a map representation of the event that matches JSON field names
	eventMap := map[string]interface{}{
		"event_id":      event.ID(),
		"event_type":    event.Type(),
		"event_version": event.SpecVersion(),
		"namespace":     event.Source(),
		"object_type":   "", // Not present in CloudEvent, unless you want to add as extension
		"object_id":     event.ID(),
		"timestamp":     event.Time(),
		"actor": map[string]interface{}{
			"type": actorType,
			"id":   actorID,
		},
		"context": map[string]interface{}{
			"request_id": contextRequestID,
			"trace_id":   contextTraceID,
		},
		"payload": payloadMap,
		// NATS metadata can be extracted from the NATS extension if needed
	}

	// Create environment with event as the root variable
	env := map[string]interface{}{
		"event": eventMap,
	}

	// Compile the expression with custom functions
	options := []expr.Option{
		expr.Env(env),
		expr.Function("has", has),
	}

	program, err := expr.Compile(criteria, options...)
	if err != nil {
		return false, fmt.Errorf("failed to compile criteria: %w", err)
	}

	// Run the compiled expression
	output, err := expr.Run(program, env)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate criteria: %w", err)
	}

	// Must return boolean
	result, ok := output.(bool)
	if !ok {
		return false, fmt.Errorf("expression did not return a boolean")
	}

	return result, nil
}

// FindMatchingTriggers finds all triggers that match the given event.
// Returns an empty slice if no matching triggers are found.
func FindMatchingTriggers(store TriggerStore, event *cloudevents.Event) ([]*Trigger, error) {
	// Get all potential triggers for the namespace (including wildcard matches)
	triggers := store.GetTriggers(event.Source())
	if len(triggers) == 0 {
		return nil, nil
	}

	// Check each trigger and collect matches
	var matchingTriggers []*Trigger
	for _, trigger := range triggers {
		matches, err := MatchTrigger(trigger, event)
		if err != nil {
			return nil, fmt.Errorf("error matching trigger %s: %w", trigger.ID, err)
		}
		if matches {
			matchingTriggers = append(matchingTriggers, trigger)
		}
	}

	return matchingTriggers, nil
}
