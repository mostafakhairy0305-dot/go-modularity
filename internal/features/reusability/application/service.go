package application

import (
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/reusability/domain"
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

// Result re-exports the domain result for consumers of the feature.
type Result = domain.Result

// Service evaluates the reusability index with a fixed weight set.
type Service struct {
	weights metrics.ReusabilityWeights
}

// NewService validates the weights and returns a reusability evaluator.
// Zero-value weights select the defaults.
func NewService(weights metrics.ReusabilityWeights) (*Service, error) {
	if (weights == metrics.ReusabilityWeights{}) {
		weights = metrics.DefaultReusabilityWeights()
	}

	err := weights.Validate()
	if err != nil {
		return nil, err
	}

	return &Service{weights: weights}, nil
}

// ComputeForType evaluates CBO and the reusability index for one type. The
// AMC and LCOM96b results are supplied by the orchestrator.
func (s *Service) ComputeForType(t *typefacts.TypeFacts, amc, lcom96b metrics.MetricResult) Result {
	return domain.Compute(t, amc, lcom96b, s.weights)
}
