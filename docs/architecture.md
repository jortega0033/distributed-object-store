# Implementer map

Status: **planned, not implemented.** This page maps the boundaries that [PROJECT-PLAN.md](PROJECT-PLAN.md) requires onto a module layout so that parallel tickets do not collide. It adds no requirement; the plan wins where they differ. The names are the orchestrator's proposal, fixed by the scaffold; a ticket that changes one corrects this page in the same pull request.

## Planned module map

| Boundary | Module | Owns | Parent ticket |
|---|---|---|---|
| Entry point | `cmd/objstore` | Subcommands `serve`, `node`, `cluster`, `verify` | #2 (`serve`), #4 (`node`, `cluster`), #7 (`verify`) |
| Codec | `internal/codec` | Field arithmetic, stripe encode and recover | #3 |
| Store interface and development adapter | `internal/store` | The interface the coordinator uses, plus the single-node development adapter with its own format | #2 |
| Manifest and head | `internal/manifest` | Immutable manifests, atomic head, version allocation, idempotency lookup | #4 |
| Node | `internal/node` | Durable shard files, node HTTP, quota accounting | #4 |
| Coordinator | `internal/coordinator` | Write, degraded read and repair orchestration, deadlines, reservations | #4, #5, #6 |
| HTTP | `internal/httpapi` | Routes, guards, ranges, error shape, static assets | #2, #5 |
| Browser assets | `web/ (embedded)` | Upload, list, inspect, status, repair | #2, #6 |
| Recovery matrix | `internal/matrix` | All-pairs loss, restart and fault runs | #7 |

Dependency direction: `httpapi` depends on `coordinator`; `coordinator` on `store`, `manifest`, `codec` and a node client; `node` on nothing from the coordinator side. `codec` imports nothing from the module.

## Freeze points

Parallel tickets rely on these shapes. A ticket that changes a frozen shape updates this page and every dependent ticket in the same pull request.

- **Store interface** (put, get, list, inspect) and **codec stripe interface**: agreed in #2 and #3 before either builds on the other.
- **Manifest schema version 1** and the head file: #4; #5 and #6 read it.
- **HTTP shapes**: object, version, inspect and error JSON (#2), per-role status and repair report (#6).

## What is verified where

Go is verified in CI only (see [CLAUDE.md](../CLAUDE.md)), on Linux, the supported platform. The durable write sequence is tested on the CI runner's local disk; that is a controlled process fault model, not power loss. Manual browser steps are owner checks in WSL or on Linux.
