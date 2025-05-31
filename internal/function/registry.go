package function

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
)

// NATSRegistry implements the Registry interface using NATS
type NATSRegistry struct {
	nc          *nats.Conn
	js          jetstream.JetStream
	kv          jetstream.KeyValue
	objectStore jetstream.ObjectStore
}

// NewNATSRegistry creates a new NATS registry
func NewNATSRegistry(nc *nats.Conn) (*NATSRegistry, error) {
	js, err := jetstream.New(nc)
	if err != nil {
		return nil, fmt.Errorf("failed to create jetstream: %w", err)
	}

	// Create or get the KV bucket
	kv, err := js.CreateKeyValue(context.Background(), jetstream.KeyValueConfig{
		Bucket: "functions",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create KV bucket: %w", err)
	}

	// Create or get the object store bucket
	objectStore, err := js.CreateObjectStore(context.Background(), jetstream.ObjectStoreConfig{
		Bucket: "function-binaries",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create object store: %w", err)
	}

	return &NATSRegistry{
		nc:          nc,
		js:          js,
		kv:          kv,
		objectStore: objectStore,
	}, nil
}

// StoreFunction stores a function's metadata and binary
func (r *NATSRegistry) StoreFunction(meta FunctionMeta, binary []byte) error {
	// Store the metadata
	metaData, err := json.Marshal(meta)
	if err != nil {
		return fmt.Errorf("failed to marshal metadata: %w", err)
	}

	_, err = r.kv.Put(context.Background(), meta.Name, metaData)
	if err != nil {
		return fmt.Errorf("failed to store metadata: %w", err)
	}

	// Store the binary
	_, err = r.objectStore.PutBytes(context.Background(), meta.Name, binary)
	if err != nil {
		return fmt.Errorf("failed to store binary: %w", err)
	}

	return nil
}

// GetFunction retrieves a function's metadata and binary
func (r *NATSRegistry) GetFunction(name string) (FunctionMeta, []byte, error) {
	// Get the metadata
	entry, err := r.kv.Get(context.Background(), name)
	if err != nil {
		return FunctionMeta{}, nil, fmt.Errorf("failed to get metadata: %w", err)
	}

	var meta FunctionMeta
	if err := json.Unmarshal(entry.Value(), &meta); err != nil {
		return FunctionMeta{}, nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
	}

	// Get the binary
	binary, err := r.objectStore.GetBytes(context.Background(), name)
	if err != nil {
		return FunctionMeta{}, nil, fmt.Errorf("failed to get binary: %w", err)
	}

	return meta, binary, nil
}

// ListFunctions returns a list of all available functions
func (r *NATSRegistry) ListFunctions() ([]FunctionMeta, error) {
	keys, err := r.kv.Keys(context.Background())
	if err != nil {
		return nil, fmt.Errorf("failed to list functions: %w", err)
	}

	var functions []FunctionMeta
	for _, key := range keys {
		entry, err := r.kv.Get(context.Background(), key)
		if err != nil {
			return nil, fmt.Errorf("failed to get function %s: %w", key, err)
		}

		var meta FunctionMeta
		if err := json.Unmarshal(entry.Value(), &meta); err != nil {
			return nil, fmt.Errorf("failed to unmarshal function %s: %w", key, err)
		}

		functions = append(functions, meta)
	}

	return functions, nil
}

// DeleteFunction removes a function
func (r *NATSRegistry) DeleteFunction(name string) error {
	// Delete the metadata
	if err := r.kv.Delete(context.Background(), name); err != nil {
		return fmt.Errorf("failed to delete metadata: %w", err)
	}

	// Delete the binary
	if err := r.objectStore.Delete(context.Background(), name); err != nil {
		return fmt.Errorf("failed to delete binary: %w", err)
	}

	return nil
}
