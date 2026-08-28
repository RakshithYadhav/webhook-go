# CLAUDE.md

This file provides guidance to Claude Code when working with code in this repository.

## What this is

A webhook delivery service in Go — the Track 1 capstone of `BACKEND-MASTERY.md`'s
go-concurrency track, modeled on the concurrency patterns from Rakshith's shippio
take-home (worker pool, keyed mutex, graceful shutdown). Personal project, built to
practice and demonstrate those mechanisms plus Docker/Kubernetes deployment — not a
system that served real customers. See the honesty boundary in `docs/RESUME.md`.

**Working method — read this before writing code.**

1. **Design decisions stay his, made through a two-step process, before code exists.**
   - **Step A — the §1-2 pass**, from his own framework at
     `C:\Users\Rakshith\Documents\Journal\Thoughts\algorithmic-thinking.md`: restate the
     issue in his own words; hunt the silences (what the ticket doesn't specify — ties,
     ordering, edge cases, concurrent access); ask what "being wrong" looks like (silent
     bad data vs. a loud failure). Flag gaps the way you'd flag them to a PM/tech lead —
     Claude answers what it can as the ticket's author, or says "that's yours to decide,
     state an assumption" for genuinely open questions.
   - **Step B — the design round**, the mentor skill's Phase 3b prediction-feedback loop:
     he proposes a design with numbered assumptions/invariants; Claude scores it against
     four bars with an explicit PASS/FAIL each (correct beyond the ACs, right-sized for
     the stated scale, simplest construction, no anti-patterns with a future cost); names
     the gap and a concrete failing input; **never supplies the fix**. Iterate fast — his
     early designs being wrong is fine and expected, the loop is the point. Once
     converged, write it up as `docs/adr/000X-<issue-name>.md`.
2. **Implementation mode is dynamic — chosen fresh each time, switchable mid-unit, never
   locked in at kickoff.**
   - Default for this project: **guided build** — Claude never writes a piece of code
     unless explicitly asked to for that specific piece. Instead it gives the next
     concrete step, he writes it, Claude checks it immediately, repeat, all the way
     through the unit.
   - He can ask Claude to just write one specific piece directly (a one-off jump to
     Rung-1-style for that piece only) and then resume guided build right after — follow
     that switch instantly, don't hold out for "the right rung."
3. **No Claude attribution in any commit, full stop.** This is a resume-facing personal
   project. No `Co-Authored-By: Claude` trailer or any other Claude/Anthropic mention, in
   any commit message, including ones for docs Claude drafted.
4. **Deployment is first-class scope, not an afterthought.** Issue #6 — containerize and
   deploy to a real Kubernetes cluster (kind/minikube is acceptable). One step past
   mrp-go, which only reached Docker.
5. **Resume harvest is a tracked deliverable.** `docs/RESUME.md` maps issues to target
   resume claims. Update it — draft/refine bullets, link evidence — whenever an issue
   closes. Formula: past-tense verb, exactly 3 capitalized keywords right after the verb,
   30-45 words, real measured numbers only, honest attribution (personal project, never a
   claimed production outcome).
6. **Race detector:** same machine constraint as `go-concurrency` — native `-race` is
   broken here (no cgo). Once there's code worth race-testing, set up a Docker-wrapped
   `test-race.ps1`, same pattern as `go-concurrency/test-race.ps1`.

## Scope — six GitHub issues (`RakshithYadhav/webhook-go`)

1. [#1](https://github.com/RakshithYadhav/webhook-go/issues/1) Accept and route an event for delivery (intake plumbing)
2. [#2](https://github.com/RakshithYadhav/webhook-go/issues/2) Deliver events via a bounded worker pool (core pillar)
3. [#3](https://github.com/RakshithYadhav/webhook-go/issues/3) Guarantee per-resource delivery ordering (core pillar — keyed mutex)
4. [#4](https://github.com/RakshithYadhav/webhook-go/issues/4) Graceful shutdown with in-flight drain (core pillar)
5. [#5](https://github.com/RakshithYadhav/webhook-go/issues/5) Minimal failure visibility, no retry/backoff (deliberately thin)
6. [#6](https://github.com/RakshithYadhav/webhook-go/issues/6) Containerize and deploy to Kubernetes (standing deployment rule)

**Non-goals for v1:** retry/backoff logic (that's `05-api-quota`'s territory in
go-concurrency — duplicating it here is redundant, not additive), a persistent
datastore, auth/webhook-signing/a multi-tenant dashboard, distributed/horizontal scaling
(punted to the separate system-design practice track below instead).

**IMPORTANT — planned v2:** v1's event queue is an in-process Go channel, deliberately
(issue #1 AC: "no external event bus required for v1") — this capstone is about Go's own
concurrency primitives, not operating a broker. Once `BACKEND-MASTERY.md` Track 8
(Distributed Systems & Microservices) has taught Kafka/RabbitMQ, **this project is the
planned target for a v2** that replaces the Go channel with a real message broker — that's
where queue-in-production practice happens, not here in v1.

## Separate, optional track

`docs/system-design-practice.md` — a zoomed-out system-design exercise on the same
domain (webhook delivery) at production scale (~50k customers, ~5k events/sec), entirely
separate from this project's actual small-scale implementation. Not required to proceed
with the real capstone; works the same prediction-feedback loop, scored against
system-design bars instead of the LLD four bars.

## Roadmap state

Docs only so far: this file, `docs/system-design-practice.md`, `docs/RESUME.md`. No
`go.mod`, no code yet. Next: the §1-2 pass and design round on issue #1.
