# Distributed Object Store

A local six-node object storage demo with 4+2 erasure coding, versioned reads and repair.

Only the scaffold exists (layout, stub commands, smoke tests, CI); the implementation is planned. The scope, architecture and acceptance criteria are in [the project plan](docs/PROJECT-PLAN.md). The plan includes a reproducible verification strategy for the future build.

- [Product requirements](docs/PROJECT-PLAN.md#product-requirements)
- [Technical design](docs/PROJECT-PLAN.md#technical-design)
- [Implementation backlog and epic](https://github.com/jortega0033/distributed-object-store/issues)

Estimated implementation and verification effort is 21–33 engineering hours.

## Program goal

Build a production-grade distributed object store in Go: consensus-backed metadata, erasure-coded storage, an S3-compatible interface, multi-site replication and online change of everything, proven under simulated faults, with no storage or consensus library.

The program is planned ([#39](https://github.com/jortega0033/distributed-object-store/issues/39)). Phase 1 is the current epic [#1](https://github.com/jortega0033/distributed-object-store/issues/1) and the only phase with tickets.

**Score.** Share of a pinned, independent S3 compatibility test suite that passes, next to the number of seeded fault runs without a lost object or a consistency violation.

| Phase | Content | Hours (rough) | Issue |
|---|---|---|---|
| 1 | The current epic and its tickets | 76 to 102 | [#1](https://github.com/jortega0033/distributed-object-store/issues/1) |
| 2 | Consensus metadata | 180 to 300 | [#40](https://github.com/jortega0033/distributed-object-store/issues/40) |
| 3 | S3 interface | 140 to 240 | [#41](https://github.com/jortega0033/distributed-object-store/issues/41) |
| 4 | Storage engine | 140 to 240 | [#42](https://github.com/jortega0033/distributed-object-store/issues/42) |
| 5 | Scale and change | 130 to 220 | [#43](https://github.com/jortega0033/distributed-object-store/issues/43) |
| 6 | Proof | 140 to 230 | [#44](https://github.com/jortega0033/distributed-object-store/issues/44) |

About 810 to 1330 engineer hours in total: rough estimates, not measurements. The numbers in the score and the finish lines are targets until a run shows them. Only phase 1 has tickets to work; the sub-issues of a later phase are written when it starts. No single session is expected to reach the end.

## License

[MIT](LICENSE) © 2026 Jake Ortega.
