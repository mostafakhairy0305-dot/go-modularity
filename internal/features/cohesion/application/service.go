package application

import (
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/cohesion/domain"
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

// Result is the cohesion feature's output for one type.
type Result struct {
	// LCOM1 is the non-sharing method-pair surplus.
	LCOM1 metrics.MetricResult
	// LCOM96b is the method-field matrix sparsity.
	LCOM96b metrics.MetricResult
	// TCC is the connected method-pair density.
	TCC metrics.MetricResult
	// CAMC is the parameter-type agreement.
	CAMC metrics.MetricResult
}

// ComputeForType evaluates all cohesion metrics for one type. transitive
// selects the transitive field-usage mode.
func ComputeForType(t *typefacts.TypeFacts, transitive bool) Result {
	methodCount := len(t.Methods)
	fieldCount := len(t.Fields)

	sets := domain.EffectiveFieldSets(t, transitive)
	pairs := domain.CountPairs(sets, fieldCount)
	accesses := domain.TotalFieldAccesses(sets)
	oneCells, distinctParams := domain.ParamMatrix(t.Methods)

	return Result{
		LCOM1:   metrics.LCOM1(pairs.NonSharing, pairs.Sharing, methodCount, fieldCount),
		LCOM96b: metrics.LCOM96b(accesses, fieldCount, methodCount),
		TCC:     metrics.TCC(pairs.Sharing, methodCount),
		CAMC:    metrics.CAMC(oneCells, methodCount, distinctParams),
	}
}
