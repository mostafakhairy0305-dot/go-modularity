# 19. Dual-mode: CLI library and golangci-lint Module Plugin

Date: 2026-07-17

## Status

Accepted

## Context

`go-modularity` already ships as a standalone CLI and library facade
(`gomodularity.Analyze`, ADR 0003) with a policy gate (ADR 0018). Teams that
standardize on golangci-lint want the same policy enforcement inside their
existing lint binary, without giving up the independent CLI for reports and
CI exit code `3`.

golangci-lint's recommended extension path is the [Module Plugin
System](https://golangci-lint.run/docs/plugins/module-plugins/), which loads
plugins that implement `register.LinterPlugin` from
`github.com/golangci/plugin-module-register` and return `go/analysis`
analyzers. A naive per-package Pass rewrite cannot compute coupling metrics
(Ca/Ce, abstractness, instability, distance): those need a whole-module load.

ADR 0006 keeps dependencies minimal and requires an ADR for new direct deps.

## Decision

Keep the hexagonal pipeline and CLI unchanged. Add an adapter layer:

1. Public package `analyzer` — builds a `go/analysis.Analyzer` that runs
   `gomodularity.Analyze` once (`sync.Once`) over configured patterns,
   evaluates the modularity policy, and emits diagnostics for the current
   package's violations. Positions come from the Pass AST (no report schema
   change). This package does **not** import plugin-module-register.
2. Public package `plugin` — registers `gomodularity` via
   `register.Plugin`, decodes custom settings into `analyzer.Settings`, and
   returns `LoadModeTypesInfo`.
3. Add `github.com/golangci/plugin-module-register` as a direct dependency,
   quarantined to `plugin/` (same pattern as yaml in ADR 0017).

Consumers build a custom binary with `.custom-gcl.yml` pointing at this
module's `plugin` import, then enable `gomodularity` under
`linters.settings.custom`.

## Consequences

One analysis core serves both modes. The Module Plugin dependency is present
in `go.mod` but unused unless a consumer imports `plugin`. Coupling metrics
remain correct because the adapter still performs a whole-module analyze.
Amend ADR 0003 to allow the two lint packages alongside the metrics facade.
