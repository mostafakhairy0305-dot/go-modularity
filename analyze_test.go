package gomodularity_test

import (
	"context"
	"math"
	"reflect"
	"sync"
	"testing"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
)

const epsilon = 1e-12

// The default-config report is loaded once and shared read-only across the
// many default-config tests — package loading dominates test time, so this
// avoids re-running the analyzer for every case. Config-varying tests (mutate
// != nil) still load fresh.
var (
	defaultOnce   sync.Once
	defaultReport gomodularity.Report
	defaultErr    error
)

func analyzeFixture(t *testing.T, mutate func(*gomodularity.Config)) gomodularity.Report {
	t.Helper()

	if mutate == nil {
		defaultOnce.Do(func() {
			defaultReport, defaultErr = gomodularity.Analyze(
				context.Background(), gomodularity.Config{Directory: "testdata/fixture"},
			)
		})

		if defaultErr != nil {
			t.Fatal(defaultErr)
		}

		return defaultReport
	}

	cfg := gomodularity.Config{Directory: "testdata/fixture"}
	mutate(&cfg)

	report, err := gomodularity.Analyze(context.Background(), cfg)
	if err != nil {
		t.Fatal(err)
	}

	return report
}

func findPackage(t *testing.T, report gomodularity.Report, path string) gomodularity.PackageReport {
	t.Helper()

	for _, pkg := range report.Packages {
		if pkg.Path == path {
			return pkg
		}
	}

	t.Fatalf("package %s not in report", path)

	return gomodularity.PackageReport{}
}

func findType(t *testing.T, pkg gomodularity.PackageReport, name string) gomodularity.TypeReport {
	t.Helper()

	for _, typ := range pkg.Types {
		if typ.Name == name {
			return typ
		}
	}

	t.Fatalf("type %s not in package %s", name, pkg.Path)

	return gomodularity.TypeReport{}
}

func metric(t *testing.T, results []gomodularity.MetricResult, name string) gomodularity.MetricResult {
	t.Helper()

	for _, r := range results {
		if r.Name == name {
			return r
		}
	}

	t.Fatalf("metric %s not present in %v", name, results)

	return gomodularity.MetricResult{}
}

func wantValue(t *testing.T, results []gomodularity.MetricResult, name string, want float64) {
	t.Helper()

	r := metric(t, results, name)
	if !r.Applicable {
		t.Fatalf("%s not applicable (%s), want %v", name, r.Reason, want)
	}

	if math.Abs(r.Value-want) > epsilon {
		t.Fatalf("%s = %v, want %v", name, r.Value, want)
	}
}

func wantNotApplicable(t *testing.T, results []gomodularity.MetricResult, name string) {
	t.Helper()

	r := metric(t, results, name)
	if r.Applicable {
		t.Fatalf("%s applicable with value %v, want n/a", name, r.Value)
	}

	if r.Reason == "" {
		t.Fatalf("%s n/a without reason", name)
	}
}

func TestAnalyzeFixtureOrdering(t *testing.T) {
	report := analyzeFixture(t, nil)

	wantOrder := []string{
		"example.com/fixture/embedding",
		"example.com/fixture/gen",
		"example.com/fixture/generics",
		"example.com/fixture/isolated",
		"example.com/fixture/multifile",
		"example.com/fixture/orders",
		"example.com/fixture/store",
	}
	if len(report.Packages) != len(wantOrder) {
		t.Fatalf("got %d packages", len(report.Packages))
	}

	for i, path := range wantOrder {
		if report.Packages[i].Path != path {
			t.Fatalf("packages[%d] = %s, want %s", i, report.Packages[i].Path, path)
		}
	}

	if report.SchemaVersion != gomodularity.SchemaVersion || report.Tool.Name != gomodularity.ToolName {
		t.Fatalf("report header = %+v", report)
	}
}

