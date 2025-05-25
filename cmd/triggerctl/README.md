# Triggerctl

A command-line tool for managing event triggers in Mycelium.

## Installation

```bash
go install mycelium/cmd/triggerctl@latest
```

## Usage

```bash
triggerctl <command> [options]
```

### Commands

- `add <yaml-file>`    - Add a trigger from YAML file
- `list`              - List all triggers
- `delete <id>`       - Delete a trigger by ID
- `examples`          - Generate example trigger definitions

### Options

- `--nats-url`        - NATS server URL (default: nats://localhost:4222)
- `--stream`          - NATS stream name (default: config-stream)

## Examples

### Add a Trigger

```bash
# Add a trigger from YAML file
triggerctl add examples/config-update.yaml
```

### List Triggers

```bash
# List all triggers
triggerctl list
```

### Delete a Trigger

```bash
# Delete a trigger by ID
triggerctl delete config-update
```

### Generate Examples

```bash
# Generate example trigger definitions
triggerctl examples
```

## Trigger Definition Format

Triggers are defined in YAML format with the following fields:

```yaml
id: string              # Unique identifier for the trigger
name: string           # Human-readable name
namespaces: []string   # List of namespace patterns to match, "*" means all namespaces
object_type: string    # Type of object to match
event_type: string     # Type of event to match
criteria: string       # Expression to evaluate (using expr language)
enabled: boolean       # Whether the trigger is enabled
action: string         # Action to take when triggered
description: string    # Optional description
```

### Criteria Expression

The criteria field uses the [expr language](https://github.com/expr-lang/expr) to evaluate conditions. Examples:

```yaml
# Simple comparison
criteria: event.payload.after.critical == true

# Compare before and after values
criteria: event.payload.before.role != event.payload.after.role

# Numeric comparison
criteria: event.payload.after.usage > 90

# Complex condition with multiple fields
criteria: |
  event.payload.after.severity == "high" &&
  event.payload.after.source_ip != "" &&
  has(event.payload.after, "attack_type")
```

### Example Triggers

1. Config Update Notification:
```yaml
id: config-update
name: Config Update Notification
namespaces: ["default"]
object_type: Config
event_type: config.updated
criteria: event.payload.after.critical == true
enabled: true
action: notify
description: Notifies when a critical config is updated
```

2. User Role Change Detection:
```yaml
id: role-change
name: User Role Change Detection
namespaces: ["*"]
object_type: User
event_type: user.updated
criteria: event.payload.before.role != event.payload.after.role
enabled: true
action: audit
description: Detects when a user's role is changed
```

3. Resource Usage Alert:
```yaml
id: resource-alert
name: High Resource Usage Alert
namespaces: ["prod"]
object_type: Resource
event_type: resource.updated
criteria: event.payload.after.usage > 90
enabled: true
action: alert
description: Alerts when resource usage exceeds 90%
```

4. Security Breach Detection:
```yaml
id: security-breach
name: Security Breach Detection
namespaces: ["*"]
object_type: Security
event_type: security.alert
criteria: |
  event.payload.after.severity == "high" &&
  event.payload.after.source_ip != "" &&
  has(event.payload.after, "attack_type")
enabled: true
action: security-response
description: Detects potential security breaches with high severity
```

## Event Object Structure

The event object in criteria expressions has the following structure:

```go
event := {
    event_id: string
    event_type: string
    event_version: string
    namespace: string
    object_type: string
    object_id: string
    timestamp: time.Time
    actor: {
        type: string
        id: string
    }
    context: {
        request_id: string
        trace_id: string
    }
    payload: {
        before: interface{}
        after: interface{}
    }
    nats_meta: {
        stream: string
        sequence: int64
        received_at: string
    }
}
```

## Namespace Patterns

Namespace patterns support wildcards:
- `"*"` matches all namespaces
- `"prod.*"` matches all namespaces starting with "prod."
- `"*.service"` matches all namespaces ending with ".service"
- `"prod.*.service"` matches namespaces like "prod.api.service" 