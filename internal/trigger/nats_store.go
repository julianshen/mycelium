package trigger

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/nats-io/nats.go"
)

type NATSStore struct {
	nc    *nats.Conn
	kv    nats.KeyValue
	index *namespaceIndex
	mu    sync.RWMutex
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
