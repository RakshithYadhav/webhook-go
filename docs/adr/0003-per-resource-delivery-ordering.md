# ADR 0003: Guarantee Per-Resource Delivery Ordering

**Issue:** [#3](https://github.com/RakshithYadhav/webhook-go/issues/3)
**Status:** Accepted
**Date:** 2026-09-03

## Acceptance criteria (given)

- Two events for the same resource ID are never delivered out of order or concurrently
  with each other.
- Events for different resource IDs are unaffected — they still deliver concurrently,
  bounded only by the worker pool from issue #2.
- The fix must not shrink the pool's effective concurrency below what #2 already
  guarantees.

## Rejected alternative: one channel + one dedicated worker per resource ID

Considered first: give each resource ID its own channel, with one worker permanently
bound to draining it, so ordering falls out for free (a channel is already FIFO).

**Why rejected:** the number of distinct resource IDs is unbounded and arbitrary — if
keeping K resources moving concurrently requires K workers assigned to them, the pool
can no longer stay fixed at the size #2's AC requires ("capped at a configured limit").
Either the pool grows past its cap, or resources beyond the cap sit with an undrained
channel — reintroducing worker-assignment coordination, plus unbounded per-resource
channel creation/teardown over the process's life. This breaks #2's design outright
rather than building on it.

## Final design (v1)

- **Keyed mutex**, not a keyed channel. The worker pool from #2 is untouched — still N
  fixed goroutines pulling from the one shared queue.
- A `sync.Mutex` per resource ID, held only for the duration of that one delivery
  (`Lock()` immediately before the HTTP send, `Unlock()` immediately after) — not for an
  entire batch, not for the life of the resource.
- Per-resource locks live in a `map[uuid.UUID]*sync.Mutex`, looked up (creating an entry
  if absent) before acquiring that resource's lock.
- That map is itself concurrently accessed by all N workers, so a **separate guard
  mutex** protects only the get-or-create step on the map — held briefly, never for the
  duration of a delivery. This is a different lock serving a different, much shorter
  critical section than the per-resource lock.
- **No cleanup of map entries.** A resource ID's "done, will never send again" can't be
  known — no signal exists for it — so no removal policy is designed. At this project's
  scale (personal capstone, not the production-scale system-design-practice track) an
  ever-growing map of small `sync.Mutex` entries is negligible; not worth the added
  complexity of proving safe removal.
- **Invariant:** a resource's lock is always released — success or panic — before the
  next queued item for that resource can proceed. Enforced via `defer Unlock()`, scoped
  correctly relative to the per-item panic-recovery boundary already in `worker.go` (not
  at the outer loop level — same class of scoping mistake already caught once in that
  file for `recover()`, logged in `MISTAKES.md`).

## Decisions and the reasoning behind them

### 1. Keyed mutex vs. keyed channel
- **Silence in the AC, raised during the §1-2 pass:** the AC says same-resource events
  can't be concurrent or out of order, but doesn't say how to serialize them without
  breaking #2's fixed-pool guarantee.
- **Decided:** keyed mutex. Workers stay generic and interchangeable, pool size stays
  fully decoupled from resource cardinality — see rejected alternative above.

### 2. Protecting the lock map itself
- **Silence in the AC, raised during the design round:** the `resourceID → lock` map is
  shared mutable state read/written by every worker; an unsynchronized get-or-create on
  it is a data race (two workers seeing "no lock for Y yet" at once, each creating and
  locking a separate mutex, defeating the whole guarantee).
- **Decided:** a guard mutex around only the lookup/create step — not around the
  delivery. Keeps the expensive part (the HTTP call) outside the map's critical section,
  so unrelated resources never block each other over map access.

### 3. Cleanup of per-resource lock entries
- **Silence in the AC, raised during the design round:** nothing says whether stale
  entries should ever be removed.
- **Decided:** no removal in v1. There's no reliable signal that a resource is
  permanently done, and map growth at this project's real scale doesn't justify solving
  a problem the project isn't actually going to hit.

### 4. Panic safety of the resource lock
- **Silence in the AC, raised during the design round:** if a delivery panics while
  holding a resource's lock and the unlock isn't guaranteed to run, every future event
  for that resource ID silently and permanently stalls — no crash, no log, just a stuck
  resource for the rest of the process's life. Exactly the "silent bad data" failure
  shape this project cares about catching, not a loud one.
- **Decided:** `defer Unlock()`, placed in the same scope as the existing per-item
  panic-recovery wrapper so it fires regardless of outcome. Exact placement is
  implementation, not design — left to guided build.

## Still open (owned by other issues)

- Shutdown/drain interaction with in-flight per-resource locks — issue #4.
- Failure recording/visibility for a panicked or failed delivery — issue #5.
