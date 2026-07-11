package gomodularity

import "github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"

// SchemaVersion is the version of the report schema produced by Analyze.
// Version 2 added the structural facts: afferent/efferent coupling and
// function counts per package, field and method counts per type.
const SchemaVersion = "2"

// ToolName is the canonical tool name embedded in reports.
const ToolName = "go-modularity"

// MetricResult aliases the metrics package's result type; see its
// documentation for the applicability contract.
type MetricResult = metrics.MetricResult

// ToolInfo identifies the tool that produced a report.
type ToolInfo struct {
	Name    string
	Version string
}

// Report is the complete, deterministic result of one analysis run.
// Packages are sorted by import path; ordering never depends on map
// iteration.
type Report struct {
	SchemaVersion string
	Tool          ToolInfo
	// Module is the analyzed main module's path, when known. Renderers
	// use it to show package paths relative to the module root.
	Module   string
	Packages []PackageReport
}

// PackageReport carries one package's structural facts, its metrics, and
// its analyzed types. Metrics appear in the fixed metric order and contain
// only the selected display set. Types are sorted by name.
type PackageReport struct {
	Path string
	// Afferent counts analyzed packages importing this package (Ca).
	Afferent int
	// Efferent counts this package's in-scope imports (Ce), honoring the
	// configured dependency scope.
	Efferent int
	// Funcs counts the package's declared functions and methods.
	Funcs   int
	Metrics []MetricResult
	Types   []TypeReport
}

// TypeReport carries one named type's structural facts and its metrics in
// the fixed metric order, restricted to the selected display set.
type TypeReport struct {
	Name string
	// Fields is the struct field count (embedded fields count one).
	Fields int
	// Methods is the declared method count.
	Methods int
	Metrics []MetricResult
}
