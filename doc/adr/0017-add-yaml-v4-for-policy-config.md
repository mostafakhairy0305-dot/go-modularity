# 17. Add go.yaml.in/yaml/v4 for policy config

Date: 2026-07-13

## Status

Accepted

## Context

The policy gate (ADR 0018) reads its conditions from a `.modularity.yml` file.
YAML is what developers expect for tool configuration — it sits beside
`.golangci.yml` and the project's other config — but the standard library has
no YAML decoder. ADR 0006 keeps the module to a single direct dependency and
requires a strong justification, and ideally its own ADR, before adding
another.

Hand-rolling a YAML subset was considered and rejected. This file is consumed
in other people's CI, where a parser that silently mis-reads valid YAML — the
kind of edge case a subset misses — is worse than an extra dependency: it would
turn a passing gate into a false green. Correct decoding, unknown-key
detection, and precise line numbers in errors are exactly what a mature library
already provides.

## Decision

Add `go.yaml.in/yaml/v4` as the second direct dependency. This is the canonical
home of the Go YAML library (the successor to `gopkg.in/yaml.v3`); v4 is chosen
over the legacy import path so the project tracks the maintained line. Its use
is quarantined to one infrastructure adapter,
`internal/infrastructure/policyconfig`, which maps the decoded document onto the
pure policy domain. The domain, the facade, and every other package remain
yaml-free; the architecture guard (ADR 0004) keeps `policy/domain` pure, so the
dependency cannot leak inward.

## Consequences

`go.mod` now carries two direct dependencies instead of one. In exchange, the
config loader is small and the schema is parsed correctly, with `KnownFields`
rejecting typos and errors carrying line numbers. Because the import is
isolated behind a hand-written DTO, the dependency can be swapped or removed
without touching the policy model. Any further dependency still needs the same
justification bar.
