# 10. Support text, JSON, and CSV output through sink adapters

Date: 2026-07-11

## Status

Accepted

## Context

Different consumers want different shapes: humans want readable text in a
terminal, CI pipelines want JSON, spreadsheets want CSV (all documented in
the README's "Output Formats"). Encoding and I/O must not leak into domain
code (ADR 0004), and adding a format should not disturb existing ones.

## Decision

Rendering is its own feature slice. `internal/features/reporting` defines
the `Format` domain type and an outbound sink port; its application service
(`Write(report, format, sink)`) renders a `Report` to the chosen format.
`internal/infrastructure/sinks` implements the port for stdout and files.
The CLI selects the format with a flag and picks text automatically when
stdout is a terminal.

## Consequences

New output formats are added inside the reporting slice plus a flag value —
no change to analysis code. Encoders (`encoding/json`, `encoding/csv`) stay
in the application/infrastructure layers where the purity guard allows them.
Each format is one more surface that the determinism contract of ADR 0009
covers.
