package domain

import (
	"strings"
	"testing"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

func TestRelPathEdges(t *testing.T) {
	if got := relPath("example.com/m/p", ""); got != "example.com/m/p" {
		t.Fatalf("empty module: got %q", got)
	}
	if got := relPath("other.com/x", "example.com/m"); got != "other.com/x" {
		t.Fatalf("outside module: got %q", got)
	}
}

func TestTextEmptyPackages(t *testing.T) {
	got := Text(gomodularity.Report{
		SchemaVersion: "1",
		Tool:          gomodularity.ToolInfo{Name: "go-modularity", Version: "test"},
		Module:        "example.com/m",
	}, TextOptions{})
	if !strings.Contains(got, "module example.com/m") || strings.Contains(got, "PATH / TYPE") {
		t.Fatalf("empty packages output unexpected:\n%s", got)
	}
}

func TestTextMultiSectionSpacerAndMissingMetrics(t *testing.T) {
	report := gomodularity.Report{
		SchemaVersion: "1",
		Tool:          gomodularity.ToolInfo{Name: "go-modularity", Version: "test"},
		Module:        "example.com/m",
		Packages: []gomodularity.PackageReport{
			{
				Path: "example.com/m",
				Metrics: []metrics.MetricResult{
					{
						Name:       metrics.MetricAbstractness,
						Scope:      metrics.ScopePackage,
						Value:      0.5,
						Applicable: true,
					},
					{
						Name:       metrics.MetricDistance,
						Scope:      metrics.ScopePackage,
						Value:      0.1,
						Applicable: true,
					},
				},
				// No types → typesTotal 0 on the root package row.
			},
			{
				Path: "example.com/m/leaf",
				Metrics: []metrics.MetricResult{
					// Missing distance (present elsewhere) → blank trailing cell.
					{
						Name:       metrics.MetricAbstractness,
						Scope:      metrics.ScopePackage,
						Applicable: false,
						Reason:     "isolated",
					},
				},
				Types: []gomodularity.TypeReport{{
					Name: "T",
					Metrics: []metrics.MetricResult{
						typeMetric(metrics.MetricAMC, 1),
						{Name: "custom", Scope: metrics.ScopeType, Value: 9, Applicable: true},
					},
				}},
			},
		},
	}

	got := Text(report, TextOptions{Color: true, Explain: true})
	if !strings.Contains(got, "\n\n") {
		t.Fatalf("expected blank spacer between sections:\n%s", got)
	}
	if !strings.Contains(got, "abstractness: isolated") {
		t.Fatalf("package metric reason missing:\n%s", got)
	}
	// Unknown metric name has no quality color.
	if strings.Contains(got, ansiGreen+"9.00") || strings.Contains(got, ansiRed+"9.00") {
		t.Fatalf("unknown metric was quality-colored:\n%q", got)
	}
}

func TestTextExplainAllTypesAndSkipEmptyNotes(t *testing.T) {
	report := gomodularity.Report{
		SchemaVersion: "1",
		Tool:          gomodularity.ToolInfo{Name: "go-modularity", Version: "test"},
		Module:        "example.com/m",
		Packages: []gomodularity.PackageReport{
			{
				Path: "example.com/m/quiet",
				Metrics: []metrics.MetricResult{
					{
						Name:       metrics.MetricAbstractness,
						Scope:      metrics.ScopePackage,
						Value:      0,
						Applicable: true,
					},
				},
			},
			{
				Path: "example.com/m/noisy",
				Types: []gomodularity.TypeReport{
					{Name: "A", Metrics: []metrics.MetricResult{
						{
							Name:       metrics.MetricTCC,
							Scope:      metrics.ScopeType,
							Applicable: false,
							Reason:     "fewer than two methods",
						},
					}},
					{Name: "B", Metrics: []metrics.MetricResult{
						{
							Name:       metrics.MetricTCC,
							Scope:      metrics.ScopeType,
							Applicable: false,
							Reason:     "fewer than two methods",
						},
					}},
				},
			},
		},
	}

	got := Text(report, TextOptions{Explain: true})
	if !strings.Contains(got, "tcc: fewer than two methods (all types)") {
		t.Fatalf("want all-types aggregation:\n%s", got)
	}
	notesIdx := strings.Index(got, "\nnotes\n")
	if notesIdx < 0 {
		t.Fatalf("notes section missing:\n%s", got)
	}
	if strings.Contains(got[notesIdx:], "quiet") {
		t.Fatalf("quiet package should be skipped in notes:\n%s", got[notesIdx:])
	}
}

func TestValueColorUnknownMetric(t *testing.T) {
	if got := valueColor("not-a-metric", 1, &columnStats{min: 0, max: 2, count: 2}); got != "" {
		t.Fatalf("valueColor = %q, want empty", got)
	}
}

func TestMeanCellNilStats(t *testing.T) {
	cell := meanCell(nil, func(float64) string { return "" })
	if cell.text != naCell {
		t.Fatalf("meanCell(nil) = %q, want %q", cell.text, naCell)
	}
}

func TestTextTrailingBlankPackageMetric(t *testing.T) {
	// Package-only columns: one package lacks the last metric so the blank
	// trailing cell is trimmed when the row is written.
	report := gomodularity.Report{
		SchemaVersion: "1",
		Tool:          gomodularity.ToolInfo{Name: "go-modularity", Version: "test"},
		Module:        "example.com/m",
		Packages: []gomodularity.PackageReport{
			{
				Path: "example.com/m/a",
				Metrics: []metrics.MetricResult{
					{
						Name:       metrics.MetricAbstractness,
						Scope:      metrics.ScopePackage,
						Value:      1,
						Applicable: true,
					},
					{
						Name:       metrics.MetricDistance,
						Scope:      metrics.ScopePackage,
						Value:      0,
						Applicable: true,
					},
				},
			},
			{
				Path: "example.com/m/b",
				Metrics: []metrics.MetricResult{
					{
						Name:       metrics.MetricAbstractness,
						Scope:      metrics.ScopePackage,
						Value:      0,
						Applicable: true,
					},
				},
			},
		},
	}
	got := Text(report, TextOptions{})
	mustMatch(t, got, `(?m)^b\s+0\.00$`)
}
