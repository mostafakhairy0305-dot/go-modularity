// Package gomodularity analyzes Go modules and reports type- and package-level
// modularity metrics.
//
// Call Analyze with a Config to load packages, compute the selected metrics,
// and receive a deterministic Report. Config selects patterns, metric set,
// field-usage mode, and dependency scope; Report carries per-package and
// per-type results in a fixed order.
//
// For policy enforcement via go/analysis, use the sibling analyzer package.
// To register that analyzer as a golangci-lint Module Plugin, blank-import
// the sibling plugin package.
package gomodularity
