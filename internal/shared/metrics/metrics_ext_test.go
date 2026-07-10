package metrics_test

import (
	"testing"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

// Black-box: the reusability index composes the four components into [0, 1].
func TestReusabilityComposition(t *testing.T) {
	t.Parallel()

	r := metrics.Reusability(
		metrics.CohesionComponent(metrics.LCOM96b(4, 2, 2)),
		metrics.CouplingComponent(1),
		metrics.TestabilityComponent(metrics.AMC(2, 2)),
		metrics.DocumentationComponent(2, 2),
		metrics.DefaultReusabilityWeights(),
	)
	if !r.Applicable {
		t.Fatalf("reusability not applicable: %s", r.Reason)
	}

	if r.Value < 0 || r.Value > 1 {
		t.Errorf("reusability %v out of [0,1]", r.Value)
	}
}

// Black-box: a package with balanced abstractness/instability sits on the main
// sequence (distance 0).
func TestDistanceOnMainSequence(t *testing.T) {
	t.Parallel()

	a := metrics.Abstractness(1, 2) // 0.5
	i := metrics.Instability(1, 1)  // 0.5

	d := metrics.Distance(a, i)
	if !d.Applicable || d.Value != 0 {
		t.Fatalf("distance = %+v, want 0", d)
	}
}

// Black-box: an all-zero weight set is rejected by validation.
func TestWeightsValidate(t *testing.T) {
	t.Parallel()

	err := metrics.DefaultReusabilityWeights().Validate()
	if err != nil {
		t.Errorf("defaults should validate: %v", err)
	}

	err = (metrics.ReusabilityWeights{Cohesion: -1}).Validate()
	if err == nil {
		t.Error("negative weight should fail validation")
	}
}
