# 18. Enforce a modularity policy gate

Date: 2026-07-13

## Status

Accepted

## Context

The tool computed metrics but could not act on them: every successful run
exited `0`, so keeping a codebase within a budget meant post-processing JSON in
each project's CI. ADR 0011 anticipated this — it reserved a distinct exit code
to separate "analysis found violations" from "the tool itself failed" — but the
violations path was never built.

## Decision

Add a policy gate. A `Policy` is a set of conditions — `max`/`min` bounds on
every structural fact (types, exported/unexported funcs, coupling, fields,
methods) and every metric. Metric bounds may be scoped to package metrics or
type metrics so a package metric such as `distance` is not confused with a type
metric such as `reusability`; legacy flat metric maps remain accepted as global
bounds. `Evaluate(report, policy)` is a pure domain function returning ordered
violations; it skips conditions on metrics that are not applicable to an entity
so n/a cells never fail a build. Conditions come from a `.modularity.yml` file
(ADR 0017), from repeatable `-max`/`-min` flags, or from a recommended strict
default, layered defaults → file → flags.

The gate is opt-in and explicit: it runs only when `-check`, `-config`, or an
override flag is given, so a committed config never changes an unrelated run.
Following ADR 0011, the CLI writes the report to stdout, prints the violation
summary to stderr, and exits `3`. Policy configuration errors are usage errors
(`2`).

## Consequences

The tool is now a CI gate with stable, documented exit semantics. The policy
model is a pure feature slice (`internal/features/policy/domain`) wired by the
thin CLI, exactly as reporting is (ADR 0003), so the facade stays the minimal
`Analyze → Report` library. Gated metrics are auto-added to the display set,
which means enabling a condition can add a column to the report. The strict
defaults will flag many existing codebases on first run; they are a starting
point to tighten toward, not a claim that every project should pass immediately.
