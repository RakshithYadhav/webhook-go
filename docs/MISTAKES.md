# Mistakes Log

One line per critical mistake per coding step — syntax or logic errors caught during
guided build, not design-round feedback (that's `FEEDBACK.md`). Newest entries at top
of each day's section.

## 2026-08-28

- `internal/registry/registry.go`: lowercased the wrong identifier when closing the exported-field race hole — unexported the `Seed` method instead of the `Endpoints` field, which would've blocked `main.go` from seeding it at all while leaving the field still directly writable from outside the package.
- `internal/registry/registry.go`: `endpoints` map field was exported (`Endpoints`) — any package could mutate it directly, bypassing `Seed`, which breaks the "read-only after startup" invariant once concurrent access starts (unsynchronized map read+write race).
- `internal/event/event.go`: struct type left unexported (`type event struct`) — compiled fine standing alone, but no other package could ever reference it; only the fields were capitalized, not the type itself.
- `internal/event/event.go`: struct declaration keyword order reversed (`type Struct event { ... }` instead of `type Event struct { ... }`) — didn't compile.