func TestAnalyzeOrderType(t *testing.T) {
	report := analyzeFixture(t, nil)
	order := findType(t, findPackage(t, report, "example.com/fixture/orders"), "Order")

	wantValue(t, order.Metrics, "amc", 5.0/3)     // CC 3 + 1 + 1 over 3 methods
	wantValue(t, order.Metrics, "lcom1", 3)       // three disjoint field sets
	wantValue(t, order.Metrics, "lcom96b", 2.0/3) // 3 accesses over 3×3 matrix
	wantValue(t, order.Metrics, "camc", 1.0/3)    // 2 one-cells over 3×2 matrix
	wantValue(t, order.Metrics, "tcc", 0)
	// cohesion 1/3, coupling 1/2, testability 3/5, documentation 2/3.
	wantRI := 0.35*(1.0/3) + 0.25*0.5 + 0.25*0.6 + 0.15*(2.0/3)
	wantValue(t, order.Metrics, "reusability", wantRI)

	// CBO is computed (reusability needs it) but not displayed by default.
	for _, r := range order.Metrics {
		if r.Name == "cbo" {
			t.Fatal("cbo displayed without being selected")
		}
	}
}

func TestAnalyzePackageMetrics(t *testing.T) {
	report := analyzeFixture(t, nil)

	store := findPackage(t, report, "example.com/fixture/store")
	wantValue(t, store.Metrics, "abstractness", 1)
	wantValue(t, store.Metrics, "instability", 0) // Ca=1 (orders), Ce=0
	wantValue(t, store.Metrics, "distance", 0)

	orders := findPackage(t, report, "example.com/fixture/orders")
	wantValue(t, orders.Metrics, "abstractness", 0)
	wantValue(t, orders.Metrics, "instability", 1) // Ca=0, Ce=1 (store; fmt out of scope)
	wantValue(t, orders.Metrics, "distance", 0)

	// An isolated package (Ca = Ce = 0) is defined as maximally stable:
	// instability 0, so distance = |0 + 0 − 1| = 1.
	isolated := findPackage(t, report, "example.com/fixture/isolated")
	wantValue(t, isolated.Metrics, "instability", 0)
	wantValue(t, isolated.Metrics, "distance", 1)
	wantValue(t, isolated.Metrics, "abstractness", 0)

	if r := metric(t, isolated.Metrics, "instability"); r.Reason == "" {
		t.Fatal("isolated instability should carry the defined-as-0 reason")
	}

	// Store declares no receiver-carrying methods: type metrics are n/a.
	storeType := findType(t, store, "Store")
	wantNotApplicable(t, storeType.Metrics, "amc")
	wantNotApplicable(t, storeType.Metrics, "lcom96b")
}

func TestAnalyzeGenericsAndEmbedding(t *testing.T) {
	report := analyzeFixture(t, nil)

	pair := findType(t, findPackage(t, report, "example.com/fixture/generics"), "Pair")
	wantValue(t, pair.Metrics, "amc", 1)
	wantValue(t, pair.Metrics, "lcom1", 0) // max(1−2, 0)
	wantValue(t, pair.Metrics, "tcc", 2.0/3)
	wantValue(t, pair.Metrics, "camc", 0.5) // T and Pair[T] stay distinct
	wantValue(t, pair.Metrics, "lcom96b", 1.0/3)

	embedding := findPackage(t, report, "example.com/fixture/embedding")
	wrapper := findType(t, embedding, "Wrapper")
	// Only Describe counts: promoted Inc is Base's method.
	wantValue(t, wrapper.Metrics, "amc", 2) // one if
	// Describe uses the Base slot and Name: full 2×1 matrix.
	wantValue(t, wrapper.Metrics, "lcom96b", 0)
	wantNotApplicable(t, wrapper.Metrics, "lcom1")

	base := findType(t, embedding, "Base")
	wantValue(t, base.Metrics, "amc", 1)
	wantValue(t, base.Metrics, "lcom96b", 0)
}

