# ADR 0001: Accept and Route an Event for Delivery

**Issue:** [#1](https://github.com/RakshithYadhav/webhook-go/issues/1)
**Status:** Accepted
**Date:** 2026-08-28

## Acceptance criteria (given)

- Event submitted via internal function call, no external event bus for v1.
- Routed to every endpoint currently registered for its event type.
- Submission rejected (not silently dropped) if the service is shutting down.

## Final design (v1)

- `Submit(event)` is the intake function. Runs synchronously on the caller's goroutine.
- Registry: in-memory map, `eventType → [endpoints]`. Seeded at startup, not persisted,
  not runtime-mutable. Single-tenant — one customer assumed, no per-customer dimension.
- `Submit()` looks up the registry once, then pushes **one channel item per (event,
  endpoint) pair** — not one item per event. Fan-out happens at intake, not in the worker.
- Queue: bounded Go channel. Full → `Submit()` blocks (does not reject).
- Zero endpoints registered for the event's type → `Submit()` returns an error
  (non-fatal — scoped to that call, does not stop other processing).
- Shutdown interaction with an already-blocked `Submit()` call: **policy decided, mechanism
  deferred to issue #4.** Policy: once shutdown is initiated, blocked calls should be
  rejected. How that's actually implemented (a plain flag check cannot interrupt a
  goroutine already parked on the channel send) is issue #4's problem to solve.

Architecture sketch: `docs/adr/assets/0001-architecture-sketch.png`
(the Redis arrow marked "v2" in that sketch is not part of this decision — see
Deferred/v2 below).

## Decisions and the reasoning that changed them

### 1. Endpoint registry storage

- **Proposed:** Redis-backed registry, workers/intake pull from Redis into memory so
  registrations survive a process restart.
- **Issue found:** (a) conflicts with `CLAUDE.md`'s stated non-goal — "a persistent
  datastore" is explicitly out of scope for v1. (b) No issue in scope adds a runtime
  registration API, so nothing ever gets written to the registry after startup — a
  static seed already survives restarts with no external dependency needed. Redis
  would be solving a problem that doesn't exist yet in v1's scope.
- **Final:** in-memory map, seeded at startup. Redis deferred to v2 (see below).

### 2. Queue mechanism

- **Decided directly from the AC:** in-process Go channel, not an external broker.
  AC says "no external event bus required for v1." Kafka/RabbitMQ deferred to v2.

### 3. Bounded vs. unbounded channel

- **Proposed:** bounded, to throttle delivery so a slow destination doesn't get
  overloaded.
- **Issue found:** a bounded intake queue doesn't rate-limit calls to any one
  destination — it's a shared pool across all destinations. N workers can still hit
  one slow endpoint concurrently regardless of the channel's capacity. The "protects
  the destination" reasoning didn't hold up.
- **Final reasoning:** bounded channel protects this process's own memory from an
  unbounded backlog, and creates backpressure on the caller. Still bounded — just for
  a different, correct reason. Matches issue #2's "bounded worker pool" theme.

### 4. Full-channel behavior: block vs. reject

- **Decided:** block, not reject. Reasoning: an event that arrives while the channel
  is full is still a valid event — there's no reason to drop it, only a temporary
  capacity limit. Reject would mean losing legitimate data for no real fault.

### 5. Shutdown vs. a blocked `Submit()` call

- **Proposed (v1):** tag each event with a `created_at` timestamp; use it to tell
  whether a currently-blocked send started before or after shutdown began, and let
  pre-shutdown ones complete.
- **Issue found:** a goroutine blocked on a channel send is not inspectable — the
  event isn't in the channel yet, it's sitting in the sender's own stack frame.
  Nothing can read `created_at` off it while it's parked mid-send. The timestamp
  can only be read *after* the send succeeds, which is too late to base a
  reject/accept decision on.
- **Proposed (v2):** caller checks a shutdown flag before sending; if shutdown has
  started, reject immediately; otherwise proceed (may block).
- **Issue found:** (a) the check and the send aren't atomic — shutdown can begin in
  the gap between them. (b) No mechanism exists to reject a call that's *already*
  blocked when shutdown begins — a flag check has no way to reach into a parked
  goroutine. (c) The obvious tool for signaling shutdown, `close(channel)`, causes a
  **panic** ("send on closed channel") on any goroutine still blocked sending to it —
  not a clean rejection.
- **Final:** policy stands (shutdown → blocked calls rejected). The actual mechanism
  is explicitly deferred to issue #4 (graceful shutdown), since it depends on how #4
  signals shutdown in the first place.

### 6. Fan-out placement: worker-side lookup vs. intake-side lookup

- **Proposed (original sketch):** channel holds the raw event; each worker looks up
  the registry itself after dequeuing, then delivers to every matched endpoint.
- **Issue found:** one event with N registered endpoints makes a single worker
  responsible for all N deliveries — serialized on one goroutine, while the rest of
  the pool sits idle even if free. Defeats the purpose of a *bounded pool* (built for
  parallel delivery). Also muddies where partial-failure logging (issue #5) happens
  when one worker is mid-loop across multiple endpoints.
- **Final:** intake (`Submit()`) does the registry lookup once and pushes one channel
  item per (event, endpoint) pair. Every worker's job is uniform: dequeue one item,
  POST once, record the result. No worker ever touches the registry.
- **Confirmed real, not hypothetical:** the AC says "routed to **every** endpoint"
  (plural), and the original sketch already drew two separate Destination boxes —
  multiple endpoints per event type was always the intended shape.

### 7. Zero endpoints registered for an event type

- **Proposed:** throw an error — reasoning "if an event type exists it needs to be
  processed, else it shouldn't exist."
- **Issue found:** under a multi-tenant registry `(customerID, eventType) →
  [endpoints]`, zero endpoints for one customer's event is a normal, frequent case
  (most customers won't subscribe to most event types) — erroring on every occurrence
  would bury real failures in noise.
- **Resolved by scoping v1 to single-tenant** (see Decision 8) — under that
  assumption, zero endpoints for a known event type is genuinely abnormal.
- **Final:** `Submit()` returns a scoped, non-fatal error when zero endpoints match.
  Scoped = a returned error value only, not a panic; doesn't stop other events from
  processing; the caller decides what to do with it.

### 8. Multi-tenancy

- **Decided:** v1 assumes a single customer/tenant. Registry is flat —
  `eventType → [endpoints]`, no customer dimension. Kept simple deliberately for
  this capstone; this is what makes Decision 7's "zero endpoints = error" correct.
  Multi-tenant scoping is deferred to v2 — and when added, Decision 7's logic will
  need revisiting (zero-endpoints-for-one-customer becomes routine again).

## Deferred / v2 (not part of this ADR's decision, tracked for later)

- Registry storage: swap in-memory map for Redis (Decision 1).
- Queue: swap Go channel for Kafka/RabbitMQ (Decision 2).
- Registry: add multi-tenant scoping, `(customerID, eventType) → [endpoints]`
  (Decision 8) — revisit Decision 7 when this lands.
- Tracked in `CLAUDE.md` (this repo) and `BACKEND-MASTERY.md` Track 5 / Track 8.

## Still open (owned by other issues)

- Shutdown mechanism itself, and how it resolves the blocked-sender problem — issue #4.
