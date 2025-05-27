package trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"strings"

	"github.com/nats-io/nats.go"
)

type NATSStore struct {
	nc    *nats.Conn
	kv    nats.KeyValue
	index *namespaceIndex
	mu    sync.RWMutex
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

// NewNATSStore creates a new NATS-based trigger store
func NewNATSStore(nc *nats.Conn, bucketName string) (*NATSStore, error) {
	if bucketName == "" {
		return nil, fmt.Errorf("bucket name cannot be empty")
	}

	js, err := nc.JetStream()
	if err != nil {
		return nil, fmt.Errorf("failed to get JetStream context: %w", err)
	}

	// Create KV bucket if it doesn't exist
	kv, err := js.CreateKeyValue(&nats.KeyValueConfig{
		Bucket: bucketName,
	})
	if err != nil {
		// If bucket exists, get it
		kv, err = js.KeyValue(bucketName)
		if err != nil {
			return nil, fmt.Errorf("failed to get/create KV bucket: %w", err)
		}
	}

	return &NATSStore{
		nc:    nc,
		kv:    kv,
		index: newNamespaceIndex(),
	}, nil
}

func (s *NATSStore) LoadAll(ctx context.Context) error {
	keys, err := s.kv.Keys()
	if err != nil {
		return fmt.Errorf("failed to list keys: %w", err)
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	// Create new index
	s.index = newNamespaceIndex()

	for _, key := range keys {
		entry, err := s.kv.Get(key)
		if err != nil {
			return fmt.Errorf("failed to get key %s: %w", key, err)
		}

		var trigger Trigger
		if err := json.Unmarshal(entry.Value(), &trigger); err != nil {
			return fmt.Errorf("failed to unmarshal trigger: %w", err)
		}

		s.index.addTrigger(&trigger)
	}

	return nil
}

func (s *NATSStore) Watch(ctx context.Context) {
	watcher, err := s.kv.WatchAll()
	if err != nil {
		return
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-watcher.Updates():
				if update == nil {
					continue
				}

				s.mu.Lock()
				if update.Operation() == nats.KeyValueDelete {
					// Handle deletion
					key := update.Key()
					s.index.removeTrigger(key)
				} else {
					// Handle create/update
					var trigger Trigger
					if err := json.Unmarshal(update.Value(), &trigger); err != nil {
						continue
					}

					// Remove existing trigger if it exists
					s.index.removeTrigger(trigger.ID)
					// Add updated trigger
					s.index.addTrigger(&trigger)
				}
				s.mu.Unlock()
			}
		}
	}()
}

func (s *NATSStore) GetTriggers(namespace string) []*Trigger {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.index.getTriggers(namespace)
}

func (s *NATSStore) GetAllTriggers() []*Trigger {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var allTriggers []*Trigger
	for _, trigger := range s.index.triggers {
		allTriggers = append(allTriggers, trigger)
	}
	return allTriggers
}

func (s *NATSStore) SaveTrigger(ctx context.Context, namespace, name string, trigger *Trigger) error {
	key := fmt.Sprintf("%s.%s", namespace, name)
	data, err := json.Marshal(trigger)
	if err != nil {
		return fmt.Errorf("failed to marshal trigger: %w", err)
	}

	if _, err := s.kv.Put(key, data); err != nil {
		return fmt.Errorf("failed to save trigger: %w", err)
	}

	return nil
}

func (s *NATSStore) DeleteTrigger(ctx context.Context, namespace, name string) error {
	key := fmt.Sprintf("%s.%s", namespace, name)
	if err := s.kv.Delete(key); err != nil {
		return fmt.Errorf("failed to delete trigger: %w", err)
	}

	return nil
}

func (s *NATSStore) Close() error {
	s.nc.Close()
	return nil
}
