# System Design Practice — Webhook Delivery Platform

Separate from this project's actual implementation (see the eventual `BRIEF.md`). This
is an optional, zoomed-out system-design exercise on the same domain, at production scale
rather than personal-project scale. Not required to proceed with the real capstone.

## Protocol

1. Claude writes the problem statement below — deliberately open-ended, no functional
   requirements, no solution.
2. Rakshith derives the functional and non-functional requirements himself.
3. Rakshith proposes a high-level architecture (components, data flow, tech choices,
   capacity estimates, trade-offs).
4. Same prediction-feedback loop as the LLD design work: Claude scores against
   system-design bars (FR/NFR coverage, capacity estimation, bottleneck identification,
   availability/consistency trade-offs, articulated trade-offs) and hands back the gap and
   a concrete failing case — never the fix. Repeat until converged or he calls it.

## Problem statement

You're a senior backend engineer at a B2B SaaS platform that other companies integrate
with. Business events happen constantly — an order is created, a payment succeeds, an
invoice is finalized, a subscription is canceled. External developers who've integrated
with your platform want to be notified about these events in near-real-time, so they can
react without polling your API.

You're asked to design the webhook delivery system: whenever a business event occurs,
deliver it as an HTTP POST to every URL each customer has registered for that event type.

Rough scale — a starting point, not a spec. Sharpen, question, or push back on any of it
as part of your own requirements gathering, the way a real interview would let you:

- ~50,000 active customer accounts, each registering 1-20 webhook endpoints.
- ~5,000 business events/second at peak, platform-wide.
- Customer endpoints are unreliable — some are slow, some go down for minutes or hours at
  a time, some return errors intermittently.
- Customers care about eventually receiving every event correctly more than instant
  delivery — but they do care about ordering for events on the same resource (e.g., never
  seeing "order shipped" before "order created" for the same order).

Design the system.

## Rounds

(Log each round here as it happens: his FRs/NFRs and design, Claude's four-bar
PASS/FAIL and findings, what changed next round.)
