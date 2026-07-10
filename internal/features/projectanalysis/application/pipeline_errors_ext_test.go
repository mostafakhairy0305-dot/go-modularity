package application_test

import (
	"context"
	"errors"
	"testing"

	projectanalysis "github.com/mostafakhairy0305-dot/go-modularity/internal/features/projectanalysis/application"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/projectanalysis/ports/inbound"
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/application"
	tfdomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	tfoutbound "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/ports/outbound"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

type errSource struct{ err error }

func (e errSource) Load(context.Context, tfoutbound.FactOptions) (string, []tfdomain.PackageExtract, error) {
	return "", nil, e.err
}

// Black-box: a fact-source failure propagates out of Analyze.
func TestPipelineLoadError(t *testing.T) {
	t.Parallel()
	sentinel := errors.New("load failed")
	pipeline := projectanalysis.NewPipeline(typefacts.NewService(errSource{err: sentinel}))
	_, err := pipeline.Analyze(context.Background(), inbound.Options{
		Patterns: []string{"./..."},
		Weights:  metrics.DefaultReusabilityWeights(),
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("error = %v, want sentinel", err)
	}
}

// Black-box: invalid reusability weights fail before any loading happens.
func TestPipelineBadWeights(t *testing.T) {
	t.Parallel()
	pipeline := projectanalysis.NewPipeline(typefacts.NewService(fakeSource{mod: "example.com/m"}))
	_, err := pipeline.Analyze(context.Background(), inbound.Options{
		Patterns:        []string{"./..."},
		SelectedMetrics: []string{metrics.MetricReusability},
		Weights:         metrics.ReusabilityWeights{Cohesion: -1, Coupling: 1},
	})
	if err == nil {
		t.Fatal("invalid weights should fail the run")
	}
}
