## Overview

This specification defines the structure and semantics of an event system built on top of **NATS**. Events represent meaningful state changes or actions related to domain objects and are serialized for distribution across services.

## Purpose

- Provide a consistent schema for events
- Enable interoperability between producers and consumers
- Support versioning and namespace-based object separation
- Persist event history in MongoDB for queryable long-term storage
- Allow triggers to execute actions when events match specific criteria

---

## Event Structure (v1.3)

### Top-Level Fields

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `event_id` | `string (UUID v4)` | ✅ | Unique event identifier |
| `event_type` | `string` | ✅ | Semantic name (e.g., `user.created`) |
| `event_version` | `string` | ✅ | Schema version (e.g., `1.3.0`) |
| `namespace` | `string` | ✅ | Logical group or tenant (e.g., `core`, `tenant_abc`) |
| `object_type` | `string` | ✅ | Entity type (e.g., `Order`, `User`) |
| `object_id` | `string` | ✅ | Unique ID of the entity |
| `timestamp` | `string (ISO-8601)` | ✅ | UTC timestamp of the event |
| `actor` | `object` | ✅ | Entity that triggered the event |
| `context` | `object` | ⭕ | Trace and correlation info |
| `payload` | `object` | ✅ | State diff or action result |
| `nats_meta` | `object` | ⭕ | Metadata from NATS JetStream delivery |

### Subfields

### `actor`

| Field | Type | Required | Description |
| --- | --- | --- | --- |
| `type` | `string` | ✅ | E.g., `user`, `system`, `service` |
| `id` | `string` | ✅ | Identifier of the actor |

### `context`

| Field | Type | Description |
| --- | --- | --- |
| `request_id` | `string` | Request-scoped ID (for tracing) |
| `trace_id` | `string` | Distributed trace correlation ID |

### `payload`

| Field | Type | Description |
| --- | --- | --- |
| `before` | `object/null` | Previous state (optional) |
| `after` | `object/null` | New state or action result |

### `nats_meta`

| Field | Type | Description |
| --- | --- | --- |
| `stream` | `string` | JetStream stream name |
| `sequence` | `number` | Sequence number in stream |
| `received_at` | `string` | Timestamp when received by consumer |

---

## NATS Subject Convention

Events are published using a standardized subject format:

```
event.<namespace>.<object_type>.<event_type>

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

| Event Type | Description |
| --- | --- |
| `object.created` | New entity created |
| `object.updated` | Entity updated |
| `object.deleted` | Entity deleted |
| `object.status_changed` | Status field changed |
| `object.<action>` | Domain-specific action (e.g., `user.logged_in`) |

---

## Event History in MongoDB

All events SHALL be persisted in a MongoDB collection named `events` for long-term storage and auditability. Each event MUST be stored as a single document, using `event_id` as the `_id` field. The document SHOULD retain the full event payload and optionally include JetStream metadata under `nats_meta`.

### Suggested Indexes

- `{ object_id: 1 }`
- `{ namespace: 1, object_type: 1, event_type: 1, timestamp: -1 }`
- `{ timestamp: -1 }`

---

## Triggers

A **trigger** defines a condition that, when matched by an incoming event, automatically sends the event to a specified URL or API endpoint.

### Trigger Format (YAML)

```yaml
- name: Notify on admin signup
  enabled: true
  criteria: event_type == "user.created" AND payload.after.role == "admin"
  action_url: <https://example.com/webhook/notify>
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

Trigger `criteria` are defined using a simple **DSL (Domain Specific Language)** for filtering events. The expression must evaluate to `true` for the trigger to activate.

### Supported Features

### Logical Operators

- `AND`, `OR`, `NOT`

### Comparison Operators

- `==`, `!=`, `<`, `<=`, `>`, `>=`

### Accessing Event Fields

- Top-level fields: `event_type`, `namespace`, `object_type`, `timestamp`, etc.
- Nested fields: `payload.after.status`, `actor.type`, `context.trace_id`

### String Literals

- Must be enclosed in double quotes: `"user.created"`

### Examples

| Expression | Meaning |
| --- | --- |
| `event_type == "order.shipped"` | Match shipped orders |
| `namespace == "auth" AND actor.type == "user"` | User-triggered events in `auth` |
| `payload.after.status == "failed"` | Detect failure states |
| `object_type == "Invoice" AND payload.after.paid == true` | Match when invoice is paid |

### Field Existence Checks

You can check if a field exists (i.e., is not null or missing):

