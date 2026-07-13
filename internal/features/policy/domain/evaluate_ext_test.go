package domain_test

import (
	"strings"
	"testing"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/policy/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

func typeMetric(name string, value float64) metrics.MetricResult {
	return metrics.MetricResult{Name: name, Scope: metrics.ScopeType, Value: value, Applicable: true}
}

func naMetric(name string) metrics.MetricResult {
	return metrics.MetricResult{Name: name, Scope: metrics.ScopeType, Applicable: false, Reason: "not applicable"}
}

// sampleReport is a deliberately "bad" one-package report: over budget on
// structure, a hot type, and one not-applicable metric that must be ignored.
func sampleReport() gomodularity.Report {
	return gomodularity.Report{
		Packages: []gomodularity.PackageReport{{
			Path:            "example.com/m/foo",
			Afferent:        2,
			Efferent:        20,
			ExportedFuncs:   20,
			UnexportedFuncs: 20,
			Metrics: []metrics.MetricResult{
				{Name: metrics.MetricDistance, Scope: metrics.ScopePackage, Value: 0.9, Applicable: true},
			},
			Types: []gomodularity.TypeReport{{
				Name:    "Big",
				Fields:  20,
				Methods: 25,
				Metrics: []metrics.MetricResult{
					typeMetric(metrics.MetricAMC, 7),
					typeMetric(metrics.MetricReusability, 0.30),
					naMetric(metrics.MetricTCC),
				},
			}},
		}},
	}
}

func TestEvaluateFlagsMaxAndMinAndSkipsNotApplicable(t *testing.T) {
	t.Parallel()

	policy := domain.Policy{
		PackageMetrics: map[string]domain.Limit{
			metrics.MetricDistance: {Max: 0.8, HasMax: true},
		},
		TypeMetrics: map[string]domain.Limit{
			metrics.MetricAMC:         {Max: 4, HasMax: true},
			metrics.MetricReusability: {Min: 0.6, HasMin: true},
			metrics.MetricTCC:         {Min: 0.5, HasMin: true}, // n/a → must be skipped
		},
	}
	policy.Package.Efferent = domain.Limit{Max: 15, HasMax: true}
	policy.Type.Fields = domain.Limit{Max: 12, HasMax: true}

	got := domain.Evaluate(sampleReport(), policy)

	// efferent (pkg) → distance (pkg) → fields, amc, reusability (type Big), in that order.
	want := []struct {
		typ, key string
		cmp      domain.Comparator
	}{
		{"", domain.KeyEfferent, domain.ComparatorMax},
		{"", metrics.MetricDistance, domain.ComparatorMax},
		{"Big", domain.KeyFields, domain.ComparatorMax},
		{"Big", metrics.MetricAMC, domain.ComparatorMax},
		{"Big", metrics.MetricReusability, domain.ComparatorMin},
	}

	if len(got) != len(want) {
		t.Fatalf("violations = %d, want %d\n%+v", len(got), len(want), got)
	}

	for i, w := range want {
		if got[i].Type != w.typ || got[i].Key != w.key || got[i].Comparator != w.cmp {
			t.Errorf("violation[%d] = (%q %q %q), want (%q %q %q)",
				i, got[i].Type, got[i].Key, got[i].Comparator, w.typ, w.key, w.cmp)
		}
	}

	for _, v := range got {
		if v.Key == metrics.MetricTCC {
			t.Errorf("not-applicable tcc was flagged: %+v", v)
		}
	}
}

func TestEvaluateCleanReportHasNoViolations(t *testing.T) {
	t.Parallel()

	clean := gomodularity.Report{
		Packages: []gomodularity.PackageReport{{
			Path:          "example.com/m/tidy",
			ExportedFuncs: 3,
			Types: []gomodularity.TypeReport{{
				Name: "Small", Fields: 2, Methods: 2,
				Metrics: []metrics.MetricResult{typeMetric(metrics.MetricAMC, 1)},
			}},
		}},
	}

	if got := domain.Evaluate(clean, domain.DefaultPolicy()); len(got) != 0 {
		t.Errorf("clean report produced violations: %+v", got)
	}
}

func TestEvaluateChecksBothBounds(t *testing.T) {
	t.Parallel()

	report := gomodularity.Report{
		Packages: []gomodularity.PackageReport{{
			Path: "p",
			Metrics: []metrics.MetricResult{
				{Name: metrics.MetricDistance, Scope: metrics.ScopePackage, Value: 0.9, Applicable: true},
			},
		}},
	}
	policy := domain.Policy{Metrics: map[string]domain.Limit{
		metrics.MetricDistance: {Min: 0.1, HasMin: true, Max: 0.6, HasMax: true},
	}}

	got := domain.Evaluate(report, policy)
	if len(got) != 1 || got[0].Comparator != domain.ComparatorMax {
		t.Fatalf("want one max violation, got %+v", got)
	}
}

func TestFormatViolations(t *testing.T) {
	t.Parallel()

	if s := domain.FormatViolations(nil); s != "" {
		t.Errorf("empty slice = %q, want empty", s)
	}

	out := domain.FormatViolations([]domain.Violation{
		{Package: "example.com/m/foo", Key: domain.KeyTypes, Value: 25, Comparator: domain.ComparatorMax, Threshold: 15},
		{Package: "example.com/m/foo", Type: "Big", Key: metrics.MetricReusability, Value: 0.42, Comparator: domain.ComparatorMin, Threshold: 0.6},
	})

	for _, want := range []string{
		"policy: 2 violations",
		"example.com/m/foo (package): types 25 exceeds max 15",
		"example.com/m/foo.Big (type): reusability 0.42 is below min 0.60",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("output missing %q\ngot:\n%s", want, out)
		}
	}

	single := domain.FormatViolations([]domain.Violation{{Package: "p", Key: domain.KeyExportedFuncs, Value: 5, Comparator: domain.ComparatorMax, Threshold: 3}})
	if !strings.HasPrefix(single, "policy: 1 violation\n") {
		t.Errorf("singular header wrong: %q", single)
	}
}

