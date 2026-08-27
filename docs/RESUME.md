# Resume Harvest — webhook-go

This project exists to substantiate specific resume claims. This file is the ledger:
which claims it targets, what evidence backs each, and the drafted bullets with their
interview dossiers. Rule: **a bullet without linked evidence doesn't get written.**

## The bullet formula (hard rule)

1. Past-tense leading verb ("Developed," "Optimized," "Implemented").
2. Exactly 3 capitalized keyword phrases, right after the verb ("Go, REST APIs, and Microservices").
3. How it was used — the concrete thing built or fixed.
4. Non-technical business reason — why it mattered, not just the engineering.
5. Quantified impact only if the number is real and measured. No number, no claim.
6. One sentence, 30-45 words.
7. Honest attribution — never claim outcomes not personally driven.

**Honesty boundary:** this is a personal project, built to practice and demonstrate the
concurrency patterns behind a real production concern (reliable webhook delivery), not a
system that served real customers. Its bullets demonstrate *mechanisms* — how to bound
concurrency, how to guarantee per-key ordering, how to shut down without dropping work —
never business-outcome numbers (revenue, uptime SLAs, org-level percentages); those belong
to real jobs, never here. If asked directly "was this in production," the honest answer is
"no — my own project, built to go deep on patterns my day job doesn't require me to build
from scratch."

## Target claims → issue map

| Issue | Target claim (mechanism) | Keywords | Evidence | Status |
|-------|---------------------------|----------|----------|--------|
| [#2](https://github.com/RakshithYadhav/webhook-go/issues/2) Bounded worker pool | Concurrent delivery bounded by a worker pool, not unbounded fan-out | Go, Worker Pools, Bounded Concurrency | — | not started |
| [#3](https://github.com/RakshithYadhav/webhook-go/issues/3) Per-resource ordering | Keyed serialization guarantees in-order delivery per resource despite concurrent workers | Concurrency Control, Keyed Locking, Race Conditions | — | not started |
| [#4](https://github.com/RakshithYadhav/webhook-go/issues/4) Graceful shutdown | Deadline-bounded drain on shutdown, zero dropped in-flight work | Graceful Shutdown, Context Cancellation, Zero-Downtime Deploys | — | not started |
| [#6](https://github.com/RakshithYadhav/webhook-go/issues/6) Docker + Kubernetes | Containerized service deployed and verified on a real Kubernetes cluster | Docker, Kubernetes, Container Orchestration | — | not started |

Issues #1 (intake) and #5 (failure logging) are plumbing/support work for the above —
they don't carry their own bullet unless something in their implementation turns out to be
non-obvious enough to be worth one.

## Drafted bullets

(Drafted here as each issue closes, following the formula above — full interview dossier
per bullet, same shape as mrp-go's `docs/RESUME.md`.)

## Harvest log

- 2026-08-27 — file created, target-claim map drafted against the six open issues.
