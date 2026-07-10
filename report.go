package gomodularity

import "github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"

// SchemaVersion is the version of the report schema produced by Analyze.
const SchemaVersion = "1"

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

// PackageReport carries one package's metrics and its analyzed types.
// Metrics appear in the fixed metric order and contain only the selected
// display set. Types are sorted by name.
type PackageReport struct {
	Path    string
	Metrics []MetricResult
	Types   []TypeReport
}

// TypeReport carries one named type's metrics in the fixed metric order,
// restricted to the selected display set.
type TypeReport struct {
	Name    string
	Metrics []MetricResult
}
