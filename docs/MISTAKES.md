# Mistakes Log

One line per critical mistake per coding step — syntax or logic errors caught during
guided build, not design-round feedback (that's `FEEDBACK.md`). Newest entries at top
of each day's section.

## 2026-08-29 (session 3 — issue #2, worker pool test, TDD)

- `internal/worker/worker_test.go`: `TestWorkerPool`'s `wg.Wait()` was wrapped in its own `go func() { ... }()`, so the test function returned immediately without ever waiting for it — the test "passed" in 0.00s without actually verifying anything happened.
- `internal/worker/worker_test.go`: each worker goroutine did a single `item := <-queue` receive instead of looping — only 3 of 15 submitted items ever got consumed, the other 12 sat in the channel untouched, but the bug above masked this since nothing was actually being waited on.

## 2026-08-29 (session 2 — issue #2, worker pool, TDD)

- `internal/worker/worker_test.go`: `TestTimeOut` wasn't updated when `SendDeliveryItem` became a method on `Worker` — kept calling the old free function, which no longer existed in that form.
- `internal/worker/worker_test.go`: `worker = New()` used `=` instead of `:=` for a not-yet-declared variable; adjacent line had a typo (`woSendDeliveryItem` instead of `worker.SendDeliveryItem`).
- `internal/worker/worker.go`: constructor written as `func (w *Worker) New() *Worker` — same backwards pattern as `IntakeService.Constructor` earlier (building the thing requires an instance of the thing to already exist); a repeat of a previously-caught mistake, not a new one.
- `internal/worker/worker.go`: struct comment repeated an earlier incorrect claim (that holding a `Client` on the struct was "for connection pooling") after that claim had already been corrected in conversation — worth double-checking comments against corrections made afterward, not just the code.
- `internal/worker/worker.go`: captured both return values of `http.Post` with `res, err :=` but used neither — Go requires every declared local variable to be used; since the minimal version didn't need them yet, the call should've been a bare statement instead.

## 2026-08-29

- `internal/intake/service_test.go`: zero-endpoint test's assertion was inverted — `if err == ErrNoEndPoints { t.Fatalf(...) }` fails the test exactly when the code behaves correctly, and would silently pass if `Submit` had a real bug (e.g. returned `nil` instead of the expected error). Should assert on the negative (`err != ErrNoEndPoints`).

## 2026-08-28 (session 2)

- `internal/intake/service.go`: `Submit` returned `ErrNoEndPoints` before its signature declared an `error` return type at all — the two changes needed to happen together, not just the return statement.
- `internal/intake/service.go`: declared a sentinel error with `const ErrNoEndPoints = errors.New(...)` — `errors.New(...)` runs at runtime, so it can't be a `const` (which requires a compile-time constant); needs `var`.
- `internal/intake/service.go`: `register` field was `registry.Registry` (by value) instead of `*registry.Registry` (pointer) — every method on `Registry` uses a pointer receiver; storing by value risks `IntakeService` holding a stale copy if `Seed()` reassigns the map after the struct was copied.
- `internal/intake/service.go`: missing trailing comma after `queue: make(...)` in the composite literal — didn't compile.
- `internal/intake/service.go`: channel capacity was a bare `100` inline instead of a named constant — no compile issue, but undocumented and hard to find/tune later.
- `internal/intake/service.go`: `make(chan deliveryitem.DeliveryItem)` created with no capacity argument at all — an unbuffered channel (capacity 0), not the bounded-with-room-to-absorb-a-burst channel the ADR calls for.
- `internal/intake/service.go`: constructor written as a method on `*IntakeService` (`func (i *IntakeService) Constructor()`) instead of a plain function — backwards, since building the instance requires an instance to already exist; the receiver `i` was also never used in the body.
- `internal/deliveryItem/deliveryItem.go`: field type written as bare `Event`/`event` instead of package-qualified `event.Event` — confused the type name with the package name twice in a row (same root cause as the import-path miss).
- `internal/deliveryItem/deliveryItem.go`: import path written as `"internal/event"` instead of the full module path `"github.com/RakshithYadhav/webhook-go/internal/event"` — Go doesn't resolve local packages by their relative folder path.
- `internal/deliveryItem/deliveryService.go`: created as a completely empty file (no `package` declaration) — broke the build for the whole module, not just this package, since every `.go` file needs at least a package clause.
- `internal/deliveryItem/deliveryItem.go`: field was `Endpoints []string` (plural) — reopened the ADR's fan-out decision; a `DeliveryItem` must hold exactly one endpoint, since intake already splits one event into one item per endpoint.

## 2026-08-28

- `internal/registry/registry.go`: lowercased the wrong identifier when closing the exported-field race hole — unexported the `Seed` method instead of the `Endpoints` field, which would've blocked `main.go` from seeding it at all while leaving the field still directly writable from outside the package.
- `internal/registry/registry.go`: `endpoints` map field was exported (`Endpoints`) — any package could mutate it directly, bypassing `Seed`, which breaks the "read-only after startup" invariant once concurrent access starts (unsynchronized map read+write race).
- `internal/event/event.go`: struct type left unexported (`type event struct`) — compiled fine standing alone, but no other package could ever reference it; only the fields were capitalized, not the type itself.
- `internal/event/event.go`: struct declaration keyword order reversed (`type Struct event { ... }` instead of `type Event struct { ... }`) — didn't compile.
