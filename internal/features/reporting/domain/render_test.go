package domain

import (
	"regexp"
	"strings"
	"testing"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity/gomodularity"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

func typeMetric(name string, value float64) metrics.MetricResult {
	return metrics.MetricResult{
		Name:       name,
		Scope:      metrics.ScopeType,
		Value:      value,
		Applicable: true,
	}
}

func tableReport() gomodularity.Report {
	return gomodularity.Report{
		SchemaVersion: "1",
		Tool:          gomodularity.ToolInfo{Name: "go-modularity", Version: "test"},
		Module:        "example.com/mod",
		Packages: []gomodularity.PackageReport{{
			Path: "example.com/mod",
			Metrics: []metrics.MetricResult{
				{
					Name:       metrics.MetricAbstractness,
					Scope:      metrics.ScopePackage,
					Value:      0.25,
					Applicable: true,
				},
				{
					Name:       metrics.MetricDistance,
					Scope:      metrics.ScopePackage,
					Value:      0.15,
					Applicable: true,
				},
			},
			Types: []gomodularity.TypeReport{
				{Name: "Cart", Metrics: []metrics.MetricResult{
					typeMetric(metrics.MetricAMC, 2),
					typeMetric(metrics.MetricTCC, 0.75),
				}},
				{Name: "Order", Metrics: []metrics.MetricResult{
					typeMetric(metrics.MetricAMC, 6),
					{
						Name:       metrics.MetricTCC,
						Scope:      metrics.ScopeType,
						Applicable: false,
						Reason:     "fewer than two methods",
					},
				}},
			},
		}},
	}
}

func mustMatch(t *testing.T, got, pattern string) {
	t.Helper()

	if !regexp.MustCompile(pattern).MatchString(got) {
		t.Errorf("output does not match %q\ngot:\n%s", pattern, got)
	}
}

func TestTextTreeTableLayout(t *testing.T) {
	got := Text(tableReport(), TextOptions{})

	mustMatch(t, got, `(?m)^module example\.com/mod$`)
	mustMatch(t, got, `(?m)^PATH / TYPE\s+Abst\s+Dist\s+AMC\s+TCC$`)
	// The module-root package renders as "." carrying its package metrics
	// plus the means of its types' metrics (0.75 averages Cart only).
	mustMatch(t, got, `(?m)^\.\s+0\.25\s+0\.15\s+4\.00\s+0\.75$`)
	mustMatch(t, got, `(?m)^├── Cart\s+2\.00\s+0\.75$`)
	mustMatch(t, got, `(?m)^└── Order\s+6\.00\s+–$`)
	mustMatch(t, got, `(?m)^– = not applicable$`)

	if strings.Contains(got, "mean") {
		t.Errorf("output still contains a separate mean row:\n%s", got)
	}

	if strings.Contains(got, "\x1b[") {
		t.Errorf("uncolored output contains ANSI escapes:\n%q", got)
	}
}

func TestTextTreeGroupsPackagesUnderSharedPath(t *testing.T) {
	report := tableReport()
	report.Packages = []gomodularity.PackageReport{
		{
			Path: "example.com/mod/internal/a",
			Metrics: []metrics.MetricResult{
				{
					Name:       metrics.MetricAbstractness,
					Scope:      metrics.ScopePackage,
					Value:      1,
					Applicable: true,
				},
			},
			Types: []gomodularity.TypeReport{
				{Name: "T1", Metrics: []metrics.MetricResult{typeMetric(metrics.MetricTCC, 0.5)}},
			},
		},
		{
			Path: "example.com/mod/internal/b/deep",
			Metrics: []metrics.MetricResult{
				{
					Name:       metrics.MetricAbstractness,
					Scope:      metrics.ScopePackage,
					Value:      0,
					Applicable: true,
				},
			},
			Types: []gomodularity.TypeReport{
				{Name: "T2", Metrics: []metrics.MetricResult{typeMetric(metrics.MetricTCC, 1)}},
			},
		},
	}

	got := Text(report, TextOptions{})

	// Shared "internal" directory heads the section and aggregates all
	// packages beneath it: A mean (1+0)/2, TCC mean (0.5+1)/2. The
	// single-child chain b → deep is compressed into one branch.
	mustMatch(t, got, `(?m)^internal\s+0\.50\s+0\.75$`)
	mustMatch(t, got, `(?m)^├── a\s+1\.00\s+0\.50$`)
	mustMatch(t, got, `(?m)^│   └── T1\s+0\.50$`)
	mustMatch(t, got, `(?m)^│$`)
	mustMatch(t, got, `(?m)^└── b/deep\s+0\.00\s+1\.00$`)
	mustMatch(t, got, `(?m)^    └── T2\s+1\.00$`)
}

func TestTextParentPackageGroupSkipsNonApplicableMetrics(t *testing.T) {
	report := tableReport()
	report.Packages = []gomodularity.PackageReport{
		{
			Path: "example.com/mod/group",
			Metrics: []metrics.MetricResult{
				{Name: metrics.MetricAbstractness, Scope: metrics.ScopePackage},
				{
					Name:       metrics.MetricInstability,
					Scope:      metrics.ScopePackage,
					Value:      0.2,
					Applicable: true,
				},
				{Name: metrics.MetricDistance, Scope: metrics.ScopePackage},
			},
		},
		{
			Path: "example.com/mod/group/child",
			Metrics: []metrics.MetricResult{
				{
					Name:       metrics.MetricAbstractness,
					Scope:      metrics.ScopePackage,
					Value:      0.6,
					Applicable: true,
				},
				{
					Name:       metrics.MetricInstability,
					Scope:      metrics.ScopePackage,
					Value:      0.8,
					Applicable: true,
				},
				{
					Name:       metrics.MetricDistance,
					Scope:      metrics.ScopePackage,
					Value:      0.4,
					Applicable: true,
				},
			},
		},
	}

	got := Text(report, TextOptions{})

	// The parent package is also a group. Its n/a abstractness and distance
	// are skipped, while instability averages both applicable package values.
	mustMatch(t, got, `(?m)^group\s+0\.60\s+0\.50\s+0\.40$`)
	mustMatch(t, got, `(?m)^└── child\s+0\.60\s+0\.80\s+0\.40$`)
}

func TestTextModuleRootSummarizesApplicablePackageMetrics(t *testing.T) {
	report := tableReport()
	report.Packages = []gomodularity.PackageReport{
		{
			Path: "example.com/mod",
			Metrics: []metrics.MetricResult{
				{Name: metrics.MetricAbstractness, Scope: metrics.ScopePackage},
				{
					Name:       metrics.MetricInstability,
					Scope:      metrics.ScopePackage,
					Value:      1,
					Applicable: true,
				},
				{Name: metrics.MetricDistance, Scope: metrics.ScopePackage},
			},
		},
		{
			Path: "example.com/mod/child",
			Metrics: []metrics.MetricResult{
				{
					Name:       metrics.MetricAbstractness,
					Scope:      metrics.ScopePackage,
					Value:      0.6,
					Applicable: true,
				},
				{
					Name:       metrics.MetricInstability,
					Scope:      metrics.ScopePackage,
					Value:      0.5,
					Applicable: true,
				},
				{
					Name:       metrics.MetricDistance,
					Scope:      metrics.ScopePackage,
					Value:      0.4,
					Applicable: true,
				},
			},
		},
	}

	got := Text(report, TextOptions{})

	// Root n/a metrics are skipped. Instability averages the applicable root
	// and child values, while abstractness and distance use the child value.
	mustMatch(t, got, `(?m)^\.\s+0\.60\s+0\.75\s+0\.40$`)
	mustMatch(t, got, `(?m)^child\s+0\.60\s+0\.50\s+0\.40$`)
}

func TestTextReasonsOnlyWithExplain(t *testing.T) {
	report := tableReport()

	if got := Text(report, TextOptions{}); strings.Contains(got, "fewer than two methods") {
		t.Errorf("reasons shown without Explain:\n%s", got)
	}

	got := Text(report, TextOptions{Explain: true})

	wantLines := []string{
		"notes",
		"  example.com/mod",
		"    tcc: fewer than two methods (Order)",
	}
	for _, line := range wantLines {
		if !strings.Contains(got, line+"\n") {
			t.Errorf("explain output missing line %q\ngot:\n%s", line, got)
		}
	}
}

func TestTextMeanSkipsNonApplicable(t *testing.T) {
	got := Text(tableReport(), TextOptions{})

	// The TCC mean averages only Cart's 0.75; Order's n/a must not drag
	// it down to 0.375.
	if strings.Contains(got, "0.38") || strings.Contains(got, "0.37") {
		t.Errorf("mean included a non-applicable value:\n%s", got)
	}
}

func TestTextColorAppliesQualityAndBold(t *testing.T) {
	got := Text(tableReport(), TextOptions{Color: true})

	if !strings.Contains(got, ansiGreen+"0.75"+ansiReset) {
		t.Errorf("high TCC not green:\n%q", got)
	}
	// AMC is unbounded and lower-better: 2 is the column best, 6 the worst.
	if !strings.Contains(got, ansiGreen+"2.00"+ansiReset) {
		t.Errorf("best AMC not green:\n%q", got)
	}

	if !strings.Contains(got, ansiRed+"6.00"+ansiReset) {
		t.Errorf("worst AMC not red:\n%q", got)
	}
	// Aggregated means on the package row render bold on top of their
	// quality color.
	if !strings.Contains(got, ansiBold+ansiGreen+"0.75"+ansiReset) {
		t.Errorf("package-row TCC mean not bold green:\n%q", got)
	}
	// Distance 0.15 is bounded lower-better: green on the package row.
	if !strings.Contains(got, ansiGreen+"0.15"+ansiReset) {
		t.Errorf("low distance not green:\n%q", got)
	}
	// Abstractness has no quality direction: never colored.
	if strings.Contains(got, ansiGreen+"0.25") || strings.Contains(got, ansiRed+"0.25") ||
		strings.Contains(got, ansiYellow+"0.25") {
		t.Errorf("abstractness was quality-colored:\n%q", got)
	}
	// Tree glyphs stay unstyled so the branches read cleanly.
	if !strings.Contains(got, "├── Cart") {
		t.Errorf("tree glyph missing or styled:\n%q", got)
	}
}

func TestTextSingleTypeLeavesUnboundedPlain(t *testing.T) {
	report := tableReport()
	report.Packages[0].Types = report.Packages[0].Types[:1]

	got := Text(report, TextOptions{Color: true})
	if strings.Contains(got, ansiGreen+"2.00"+ansiReset) ||
		strings.Contains(got, ansiRed+"2.00"+ansiReset) ||
		strings.Contains(got, ansiYellow+"2.00"+ansiReset) {
		t.Errorf("lone AMC value was relatively colored:\n%q", got)
	}
}

func TestFormatCell(t *testing.T) {
	cases := map[float64]string{
		12:        "12.00",
		0:         "0.00",
		4.25:      "4.25",
		2.0 / 3.0: "0.67",
		0.5:       "0.50",
	}
	for value, want := range cases {
		if got := formatCell(value); got != want {
			t.Errorf("formatCell(%v) = %q, want %q", value, got, want)
		}
	}
}
