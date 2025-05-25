# Triggerd

A daemon service that watches for events and executes triggers in Mycelium.

## Overview

Triggerd is a daemon service that:
1. Connects to NATS and subscribes to events
2. Loads trigger definitions from NATS KV store
3. Matches incoming events against trigger criteria
4. Executes actions when triggers match

## Installation

```bash
go install mycelium/cmd/triggerd@latest
```

## Usage

```bash
triggerd [options]
```

### Options

- `--nats-url`        - NATS server URL (default: nats://localhost:4222)
- `--stream`          - NATS stream name (default: config-stream)
- `--queue-group`     - Queue group name for load balancing (default: triggerd)

## Configuration

### NATS Connection

The daemon connects to NATS and requires:
- NATS server with JetStream enabled
- KV bucket named `triggers` for storing trigger definitions
- Stream for receiving events

### Queue Groups

Multiple instances of triggerd can run in a queue group for load balancing:
- All instances in the same queue group share the event processing load
- Each event is processed by exactly one instance
- Instances can be added/removed without affecting event processing

## Event Processing

1. **Event Reception**
   - Subscribes to events from the configured stream
   - Events are received in order within each queue group

2. **Trigger Matching**
   - Loads trigger definitions from NATS KV store
   - Matches event against trigger criteria using expr language
   - Supports complex conditions and pattern matching

3. **Action Execution**
   - When a trigger matches, executes the configured action
   - Actions are executed asynchronously
   - Failed actions are logged but don't block event processing

## Example Setup

1. Start NATS with JetStream:
```bash
nats-server -js
```

2. Start triggerd:
```bash
# Single instance
triggerd

# Multiple instances in queue group
triggerd --queue-group=prod
triggerd --queue-group=prod
```

3. Add triggers using triggerctl:
```bash
triggerctl add examples/config-update.yaml
```

## Monitoring

The daemon logs:
- Connection status to NATS
- Trigger matches and actions
- Errors in event processing
- Action execution results

## Troubleshooting

### Common Issues

1. **NATS Connection Failed**
   - Check NATS server is running
   - Verify NATS URL is correct
   - Ensure JetStream is enabled

2. **No Triggers Loaded**
   - Verify KV bucket exists
   - Check trigger definitions are added
   - Look for errors in trigger parsing

3. **Actions Not Executing**
   - Check trigger criteria matches events
   - Verify action configuration
   - Look for errors in action execution

### Logging

Enable debug logging for more details:
```bash
triggerd --log-level=debug
```

## Development

### Building

```bash
go build -o triggerd cmd/triggerd/main.go
```

### Testing

```bash
go test ./cmd/triggerd/...
```

## Architecture

```
                    ┌─────────────┐
                    │    NATS     │
                    │   Server    │
                    └──────┬──────┘
                           │
                           ▼
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│  Triggerd   │    │  Triggerd   │    │  Triggerd   │
│ Instance 1  │    │ Instance 2  │    │ Instance 3  │
└──────┬──────┘    └──────┬──────┘    └──────┬──────┘
       │                  │                  │
       └──────────────────┼──────────────────┘
                          │
                          ▼
                    ┌─────────────┐
                    │   Actions   │
                    └─────────────┘
```

- Multiple triggerd instances form a queue group
- Events are distributed among instances
- Each instance processes its events independently
- Actions are executed by the matching instance 