package function

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"
	"time"

	ce "github.com/cloudevents/sdk-go/v2"
	"github.com/nats-io/nats.go"
	"github.com/nats-io/nats.go/jetstream"
	"github.com/nats-io/nats.go/micro"
	"google.golang.org/grpc"

	pb "mycelium/internal/function/proto"
)

// Service handles function execution through gRPC
type Service struct {
	js       jetstream.JetStream
	kv       jetstream.KeyValue
	store    jetstream.ObjectStore
	registry *Registry
	server   *grpc.Server
	pb.UnimplementedFunctionServiceServer
}

// RuntimeService represents the function runtime service using NATS Service API
type RuntimeService struct {
	natsConn *nats.Conn
	service  micro.Service
	registry Registry
	plugins  map[string]Plugin
	metrics  MetricsCollector
	logger   Logger
	mu       sync.RWMutex
}

// RuntimeServiceConfig holds the configuration for the runtime service
type RuntimeServiceConfig struct {
	NATSURL     string
	ServiceName string
	Version     string
	Description string
	Registry    Registry
	Metrics     MetricsCollector
	Logger      Logger
}

// NewService creates a new function service
func NewService(js jetstream.JetStream, kv jetstream.KeyValue, store jetstream.ObjectStore) *Service {
	return &Service{
		js:    js,
		kv:    kv,
		store: store,
		// registry will be set when needed
	}
}

// NewRuntimeService creates a new runtime service using NATS Service API
func NewRuntimeService(cfg RuntimeServiceConfig) (*RuntimeService, error) {
	nc, err := nats.Connect(cfg.NATSURL)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to NATS: %w", err)
	}

	if cfg.ServiceName == "" {
		cfg.ServiceName = "function-runtime"
	}
	if cfg.Version == "" {
		cfg.Version = "1.0.0"
	}
	if cfg.Description == "" {
		cfg.Description = "Serverless function runtime service"
	}

	rs := &RuntimeService{
		natsConn: nc,
		registry: cfg.Registry,
		plugins:  make(map[string]Plugin),
		metrics:  cfg.Metrics,
		logger:   cfg.Logger,
	}

	// Create the NATS service
	serviceConfig := micro.Config{
		Name:        cfg.ServiceName,
		Version:     cfg.Version,
		Description: cfg.Description,
	}

	service, err := micro.AddService(nc, serviceConfig)
	if err != nil {
		nc.Close()
		return nil, fmt.Errorf("failed to create NATS service: %w", err)
	}

	rs.service = service

	// Add the function execution endpoint
	err = service.AddEndpoint("invoke", micro.HandlerFunc(rs.handleFunctionInvocation),
		micro.WithEndpointSubject("function.invoke"),
		micro.WithEndpointMetadata(map[string]string{
			"description": "Execute a serverless function with CloudEvents",
			"format":      "application/json",
		}))
	if err != nil {
		service.Stop()
		nc.Close()
		return nil, fmt.Errorf("failed to add invoke endpoint: %w", err)
	}

	return rs, nil
}

// Start starts the function service
func (s *Service) Start(ctx context.Context) error {
	// Create gRPC server
	s.server = grpc.NewServer()
	pb.RegisterFunctionServiceServer(s.server, s)

	// Start listening
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	// Watch for function updates in KV store
	watch, err := s.kv.Watch(ctx, "function.*")
	if err != nil {
		return fmt.Errorf("failed to watch function updates: %w", err)
	}
	defer func() {
		if err := watch.Stop(); err != nil {
			fmt.Printf("Error stopping watch: %v\n", err)
		}
	}()

	// Start watching for updates in a goroutine
	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case update := <-watch.Updates():
				if update == nil {
					continue
				}
				if err := s.handleFunctionUpdate(ctx, update); err != nil {
					fmt.Printf("Error handling function update: %v\n", err)
				}
			}
		}
	}()

	// Start gRPC server
	if err := s.server.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// ExecuteFunction implements the gRPC service
func (s *Service) ExecuteFunction(ctx context.Context, req *pb.ExecuteFunctionRequest) (*pb.ExecuteFunctionResponse, error) {
	// Convert protobuf CloudEvent to SDK CloudEvent
	event := ce.NewEvent()
	event.SetID(req.Event.Id)
	event.SetSource(req.Event.Source)
	event.SetSpecVersion(req.Event.SpecVersion)
	event.SetType(req.Event.Type)
	event.SetDataContentType(req.Event.DataContentType)
	event.SetDataSchema(req.Event.DataSchema)
	event.SetSubject(req.Event.Subject)
	event.SetTime(req.Event.Time.AsTime())
	if req.Event.Data != nil {
		event.SetData(req.Event.DataContentType, req.Event.Data)
	}
	for k, v := range req.Event.Extensions {
		event.SetExtension(k, v)
	}

	// For MVP, return an error since function execution is not implemented yet
	return &pb.ExecuteFunctionResponse{
		Result: &pb.ExecuteFunctionResponse_Error{
			Error: "function execution not implemented",
		},
	}, nil
}

