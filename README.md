# Mycelium

A distributed event processing system with NATS integration, supporting event triggers and actions.

## Overview

Mycelium is a system for processing events and executing triggers based on configurable criteria. It consists of two main components:

1. **Triggerd**: A daemon service that watches for events and executes triggers
2. **Triggerctl**: A CLI tool for managing triggers

## Features

- Event processing with NATS JetStream
- Expression-based trigger criteria
- Namespace-based event routing
- Queue group support for load balancing
- YAML-based trigger configuration
- CLI for trigger management
- Test utilities for event simulation

## Components

### Triggerd

The daemon service that:
- Connects to NATS and subscribes to events
- Loads trigger definitions from NATS KV store
- Matches events against trigger criteria
- Executes actions when triggers match

[More details in triggerd README](cmd/triggerd/README.md)

### Triggerctl

The CLI tool for:
- Adding triggers from YAML files
- Listing existing triggers
- Deleting triggers
- Generating example trigger definitions

[More details in triggerctl README](cmd/triggerctl/README.md)

## Quick Start

1. Start NATS with JetStream:
```bash
nats-server -js
```

2. Start the trigger daemon:
```bash
go run cmd/triggerd/main.go
```

3. Add a trigger:
```bash
go run cmd/triggerctl/main.go add examples/config-update.yaml
```

4. Test with sample events:
```bash
go run cmd/triggerd/test/emit_test_events.go
```

## Installation

### From Source

```bash
# Clone the repository
git clone https://github.com/julianshen/mycelium.git
cd mycelium

# Build the binaries
go build -o bin/triggerd cmd/triggerd/main.go
go build -o bin/triggerctl cmd/triggerctl/main.go
```

### Using Go Install

```bash
go install github.com/julianshen/mycelium/cmd/triggerd@latest
go install github.com/julianshen/mycelium/cmd/triggerctl@latest
```

## Development

### Prerequisites

- Go 1.21 or later
- NATS server with JetStream enabled

### Building

```bash
# Build all components
go build ./...

# Run tests
go test ./...

# Run linter
golangci-lint run
```

### Project Structure

```
.
├── cmd/
│   ├── triggerd/          # Daemon service
│   │   ├── main.go
│   │   ├── README.md
│   │   └── test/          # Test utilities
│   └── triggerctl/        # CLI tool
│       ├── main.go
│       ├── README.md
│       └── examples/      # Example triggers
├── internal/
│   ├── event/            # Event types and watcher
│   └── trigger/          # Trigger types and matcher
└── .github/
    └── workflows/        # CI/CD configuration
```

## Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

## License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## Acknowledgments

- [NATS](https://nats.io/) for the messaging system
- [expr](https://github.com/expr-lang/expr) for the expression language 