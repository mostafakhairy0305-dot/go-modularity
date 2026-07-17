package application_test

import (
	"context"
	"testing"

	projectanalysis "github.com/mostafakhairy0305-dot/go-modularity/internal/features/projectanalysis/application"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/projectanalysis/ports/inbound"
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/application"
	tfdomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	tfoutbound "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/ports/outbound"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

// fakeSource feeds canned extracts so the whole pipeline runs without loading
// real packages through go/packages.
type fakeSource struct {
	mod  string
	pkgs []tfdomain.PackageExtract
}

func (f fakeSource) Load(context.Context, tfoutbound.FactOptions) (string, []tfdomain.PackageExtract, error) {
	return f.mod, f.pkgs, nil
}

func findMetric(t *testing.T, results []metrics.MetricResult, name string) metrics.MetricResult {
	t.Helper()

	for _, r := range results {
		if r.Name == name {
			return r
		}
	}

	t.Fatalf("metric %q not present in %v", name, results)

	return metrics.MetricResult{}
}

// Black-box: the pipeline turns extracts into a deterministic report with the
// selected package- and type-level metrics.
func TestPipelineAnalyzeEndToEnd(t *testing.T) {
	t.Parallel()

	src := fakeSource{
		mod: "example.com/m",
		pkgs: []tfdomain.PackageExtract{
			{
				Path: "example.com/m/a", InModule: true, Imports: []string{"example.com/m/b"},
				Types: []tfdomain.TypeExtract{
					{
						Name: "A", Exported: true, Kind: tfdomain.KindStruct,
						Methods: []tfdomain.MethodFacts{{Name: "Do", Exported: true, Branches: tfdomain.BranchStats{Ifs: 1}}},
					},
				},
			},
			{
				Path: "example.com/m/b", InModule: true,
				Types: []tfdomain.TypeExtract{{Name: "B", Exported: true, Kind: tfdomain.KindInterface}},
			},
		},
	}
	pipeline := projectanalysis.NewPipeline(typefacts.NewService(src))

	result, err := pipeline.Analyze(context.Background(), inbound.Options{
		Patterns:        []string{"./..."},
		DependencyScope: "project",
		SelectedMetrics: []string{
			metrics.MetricAMC,
			metrics.MetricLCOM1,
			metrics.MetricLCOM96b,
			metrics.MetricCAMC,
			metrics.MetricTCC,
			metrics.MetricCBO,
			metrics.MetricReusability,
			metrics.MetricAbstractness,
			metrics.MetricInstability,
			metrics.MetricDistance,
		},
		Weights: metrics.DefaultReusabilityWeights(),
	})
	if err != nil {
		t.Fatal(err)
	}

	if result.ModulePath != "example.com/m" {
		t.Fatalf("module = %q", result.ModulePath)
	}

	if len(result.Packages) != 2 || result.Packages[0].Path != "example.com/m/a" ||
		result.Packages[1].Path != "example.com/m/b" {
		t.Fatalf("packages not sorted by path: %+v", result.Packages)
	}

	// a: concrete, imports b → A=0, I=1, D=0. Type A has one CC-2 method → AMC=2.
	pkgA := result.Packages[0]
	if got := findMetric(t, pkgA.Metrics, metrics.MetricInstability); got.Value != 1 {
		t.Errorf("a instability = %v, want 1", got.Value)
	}

	if got := findMetric(t, pkgA.Metrics, metrics.MetricAbstractness); got.Value != 0 {
		t.Errorf("a abstractness = %v, want 0", got.Value)
	}

	if len(pkgA.Types) != 1 || pkgA.Types[0].Name != "A" {
		t.Fatalf("a types = %+v", pkgA.Types)
	}

	if got := findMetric(t, pkgA.Types[0].Metrics, metrics.MetricAMC); got.Value != 2 {
		t.Errorf("A amc = %v, want 2", got.Value)
	}

	// b: all-interface → A=1.
	if got := findMetric(t, result.Packages[1].Metrics, metrics.MetricAbstractness); got.Value != 1 {
		t.Errorf("b abstractness = %v, want 1", got.Value)
	}
}

// Black-box: a cancelled context aborts before doing work.
func TestPipelineCancelled(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	pipeline := projectanalysis.NewPipeline(typefacts.NewService(fakeSource{mod: "m"}))
	if _, err := pipeline.Analyze(ctx, inbound.Options{Patterns: []string{"./..."}, Weights: metrics.DefaultReusabilityWeights()}); err == nil {
		t.Fatal("cancelled context should abort")
	}
}
