package application

import (
	"context"
	"errors"
	"testing"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/projectanalysis/ports/inbound"
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/application"
	tfdomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	tfoutbound "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/ports/outbound"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

type coverageSource struct{}

func (coverageSource) Load(context.Context, tfoutbound.FactOptions) (string, []tfdomain.PackageExtract, error) {
	return "example.com/m", []tfdomain.PackageExtract{{
		Path: "example.com/m/a", InModule: true,
		Types: []tfdomain.TypeExtract{{
			Name: "A", Exported: true, Kind: tfdomain.KindStruct,
			Methods: []tfdomain.MethodFacts{{Name: "Do", Exported: true, Branches: tfdomain.BranchStats{Ifs: 1}}},
		}},
	}}, nil
}

func TestAssembleResultWorkerError(t *testing.T) {
	original := runWorkers
	t.Cleanup(func() { runWorkers = original })

	sentinel := errors.New("workers failed")
	runWorkers = func(context.Context, int, int, func(int) error) error {
		return sentinel
	}

	pipeline := NewPipeline(typefacts.NewService(coverageSource{}))
	_, err := pipeline.Analyze(context.Background(), inbound.Options{
		Patterns:        []string{"./..."},
		SelectedMetrics: []string{metrics.MetricAMC},
		Weights:         metrics.DefaultReusabilityWeights(),
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("Analyze error = %v, want sentinel", err)
	}
}

func TestPartialSelectedMetricsSkipsDisplay(t *testing.T) {
	// Selecting distance (and amc) expands compute to abstractness/instability
	// and other type deps, but only the selected names appear in the report.
	pipeline := NewPipeline(typefacts.NewService(coverageSource{}))
	result, err := pipeline.Analyze(context.Background(), inbound.Options{
		Patterns:        []string{"./..."},
		DependencyScope: "project",
		SelectedMetrics: []string{metrics.MetricDistance, metrics.MetricAMC},
		Weights:         metrics.DefaultReusabilityWeights(),
	})
	if err != nil {
		t.Fatal(err)
	}

	pkg := result.Packages[0]
	for _, r := range pkg.Metrics {
		if r.Name != metrics.MetricDistance {
			t.Fatalf("unexpected package metric %q in display set", r.Name)
		}
	}
	for _, typ := range pkg.Types {
		for _, r := range typ.Metrics {
			if r.Name != metrics.MetricAMC {
				t.Fatalf("unexpected type metric %q in display set", r.Name)
			}
		}
	}
}