// handleFunctionUpdate processes function updates from KV store
func (s *Service) handleFunctionUpdate(ctx context.Context, update jetstream.KeyValueEntry) error {
	// Get function code from object store
	obj, err := s.store.Get(ctx, update.Key())
	if err != nil {
		return fmt.Errorf("failed to get function code: %w", err)
	}
	defer obj.Close()

	// Read function code
	_, err = io.ReadAll(obj)
	if err != nil {
		return fmt.Errorf("failed to read function code: %w", err)
	}

	// TODO: Compile and load function plugin
	// This would involve:
	// 1. Writing the code to a temporary file
	// 2. Compiling it as a plugin
	// 3. Loading it using go-plugin
	// 4. Registering it with the registry

	return nil
}

// Stop stops the service
func (s *Service) Stop() {
	if s.server != nil {
		s.server.GracefulStop()
	}
}

// Start starts the runtime service
func (rs *RuntimeService) Start() error {
	rs.logger.Info("Runtime service started",
		Field{Key: "serviceName", Value: rs.service.Info().Name},
		Field{Key: "version", Value: rs.service.Info().Version})
	return nil
}

// Stop stops the runtime service
func (rs *RuntimeService) Stop() error {
	if rs.service != nil {
		rs.service.Stop()
	}
	if rs.natsConn != nil {
		rs.natsConn.Close()
	}
	rs.logger.Info("Runtime service stopped")
	return nil
}

// handleFunctionInvocation handles function invocation requests via NATS Service API
func (rs *RuntimeService) handleFunctionInvocation(req micro.Request) {
	var request struct {
		FunctionName string    `json:"functionName"`
		Event        *ce.Event `json:"event"`
	}

	if err := json.Unmarshal(req.Data(), &request); err != nil {
		rs.logger.Error("Failed to unmarshal request", Field{Key: "error", Value: err})
		rs.respondWithError(req, "invalid_request", err)
		return
	}

	// Get the function plugin
	plugin, err := rs.getPlugin(request.FunctionName)
	if err != nil {
		rs.logger.Error("Failed to get function plugin",
			Field{Key: "functionName", Value: request.FunctionName},
			Field{Key: "error", Value: err})
		rs.respondWithError(req, "plugin_not_found", err)
		return
	}

	// Execute the function
	start := time.Now()
	events, err := plugin.Function().Execute(context.Background(), request.Event)
	duration := time.Since(start)

	if err != nil {
		rs.metrics.RecordFunctionError(request.FunctionName, "execution_error")
		rs.logger.Error("Function execution failed",
			Field{Key: "functionName", Value: request.FunctionName},
			Field{Key: "error", Value: err})
		rs.respondWithError(req, "execution_error", err)
		return
	}

	// Record metrics
	rs.metrics.RecordFunctionInvocation(request.FunctionName, duration, "success")

	// Send response
	response := struct {
		Events []*ce.Event `json:"events"`
	}{
		Events: events,
	}

	responseData, err := json.Marshal(response)
	if err != nil {
		rs.logger.Error("Failed to marshal response", Field{Key: "error", Value: err})
		rs.respondWithError(req, "response_error", err)
		return
	}

	if err := req.Respond(responseData); err != nil {
		rs.logger.Error("Failed to send response", Field{Key: "error", Value: err})
	}
}

// getPlugin returns a function plugin by name
func (rs *RuntimeService) getPlugin(name string) (Plugin, error) {
	rs.mu.RLock()
	plugin, exists := rs.plugins[name]
	rs.mu.RUnlock()

	if exists {
		return plugin, nil
	}

	// Load the function from registry
	meta, binary, err := rs.registry.GetFunction(name)
	if err != nil {
		return nil, fmt.Errorf("failed to get function from registry: %w", err)
	}

	// Load the plugin
	plugin, err = rs.loadPlugin(meta, binary)
	if err != nil {
		return nil, fmt.Errorf("failed to load plugin: %w", err)
	}

	// Store the plugin
	rs.mu.Lock()
	rs.plugins[name] = plugin
	rs.mu.Unlock()

	return plugin, nil
}

// loadPlugin loads a function plugin
func (rs *RuntimeService) loadPlugin(meta FunctionMeta, binary []byte) (Plugin, error) {
	// For MVP, support built-in functions and basic plugin types
	switch meta.Type {
	case "builtin":
		// For built-in functions, we expect them to be pre-registered
		// This is a simple implementation for MVP
		if meta.Name == "example" {
			exampleFunc := &ExampleFunction{name: meta.Name}
			return &ExamplePlugin{
				meta: meta,
				fn:   exampleFunc,
			}, nil
		}
		return nil, fmt.Errorf("built-in function %s not found", meta.Name)

	case "hashicorp-plugin":
		// For HashiCorp plugins, use the plugin manager
		pluginManager := NewPluginManager()
		return pluginManager.LoadPlugin(meta, binary)

	default:
		return nil, fmt.Errorf("unsupported plugin type: %s", meta.Type)
	}
}

// respondWithError sends an error response
func (rs *RuntimeService) respondWithError(req micro.Request, errorType string, err error) {
	response := struct {
		Error     string `json:"error"`
		ErrorType string `json:"errorType"`
	}{
		Error:     err.Error(),
		ErrorType: errorType,
	}

	responseData, marshalErr := json.Marshal(response)
	if marshalErr != nil {
		rs.logger.Error("Failed to marshal error response", Field{Key: "error", Value: marshalErr})
		return
	}

	if err := req.Respond(responseData); err != nil {
		rs.logger.Error("Failed to send error response", Field{Key: "error", Value: err})
	}
}
