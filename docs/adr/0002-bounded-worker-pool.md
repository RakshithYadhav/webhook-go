# ADR 0002: Deliver Events via a Bounded Worker Pool

**Issue:** [#2](https://github.com/RakshithYadhav/webhook-go/issues/2)
**Status:** Accepted
**Date:** 2026-08-29

## Acceptance criteria (given)

- Deliveries happen concurrently, not one-at-a-time.
- The number of concurrent in-flight deliveries is capped at a configured limit.
- A burst of events beyond the cap queues rather than spawning unbounded goroutines.

## Final design (v1)

- Fixed-size pool of goroutines, size passed as a **constructor parameter** (not
  hardcoded) — matches how `Registry` was injected into `IntakeService`.
- Each worker pulls one `DeliveryItem` from the channel `IntakeService` already owns,
  does exactly one HTTP POST to that item's endpoint, moves to the next item. No
  fan-out, no registry access — inherited directly from issue #1's design, where
  fan-out already happened at intake.
- HTTP client has a **2-second timeout**, as a named constant. Prevents a single
  slow/dead destination from hanging a worker indefinitely and silently shrinking
  effective pool capacity for the life of the process.
- Each worker **recovers from a panic per item** — one bad delivery must not take down
  the whole process (Go's default: an unrecovered panic in any goroutine kills every
  goroutine, not just the one that panicked). Recovery is scoped to a single item; the
  worker keeps running afterward.
- **No logging/failure hook.** On any delivery error (network failure, timeout,
  non-2xx), the worker just moves to the next item. Recording/visibility of failures is
  explicitly issue #5's job — #2 doesn't absorb it.
- **Shutdown is explicitly out of scope, deferred to issue #4.** The pool's consume
  loop runs indefinitely for now; no stop signal, no drain logic designed here. Same
  move as issue #1's blocked-sender-during-shutdown deferral — stated on purpose, not a
  silent gap.

## Decisions and the reasoning behind them

### 1. Pool size configuration
- **Decided:** constructor parameter, not a hardcoded constant. The AC says "configured
  limit" — a bare magic number wouldn't satisfy that, and this project already has
  precedent (`Registry` injected into `IntakeService`) for passing dependencies in
  rather than hardcoding them.

### 2. HTTP client timeout
- **Silence in the AC, raised during the §1-2 pass:** nothing required a timeout, but a
  request with none can hang a worker forever on a dead endpoint — silently shrinking
  the pool's real capacity over time, the "quiet failure" shape this project cares about
  catching.
- **Decided:** 2 seconds, as a named constant (same pattern as `queueCapacity` from #1).

### 3. Panic isolation
- **Silence in the AC, raised during the §1-2 pass:** Go's default behavior is that an
  unrecovered panic in any one goroutine kills the entire process — every other
  in-flight delivery along with it, not just the one that failed.
- **Decided:** each worker recovers from a panic scoped to the single item being
  processed, then continues to the next item. One bad delivery stays a one-item problem.

### 4. Boundary with issue #5 (failure visibility)
- **Silence in the AC, raised during the §1-2 pass:** could #2 quietly grow into doing
  #5's job (recording/logging failures) just because it's convenient to add there?
- **Decided:** no. #2's worker fails a delivery and moves on — nothing more. Recording
  what failed and why is entirely issue #5's responsibility, kept separate on purpose.

### 5. Shutdown
- **Silence in the AC, raised during the §1-2 pass:** the pool as designed here loops
  forever — nothing stops it.
- **Decided:** explicitly deferred to issue #4, stated as an assumption rather than left
  silent. Same precedent as issue #1's blocked-sender-during-shutdown deferral.

## Still open (owned by other issues)

- Shutdown mechanism for the pool itself — issue #4.
- Failure recording/visibility — issue #5.
