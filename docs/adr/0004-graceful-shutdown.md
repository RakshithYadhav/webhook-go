# ADR 0004: Graceful Shutdown with In-Flight Drain

**Issue:** [#4](https://github.com/RakshithYadhav/webhook-go/issues/4)
**Status:** Accepted
**Date:** 2026-09-03

## Acceptance criteria (given)

- Once shutdown begins, new event submissions are rejected with a clear error, not
  silently accepted and dropped.
- Deliveries already in flight are allowed to finish, bounded by a grace-period deadline.
- If the deadline is hit before draining completes, the service returns/logs that
  explicitly rather than hanging forever.

## Final design (v1)

Shutdown is two separate mechanisms, one per component, because "stop accepting" and
"finish what's already accepted" are different problems with different shapes:

- **`IntakeService.Shutdown()`** — no arguments, no waiting. Flips an `atomic.Bool`
  that `Submit()` checks first; once set, `Submit()` returns `ErrShutdownError`
  immediately instead of queueing anything.
- **`Pool.Shutdown(ctx context.Context) error`** — races a `sync.WaitGroup` reaching
  zero against the caller-supplied context's deadline. Returns `nil` if the drain
  finishes first, `ctx.Err()` if the deadline is hit first. The grace period itself is
  not stored on `Pool` at all — it's whatever deadline the caller builds into the
  context it passes in (`main.go` uses `context.WithTimeout(ctx, 10*time.Second)`).
- **One `*sync.WaitGroup`, created once in `main.go`**, passed by pointer into both
  `intake.New(...)` and `worker.NewPool(...)`. `Add(1)` is called inside `Submit`'s
  existing fan-out loop — once per `DeliveryItem` produced, before that item is sent to
  the channel. `Done()` is called inside the worker's per-item inner function, deferred
  alongside the existing panic recovery and the resource-lock unlock, so it always fires
  regardless of outcome.
- **No shutdown check inside the worker's consume loop.** Workers keep pulling from the
  channel and processing whatever's there unconditionally — the WaitGroup is what tells
  `Shutdown` when drainage is actually complete, not a flag the loop consults per item.
- **Orchestration lives in `main.go`**, not inside either component: on receiving
  `os.Interrupt`/`SIGTERM` (via `signal.NotifyContext`, `main()` blocked on
  `<-ctx.Done()`), it calls `intakeService.Shutdown()` first, then
  `pool.Shutdown(shCtx)` — in that order, so the pool's deadline clock starts against a
  fixed, no-longer-growing backlog rather than a moving target.

## Rejected alternative: a shutdown flag checked inside the worker's loop

Considered first, symmetric with intake's design: give `Pool` the same
`atomic.Bool`-checked-per-item pattern as `IntakeService.Submit`.

**Why rejected:** intake's flag gates *new* work arriving; the worker's loop only ever
sees work that was already accepted. A flag check per item would reject items already
sitting in the queue at the moment `Shutdown` is called — directly violating "already
in flight [accepted] deliveries are allowed to finish." Concrete case: 3 items are
queued but unstarted when shutdown begins; a per-item flag check would throw them all
out instead of draining them.

## Rejected alternative: narrowing "in flight" to only the currently-executing item

Considered next: only guarantee completion for whatever a worker is mid-delivery on
*right now*; anything still sitting in the queue, untouched, is abandoned on shutdown.

**Why rejected:** the issue's own user story states the purpose directly — "so that a
deploy never silently drops an event that was already accepted." An event that's been
`Submit`'d successfully (no error) has been accepted; abandoning it because no worker
happened to reach it yet is exactly the silent-drop failure the issue exists to prevent.

## Decisions and the reasoning behind them

### 1. Two independent shutdown mechanisms, not one shared one
- **Silence in the AC, raised during the design round:** the AC bundles "reject new
  work" and "drain existing work" together, but they're owned by different components
  (`IntakeService` vs. `Pool`) with genuinely different shapes — one is instant, one is
  bounded by a deadline.
- **Decided:** separate methods, separate mechanisms, orchestrated by the caller.

### 2. WaitGroup ownership and sharing
- **Silence in the AC, raised during the design round:** both components need to
  operate on the same completion-tracking state. Same class of problem this project
  already solved for `Registry` (needs `*Registry`, not a value, so `IntakeService`
  always sees live state) and `ResourceLock` (needs `*ResourceLock`, so every worker
  goroutine shares one map of locks) — `sync.WaitGroup` can't be safely copied either.
- **Decided:** created once in `main.go`, injected as `*sync.WaitGroup` into both
  constructors.

### 3. `Add()` placement: once per `DeliveryItem`, before the channel send
- **Silence in the AC, raised during the design round:** `Submit` already fans one
  event out into multiple `DeliveryItem`s (one per registered endpoint, from issue #1's
  design). `Done()` naturally happens once per delivered item. If `Add()` were called
  once per event instead of once per item, a multi-endpoint event would produce more
  `Done()` calls than `Add()` calls — `sync.WaitGroup` panics ("negative WaitGroup
  counter") the moment that happens, not just miscounts silently.
- **Decided:** `Add(1)` inside the existing fan-out loop, once per item, and
  specifically *before* `i.queue <- item` — the item becomes visible to a worker (which
  could immediately finish and call `Done()`) the instant it's queued, so `Add()` must
  happen first.

### 4. No flag check in the worker's loop
- Covered above under rejected alternatives — the drain requirement directly rules this
  out.

### 5. Orchestration order: intake first, then pool
- **Silence in the AC, raised during the design round:** if the pool's deadline clock
  started before intake stopped accepting new submissions, new events could keep
  arriving throughout the grace period, and the pool would be racing a deadline against
  a backlog that never stops growing.
- **Decided:** `main.go` calls `intakeService.Shutdown()` before `pool.Shutdown(shCtx)`,
  guaranteeing the backlog is fixed and known before the deadline starts.

### 6. Grace period: caller-supplied via `context.Context`, not stored on `Pool`
- **Open question from the design round:** constructor parameter (like pool size) or
  named constant (like the HTTP timeout)?
- **Decided, once actually wired up:** neither — `Pool.Shutdown` takes a
  `context.Context` and races its deadline, so the grace period is whatever the caller
  builds into that context. This is the standard Go idiom for propagating a deadline,
  and it means `Pool` itself doesn't need to know or care what the grace period is.

## Still open (owned by other issues)

- What happens to items still undelivered when the deadline is hit (beyond returning an
  error from `Shutdown`) — logging/recording that explicitly is issue #5's job.
- No automated tests for the shutdown behavior itself yet (submission rejection
  post-shutdown, the drain-vs-deadline race in both outcomes) — deliberately skipped for
  this unit, tests deferred.
