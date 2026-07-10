package application

import (
	"encoding/json"
	"io"
	"strings"
	"testing"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/reporting/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

func sampleReport() gomodularity.Report {
	applicable := metrics.MetricResult{Name: "abstractness", Scope: metrics.ScopePackage, Value: 0.5, Applicable: true, Definition: "d"}
	na := metrics.MetricResult{Name: "amc", Scope: metrics.ScopeType, Applicable: false, Reason: "no methods", Definition: "d"}
	return gomodularity.Report{
		SchemaVersion: "1",
		Tool:          gomodularity.ToolInfo{Name: "go-modularity", Version: "test"},
		Module:        "example.com/m",
		Packages: []gomodularity.PackageReport{{
			Path:    "example.com/m/a",
			Metrics: []metrics.MetricResult{applicable},
			Types:   []gomodularity.TypeReport{{Name: "A", Metrics: []metrics.MetricResult{na}}},
		}},
	}
}

// White-box: the JSON envelope round-trips and honors the applicability
// contract (applicable → value present; n/a → value omitted).
func TestRenderJSONContract(t *testing.T) {
	t.Parallel()
	var buf strings.Builder
	if err := renderJSON(&buf, sampleReport()); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	if err := json.Unmarshal([]byte(buf.String()), &got); err != nil {
		t.Fatalf("invalid JSON: %v\n%s", err, buf.String())
	}
	if got["schema_version"] != "1" {
		t.Errorf("schema_version = %v", got["schema_version"])
	}
	pkg := got["packages"].([]any)[0].(map[string]any)
	abst := pkg["metrics"].(map[string]any)["abstractness"].(map[string]any)
	if abst["value"].(float64) != 0.5 {
		t.Errorf("abstractness value = %v", abst["value"])
	}
	amc := pkg["types"].([]any)[0].(map[string]any)["metrics"].(map[string]any)["amc"].(map[string]any)
	if _, has := amc["value"]; has {
		t.Error("non-applicable metric must omit value")
	}
	if amc["applicable"] != false {
		t.Error("amc must be not applicable")
	}
}

// White-box: ordered metric objects keep the given slice order.
func TestEncodeOrderedMetricsPreservesOrder(t *testing.T) {
	t.Parallel()
	got, err := encodeOrderedMetrics([]metrics.MetricResult{
		{Name: "amc", Scope: metrics.ScopeType, Value: 1, Applicable: true, Definition: "d"},
		{Name: "tcc", Scope: metrics.ScopeType, Applicable: false, Reason: "x", Definition: "d"},
	})
	if err != nil {
		t.Fatal(err)
	}
	s := string(got)
	if !strings.HasPrefix(s, `{"amc":`) || strings.Index(s, "amc") > strings.Index(s, "tcc") {
		t.Errorf("order not preserved: %s", s)
	}
}

// White-box: an unknown format is rejected.
func TestRenderUnknownFormat(t *testing.T) {
	t.Parallel()
	if err := render(io.Discard, sampleReport(), domain.Format("xml"), domain.TextOptions{}); err == nil {
		t.Fatal("unknown format should error")
	}
}