func TestAnalyzeCBOSelected(t *testing.T) {
	report := analyzeFixture(t, func(cfg *gomodularity.Config) {
		cfg.SelectedMetrics = []gomodularity.MetricName{gomodularity.MetricCBO}
	})

	order := findType(t, findPackage(t, report, "example.com/fixture/orders"), "Order")
	if len(order.Metrics) != 1 {
		t.Fatalf("metrics = %v, want cbo only", order.Metrics)
	}

	wantValue(t, order.Metrics, "cbo", 1) // references store.Store

	wrapper := findType(t, findPackage(t, report, "example.com/fixture/embedding"), "Wrapper")
	wantValue(t, wrapper.Metrics, "cbo", 1) // embedded Base

	pair := findType(t, findPackage(t, report, "example.com/fixture/generics"), "Pair")
	wantValue(t, pair.Metrics, "cbo", 0) // type params and self excluded

	// Package metrics were neither computed nor displayed.
	if pkg := findPackage(t, report, "example.com/fixture/orders"); len(pkg.Metrics) != 0 {
		t.Fatalf("package metrics = %v, want none", pkg.Metrics)
	}
}

func TestAnalyzeTransitiveFieldUsage(t *testing.T) {
	direct := analyzeFixture(t, nil)
	counter := findType(t, findPackage(t, direct, "example.com/fixture/multifile"), "Counter")
	wantValue(t, counter.Metrics, "lcom1", 1)
	wantValue(t, counter.Metrics, "tcc", 1.0/3)
	wantValue(t, counter.Metrics, "lcom96b", 0.5)

	transitive := analyzeFixture(t, func(cfg *gomodularity.Config) {
		cfg.FieldUsageMode = gomodularity.FieldUsageTransitive
	})
	counter = findType(t, findPackage(t, transitive, "example.com/fixture/multifile"), "Counter")
	wantValue(t, counter.Metrics, "lcom1", 0)
	wantValue(t, counter.Metrics, "tcc", 1)
	wantValue(t, counter.Metrics, "lcom96b", 1.0/6)
}

func TestAnalyzeGeneratedFiles(t *testing.T) {
	report := analyzeFixture(t, nil)

	gen := findPackage(t, report, "example.com/fixture/gen")
	if len(gen.Types) != 0 {
		t.Fatalf("generated types analyzed by default: %v", gen.Types)
	}

	wantNotApplicable(t, gen.Metrics, "abstractness")

	report = analyzeFixture(t, func(cfg *gomodularity.Config) { cfg.IncludeGenerated = true })
	gen = findPackage(t, report, "example.com/fixture/gen")
	machine := findType(t, gen, "Machine")
	wantValue(t, machine.Metrics, "amc", 1)
}

func TestAnalyzeDeterminism(t *testing.T) {
	first := analyzeFixture(t, func(cfg *gomodularity.Config) { cfg.Workers = 1 })

	second := analyzeFixture(t, func(cfg *gomodularity.Config) { cfg.Workers = 8 })
	if !reflect.DeepEqual(first, second) {
		t.Fatal("reports differ across worker counts")
	}

	third := analyzeFixture(t, func(cfg *gomodularity.Config) { cfg.Workers = 8 })
	if !reflect.DeepEqual(second, third) {
		t.Fatal("repeated runs differ")
	}
}

func TestAnalyzeInvalidConfig(t *testing.T) {
	ctx := context.Background()
	base := gomodularity.Config{Directory: "testdata/fixture"}

	bad := base

	bad.DependencyScope = "galaxy"
	if _, err := gomodularity.Analyze(ctx, bad); err == nil {
		t.Fatal("invalid scope accepted")
	}

	bad = base

	bad.SelectedMetrics = []gomodularity.MetricName{"nope"}
	if _, err := gomodularity.Analyze(ctx, bad); err == nil {
		t.Fatal("unknown metric accepted")
	}

	bad = base

	bad.ReusabilityWeights = gomodularity.ReusabilityWeights{Cohesion: -1, Coupling: 2}
	if _, err := gomodularity.Analyze(ctx, bad); err == nil {
		t.Fatal("negative weight accepted")
	}
}

func TestAnalyzeCancelledContext(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	if _, err := gomodularity.Analyze(ctx, gomodularity.Config{Directory: "testdata/fixture"}); err == nil {
		t.Fatal("cancelled context accepted")
	}
}
