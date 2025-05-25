# Trigger Test Events

This directory contains tools for testing triggers by emitting test events to NATS.

## Test Events

The test script emits four types of events that match the example triggers:

1. **Config Update Event**
   - Event Type: `config.updated`
   - Namespace: `default`
   - Object Type: `Config`
   - Payload: Changes `critical` flag from `false` to `true`

2. **User Role Change Event**
   - Event Type: `user.updated`
   - Namespace: `default`
   - Object Type: `User`
   - Payload: Changes `role` from `user` to `admin`

3. **Resource Usage Event**
   - Event Type: `resource.updated`
   - Namespace: `prod`
   - Object Type: `Resource`
   - Payload: Changes `usage` from `75.5` to `95.2`

4. **Security Alert Event**
   - Event Type: `security.alert`
   - Namespace: `prod`
   - Object Type: `Security`
   - Payload: Changes severity and adds attack details

## Usage

1. Start NATS with JetStream:
```bash
nats-server -js
```

2. Start triggerd:
```bash
triggerd
```

3. Add test triggers:
```bash
triggerctl add examples/config-update.yaml
triggerctl add examples/role-change.yaml
triggerctl add examples/resource-alert.yaml
triggerctl add examples/security-breach.yaml
```

4. Run the test events:
```bash
go run cmd/triggerd/test/emit_test_events.go
```

## Command Line Options

The test script supports the following options:

- `--nats-url` - NATS server URL (default: nats://localhost:4222)
- `--stream` - NATS stream name (default: config-stream)

Example:
```bash
go run cmd/triggerd/test/emit_test_events.go --nats-url=nats://nats:4222 --stream=my-stream
```

## Expected Results

When running the test events, you should see:

1. Config Update Trigger:
   - Matches when `critical` becomes `true`
   - Action: `notify`

2. User Role Change Trigger:
   - Matches when role changes
   - Action: `audit`

3. Resource Usage Trigger:
   - Matches when usage exceeds 90%
   - Action: `alert`

4. Security Breach Trigger:
   - Matches when severity is high and attack details present
   - Action: `security-response`

## Troubleshooting

If triggers are not matching:

1. Check triggerd logs for event reception
2. Verify trigger criteria matches event payload
3. Ensure events are being published to correct stream
4. Check NATS connection and stream configuration

## Adding Custom Test Events

To add custom test events:

1. Create a new function in `emit_test_events.go`
2. Use `event.NewEvent()` to create the event
3. Set context and payload using helper methods
4. Add the function call to `main()`
5. Publish to appropriate subject

Example:
```go
func emitCustomEvent(ctx context.Context, js nats.JetStreamContext) {
    evt := event.NewEvent(
        "custom.event",
        "namespace",
        "ObjectType",
        "object-id",
        event.Actor{
            Type: "system",
            ID:   "test",
        },
    )
    
    evt.SetContext("req-id", "trace-id")
    evt.SetPayload(
        map[string]interface{}{
            "before": "old",
        },
        map[string]interface{}{
            "after": "new",
        },
    )
    
    data, _ := json.Marshal(evt)
    js.Publish("events.custom.event", data)
}
``` 