func TestValidate(t *testing.T) {
	t.Parallel()

	if err := domain.Validate(domain.DefaultPolicy()); err != nil {
		t.Errorf("DefaultPolicy invalid: %v", err)
	}

	cases := map[string]domain.Policy{
		"unknown metric": {Metrics: map[string]domain.Limit{"nope": {Max: 1, HasMax: true}}},
		"min over max":   {Metrics: map[string]domain.Limit{metrics.MetricAMC: {Min: 5, HasMin: true, Max: 2, HasMax: true}}},
		"wrong package metric": {
			PackageMetrics: map[string]domain.Limit{metrics.MetricAMC: {Max: 1, HasMax: true}},
		},
		"wrong type metric": {
			TypeMetrics: map[string]domain.Limit{metrics.MetricDistance: {Max: 1, HasMax: true}},
		},
	}
	for name, policy := range cases {
		if err := domain.Validate(policy); err == nil {
			t.Errorf("%s: want error, got nil", name)
		}
	}
}

func TestApplyOverride(t *testing.T) {
	t.Parallel()

	var policy domain.Policy

	if err := domain.ApplyOverride(&policy, domain.KeyTypes, domain.ComparatorMax, 10); err != nil {
		t.Fatal(err)
	}

	if err := domain.ApplyOverride(&policy, metrics.MetricCBO, domain.ComparatorMax, 8); err != nil {
		t.Fatal(err)
	}

	if err := domain.ApplyOverride(&policy, "type."+metrics.MetricCBO, domain.ComparatorMax, 3); err != nil {
		t.Fatal(err)
	}

	if err := domain.ApplyOverride(&policy, "package."+metrics.MetricDistance, domain.ComparatorMax, 0.5); err != nil {
		t.Fatal(err)
	}

	if err := domain.ApplyOverride(&policy, "bogus", domain.ComparatorMax, 1); err == nil {
		t.Error("unknown key: want error, got nil")
	}

	if !policy.Package.Types.HasMax || policy.Package.Types.Max != 10 {
		t.Errorf("types override not applied: %+v", policy.Package.Types)
	}

	if l := policy.Metrics[metrics.MetricCBO]; !l.HasMax || l.Max != 8 {
		t.Errorf("cbo override not applied: %+v", l)
	}

	if l := policy.TypeMetrics[metrics.MetricCBO]; !l.HasMax || l.Max != 3 {
		t.Errorf("type.cbo override not applied: %+v", l)
	}

	if l := policy.PackageMetrics[metrics.MetricDistance]; !l.HasMax || l.Max != 0.5 {
		t.Errorf("package.distance override not applied: %+v", l)
	}
}

func TestMetricNamesSorted(t *testing.T) {
	t.Parallel()

	policy := domain.Policy{Metrics: map[string]domain.Limit{
		metrics.MetricTCC: {Min: 0.3, HasMin: true},
		metrics.MetricCBO: {}, // no bound → excluded
	}, TypeMetrics: map[string]domain.Limit{
		metrics.MetricAMC: {Max: 4, HasMax: true},
	}, PackageMetrics: map[string]domain.Limit{
		metrics.MetricDistance: {Max: 0.4, HasMax: true},
	}}

	got := domain.MetricNames(policy)
	want := []string{metrics.MetricAMC, metrics.MetricDistance, metrics.MetricTCC}

	if len(got) != len(want) {
		t.Fatalf("names = %v, want %v", got, want)
	}

	for i := range want {
		if got[i] != want[i] {
			t.Errorf("names = %v, want %v", got, want)
		}
	}
}
