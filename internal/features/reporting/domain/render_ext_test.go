package domain_test

import (
	"strings"
	"testing"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
	reporting "github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

// Black-box: format parsing accepts the known encodings and rejects others.
func TestParseFormat(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"text", "json", "csv"} {
		if f, ok := reporting.ParseFormat(name); !ok || string(f) != name {
			t.Errorf("ParseFormat(%q) = %v,%v", name, f, ok)
		}
	}
	if _, ok := reporting.ParseFormat("xml"); ok {
		t.Error("xml must be rejected")
	}
}

// Black-box: the text and CSV renderers emit the module and a per-metric
// header/records.
func TestTextAndCSVRendering(t *testing.T) {
	t.Parallel()
	rep := gomodularity.Report{
		SchemaVersion: "1",
		Tool:          gomodularity.ToolInfo{Name: "go-modularity", Version: "t"},
		Module:        "example.com/m",
		Packages: []gomodularity.PackageReport{{
			Path:    "example.com/m/a",
			Metrics: []metrics.MetricResult{{Name: "abstractness", Scope: metrics.ScopePackage, Value: 0.5, Applicable: true}},
			Types: []gomodularity.TypeReport{{Name: "A", Metrics: []metrics.MetricResult{
				{Name: "amc", Scope: metrics.ScopeType, Value: 2, Applicable: true},
			}}},
		}},
	}

	text := reporting.Text(rep, reporting.TextOptions{})
	if !strings.Contains(text, "example.com/m") || !strings.Contains(text, "A") {
		t.Errorf("text output missing content:\n%s", text)
	}
	if len(reporting.CSVHeader()) == 0 {
		t.Error("empty CSV header")
	}
	if len(reporting.CSVRecords(rep)) == 0 {
		t.Error("no CSV records produced")
	}
	if reporting.FormatValue(0.5) == "" {
		t.Error("FormatValue produced empty string")
	}
}
