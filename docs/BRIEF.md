# Webhook Delivery Service — Project Brief

| | |
|---|---|
| Status | Approved (lightweight) |
| Sponsor | Head of Platform |
| Engineering owner | Rakshith (design + delivery) |
| Target | Core pillars (worker pool, keyed ordering, graceful shutdown) + Kubernetes deploy |

This is a lightweight brief, not the full template — scope is tracked as GitHub issues
(`RakshithYadhav/webhook-go`) rather than numbered FRs, to move fast. Sections 1 and 2
are never skipped regardless of how lightweight the rest of the kickoff is.

## 1. Background

We sell a production-scheduling SaaS product to manufacturing companies (the same
product behind OPS-1743 and OPS-1809). Several enterprise customers have asked for
real-time notifications when schedule-relevant events happen on their account — a work
order's status changes, a schedule conflict gets resolved, a due date slips — so their
own ERP or dashboard systems can react immediately instead of polling our API. Two
customers have already built their own polling workarounds; one is hammering our API
every 30 seconds across their whole account, and both have missed fast-moving changes
that happened between polls. Sales has flagged "real-time webhooks" as a checkbox item
in two enterprise deals currently in procurement.

## 2. Problem statement

There is no way today for a customer's system to be notified when a schedule-relevant
event happens on their account — they must poll, which is unreliable (events can be
missed between polls) and puts unnecessary load on our API. We need a webhook delivery
system that reliably delivers each event to every endpoint a customer has registered,
in the order those events happened for a given resource, even when a customer's
endpoint is slow or occasionally unavailable.

## Scope

Tracked as GitHub issues, not a numbered FR list:
[#1](https://github.com/RakshithYadhav/webhook-go/issues/1)-[#6](https://github.com/RakshithYadhav/webhook-go/issues/6).
Non-goals recorded in `CLAUDE.md`.
