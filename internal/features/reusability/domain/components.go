// Package domain holds the reusability feature's entity logic: deriving the
// four index components from source facts and upstream metric results, and
// computing type-level CBO from the referenced-types fact.
//
// Upstream results arrive as plain metric results handed over by the
// orchestrator — this feature never imports the cohesion or complexity
// features, and package instability is never used as type-level coupling.
package domain

import (
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

// Result is the reusability feature's output for one type.
type Result struct {
	// CBO is the normalized coupling input, reported standalone only when
	// selected.
	CBO metrics.MetricResult
	// Reusability is the experimental reusability index.
	Reusability metrics.MetricResult
}

// Compute derives CBO from the type's referenced-types fact, assembles the
// four components (dropping the not-applicable ones), and evaluates the
// index with renormalized weights.
func Compute(
	t *typefacts.TypeFacts,
	amc, lcom96b metrics.MetricResult,
	weights metrics.ReusabilityWeights,
) Result {
	cbo := len(t.ReferencedTypeIDs)
	return Result{
		CBO: metrics.CBO(cbo),
		Reusability: metrics.Reusability(
			metrics.CohesionComponent(lcom96b),
			metrics.CouplingComponent(cbo),
			metrics.TestabilityComponent(amc),
			metrics.DocumentationComponent(t.DocumentedExportedMembers, t.ExportedMembers),
			weights,
		),
	}
}
