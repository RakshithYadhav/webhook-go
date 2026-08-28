# Design Feedback Log

Design-skill feedback after each design round on this project — not coding mistakes
(that's a `MISTAKES.md`-style log, if one gets added later), specifically about how the
*design* reasoning went. One dated entry per round.

## 2026-08-28 — Issue #1 design round (Accept and route an event)

Context: first design round on this project, issue #1. ~1 hour, 8 decisions resolved,
written up in `docs/adr/0001-accept-and-route-event.md`.

**Strengths:**

- Updates on evidence fast, without digging in. Bounded-channel reasoning was wrong
  ("protects the destination") — rebuilt the justification correctly the moment a
  counter-example was shown, same with fan-out placement. Didn't defend a claim just
  because it had already been said out loud.
- Caught own scope creep against own written constraints. Pushed for Redis, recognized
  it conflicted with `CLAUDE.md`'s stated non-goals, chose to defer it as a stated v2
  decision rather than dropping the idea or building it anyway.
- Asked for a shape-level pass before designing issue #1 in isolation — avoided
  designing one component blind to the boxes around it.

**Growth areas:**

- Default fix for a concurrency problem was "attach more data" (a `created_at`
  timestamp) rather than "use the right control-flow primitive." The
  blocked-sender-during-shutdown race is fundamentally about *what code is running, on
  which goroutine, at which moment* — not something a data field can resolve after the
  fact. Watch for this pattern: when a race feels solvable by "just checking a field,"
  that's often the signal a different tool is needed (signal channel, `select`,
  `context`), not more data on the struct.
- Reasoning that sounded right in prose broke under a concrete traced-through example,
  more than once ("workers should know the destination," "bounded protects the
  destination," the timestamp idea). Exercise for next round: trace one concrete
  example through a proposed mechanism *before* presenting it, not after being
  challenged on it.

**Pace note:** an hour to 8 resolved decisions + a written ADR is a good pace for a
first real design round — the friction was doing the actual work of finding a first
answer was wrong, not slowness.
