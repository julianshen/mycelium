## Overview

This specification defines the structure and semantics of an event system built on top of **NATS** using the **CloudEvents** specification. Events represent meaningful state changes or actions related to domain objects and are serialized for distribution across services.

## Purpose

- Provide a consistent schema for events using CloudEvents
- Enable interoperability between producers and consumers
- Support versioning and namespace-based object separation
- Persist event history in MongoDB for queryable long-term storage
- Allow triggers to execute actions when events match specific criteria

---

## Event Structure (v1.4)

### CloudEvents Required Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `id` | `string (UUID v4)` | ✅ | Unique event identifier |
| `source` | `string` | ✅ | Event source (e.g., `mycelium/core`) |
| `specversion` | `string` | ✅ | CloudEvents spec version (e.g., `1.0`) |
| `type` | `string` | ✅ | Semantic name (e.g., `user.created`) |
| `time` | `string (ISO-8601)` | ✅ | UTC timestamp of the event |
| `data` | `object` | ✅ | Event payload |

### CloudEvents Optional Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `datacontenttype` | `string` | ⭕ | Content type of data (e.g., `application/json`) |
| `dataschema` | `string` | ⭕ | Schema URL for data validation |
| `subject` | `string` | ⭕ | Subject of the event |

### Custom Extensions

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `namespace` | `string` | ✅ | Logical group or tenant (e.g., `core`, `tenant_abc`) |
| `objecttype` | `string` | ✅ | Entity type (e.g., `Order`, `User`) |
| `objectid` | `string` | ✅ | Unique ID of the entity |
| `eventversion` | `string` | ✅ | Schema version (e.g., `1.3.0`) |
| `actor` | `object` | ✅ | Entity that triggered the event |
| `context` | `object` | ⭕ | Trace and correlation info |
| `natsmeta` | `object` | ⭕ | Metadata from NATS JetStream delivery |

### Subfields

### `actor`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `type` | `string` | ✅ | E.g., `user`, `system`, `service` |
| `id` | `string` | ✅ | Identifier of the actor |

### `context`

| Field | Type | Description |
| --- | --- | --- |
| `requestid` | `string` | Request-scoped ID (for tracing) |
| `traceid` | `string` | Distributed trace correlation ID |

### `data`

| Field | Type | Description |
| --- | --- | --- |
| `before` | `object/null` | Previous state (optional) |
| `after` | `object/null` | New state or action result |

### `natsmeta`

| Field | Type | Description |
| --- | --- | --- |
| `stream` | `string` | JetStream stream name |
| `sequence` | `number` | Sequence number in stream |
| `receivedat` | `string` | Timestamp when received by consumer |

---

## NATS Subject Convention

Events are published using a standardized subject format:

```
event.<namespace>.<objecttype>.<type>

```

### Examples

- `event.tenant_abc.order.status_changed`
- `event.core.user.created`
- `event.auth.session.expired`

### Wildcard Subscriptions

- `event.tenant_abc.>` → All events for a tenant
- `event.*.user.*` → All user events across namespaces

---

## Event Type Conventions
The format should be "$namespace.object.{command|event}"

| Event Type | Description | Example |
| --- | --- | --- |
| `object.created` | New entity created | `core.user.created` |
| `object.updated` | Entity updated | `core.user.updated` |
| `object.deleted` | Entity deleted | `core.user.deleted` |
| `object.status_changed` | Status field changed | `core.user.status_changed` |
| `object.<action>` | Domain-specific action | `core.user.logged_in` |

---

## Event History in MongoDB

All events SHALL be persisted in a MongoDB collection named `events` for long-term storage and auditability. Each event MUST be stored as a single document, using `id` as the `_id` field. The document SHOULD retain the full event payload and optionally include JetStream metadata under `natsmeta`.

### Suggested Indexes

- `{ objectid: 1 }`
- `{ namespace: 1, objecttype: 1, type: 1, time: -1 }`
- `{ time: -1 }`

---

## Triggers

A **trigger** defines a condition that, when matched by an incoming event, automatically sends the event to a specified URL or API endpoint.

### Trigger Format (YAML)

```yaml
- name: Notify on admin signup
  enabled: true
  criteria: type == "user.created" AND data.after.role == "admin"
  action_url: https://example.com/webhook/notify
  retry_count: 3
  timeout: 5
```

### Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `name` | `string` | ✅ | Trigger name |
| `enabled` | `bool` | ✅ | Whether the trigger is active |
| `criteria` | `string` | ✅ | Expression that defines match logic |
| `action_url` | `string` | ✅ | Webhook or endpoint to send matching events |
| `retry_count` | `int` | ⭕ | Number of retry attempts on failure |
| `timeout` | `int` | ⭕ | Request timeout in seconds |

---

## DSL for Trigger Criteria

The DSL (Domain Specific Language) is used to define logical criteria for matching events. If an incoming event satisfies the criteria, the trigger is activated and its action is invoked.

### Supported Features

### Logical Operators

- `AND`, `OR`, `NOT`

### Comparison Operators

- `==`, `!=`, `<`, `<=`, `>`, `>=`

### Accessing Event Fields

- Top-level fields: `type`, `namespace`, `objecttype`, `time`, etc.
- Nested fields: `data.after.status`, `actor.type`, `context.traceid`

### String Literals

- Must be enclosed in double quotes: `"user.created"`

### Examples

| Expression | Meaning |
| --- | --- |
| `type == "order.shipped"` | Match shipped orders |
| `namespace == "auth" AND actor.type == "user"` | User-triggered events in `auth` |
| `data.after.status == "failed"` | Detect failure states |
| `objecttype == "Invoice" AND data.after.paid == true` | Match when invoice is paid |

### Field Existence Checks

You can check if a field exists (i.e., is not null or missing):

### Method 1: Null Comparison

```
data.after.status != null
```

This returns `true` if `status` is present and not null.

### Method 2: `has()` Function (preferred if supported)

```
has(data.after.status)
```

Returns `true` if the field exists, even if its value is null. Safer in strict evaluators.

### Notes

- Fields not present in the event are treated as `null`
- Boolean values should use `true`/`false`
- Strings must be quoted with `"` (double quotes)

---

## Example Event (JSON)

```json
{
  "id": "13fc370e-63a3-43e7-b1f2-9db57b6f788d",
  "source": "mycelium/core",
  "specversion": "1.0",
  "type": "user.updated",
  "time": "2025-04-06T12:30:00Z",
  "datacontenttype": "application/json",
  "namespace": "auth",
  "objecttype": "User",
  "objectid": "user_002",
  "eventversion": "1.3.0",
  "actor": {
    "type": "system",
    "id": "sync_service"
  },
  "context": {
    "requestid": "req_xyz789",
    "traceid": "trace_abcd1234"
  },
  "data": {
    "before": {
      "email": "old@example.com"
    },
    "after": {
      "email": "new@example.com"
    }
  },
  "natsmeta": {
    "stream": "EVENTS",
    "sequence": 1034,
    "receivedat": "2025-04-06T12:30:01Z"
  }
}
```

## Version History

| Version | Date | Notes |
| --- | --- | --- |
| 1.0.0 | 2025-04-06 | Initial specification |
| 1.1.0 | 2025-04-06 | Added `namespace` and subject formatting rules |
| 1.2.0 | 2025-04-06 | Added MongoDB event store and `natsmeta` field |
| 1.3.0 | 2025-04-06 | Added `Trigger` support and DSL for event-driven actions |
| 1.4.0 | 2025-05-28 | Migrated to CloudEvents specification |
