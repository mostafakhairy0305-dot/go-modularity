# 3. Expose a single public facade package

Date: 2026-07-11

## Status

Accepted

## Context

The analyzer is consumed two ways: as a CLI and as an importable library.
Library consumers need a stable, documented API, but the implementation —
pipelines, loaders, metric packages — needs the freedom to change shape
without breaking anyone.

## Decision

The root package `gomodularity` is the only public API. It exposes
`Analyze(ctx context.Context, config Config) (Report, error)` plus the
`Config` and `Report` types (`gomodularity.go`, `config.go`, `report.go`).
Every implementation package lives under `internal/`, which the Go compiler
makes unimportable from outside the module.

## Consequences

Internals can be refactored freely — the ports, features, and adapters of
ADR 0002 are invisible to consumers. There is exactly one entry point to
document and keep backward compatible. The flip side is deliberate: nothing
inside `internal/` can be reused by third parties, and any capability worth
exposing must be surfaced through the facade's `Config` and `Report`.