### Method 1: Null Comparison

```
payload.after.status != null

```

This returns `true` if `status` is present and not null.

### Method 2: `has()` Function (preferred if supported)

```
has(payload.after.status)

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
  "event_id": "13fc370e-63a3-43e7-b1f2-9db57b6f788d",
  "event_type": "user.updated",
  "event_version": "1.3.0",
  "namespace": "auth",
  "object_type": "User",
  "object_id": "user_002",
  "timestamp": "2025-04-06T12:30:00Z",
  "actor": {
    "type": "system",
    "id": "sync_service"
  },
  "context": {
    "request_id": "req_xyz789",
    "trace_id": "trace_abcd1234"
  },
  "payload": {
    "before": {
      "email": "old@example.com"
    },
    "after": {
      "email": "new@example.com"
    }
  },
  "nats_meta": {
    "stream": "EVENTS",
    "sequence": 1034,
    "received_at": "2025-04-06T12:30:01Z"
  }
}

```

---

## Job Object Overview

A **Job** is:

- An activity or unit of work
- Identified by `job_id`
- Categorized by `job_type` (e.g., `ComputeJob`, `ImageScaleJob`)
- Lives within a `namespace` to support multi-tenancy and event segregation
- Triggered by lifecycle events: `job.started`, `job.completed`, `job.failed`
- Carries inputs, status, optional result, and error information

### Job Event Lifecycle

| Event Type | Description |
| --- | --- |
| `job.started` | Job has begun |
| `job.completed` | Job finished successfully |
| `job.failed` | Job failed |

These events are emitted as `event.<namespace>.<job_type>.<event_type>` with `object_type` set to the specific job type (e.g., `ComputeJob`).

### Job Event Payload Structure

```json
{
  "job_id": "job_abc123",
  "job_type": "resize_image",
  "input": {
    "image_url": "https://...",
    "size": "800x600"
  },
  "status": "started | completed | failed",
  "result": {
    "output_url": "https://..."
  },
  "error": {
    "code": "IMG_TOO_LARGE",
    "message": "Resize failed",
    "details": {
      "max_size_mb": 10,
      "actual_size_mb": 15.2
    },
    "timestamp": "2025-04-16T14:05:00Z"
  }
}
```

### Full Event Example with Job Object

```json
{
  "event_id": "abc12345-6789-4def-0123-456789abcdef",
  "event_type": "job.started",
  "event_version": "1.3.0",
  "namespace": "mynamespace",
  "object_type": "ComputeJob",
  "object_id": "job_abc123",
  "timestamp": "2025-04-16T14:00:00Z",
  "actor": {
    "type": "system",
    "id": "job_scheduler"
  },
  "context": {
    "request_id": "req_987654321",
    "trace_id": "trace_abcdef123456"
  },
  "payload": {
    "before": null,
    "after": {
      "job_id": "job_abc123",
      "job_type": "resize_image",
      "input": {
        "image_url": "https://example.com/image.jpg",
        "size": "800x600"
      },
      "status": "started",
      "result": null,
      "error": null,
      "created_at": "2025-04-16T13:59:00Z",
      "started_at": "2025-04-16T14:00:00Z",
      "completed_at": null,
      "retries": 0,
      "depends_on": ["job_xyz789"],
      "triggered_by": "event_1234567890"
    }
  },
  "nats_meta": {
    "stream": "EVENTS",
    "sequence": 2048,
    "received_at": "2025-04-16T14:00:01Z"
  }
}
```

### NATS Subjects

```
event.mynamespace.compute_job.started
event.mynamespace.image_scale_job.completed
```

---

## 🧩 Example Event

---

⛏ Optional Features

- `created_at`, `started_at`, `completed_at`
- `retries`: number of retry attempts (if failure allowed)
- `depends_on`: array of task IDs for DAG-style orchestration
- `triggered_by`: event that caused the task to start

---

## 👀 Use Case Scenarios

- Asynchronous job system (e.g., image processing, data sync)
- Workflow steps in a saga
- Retryable background workers
- Chained tasks with manual or auto trigger

## Version History

| Version | Date | Notes |
| --- | --- | --- |
| 1.0.0 | 2025-04-06 | Initial specification |
| 1.1.0 | 2025-04-06 | Added `namespace` and subject formatting rules |
| 1.2.0 | 2025-04-06 | Added MongoDB event store and `nats_meta` field |
| 1.3.0 | 2025-04-06 | Added `Trigger` support and DSL for event-driven actions |
