package application

import (
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/architecture/domain"
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

// Result is the architecture feature's output for one package.
type Result struct {
	// Abstractness is the package's interface ratio.
	Abstractness metrics.MetricResult
	// Instability is the package's efferent coupling ratio.
	Instability metrics.MetricResult
	// Distance is the package's distance from the main sequence.
	Distance metrics.MetricResult
}

// ComputeForPackages evaluates the package-level metrics for every analyzed
// package, indexed by package ID.
func ComputeForPackages(facts *typefacts.ProjectFacts, scope domain.Scope) []Result {
	graph := domain.BuildDependencyGraph(facts, scope)

	results := make([]Result, len(facts.Packages))
	for i := range facts.Packages {
		counts := domain.CountTypes(facts, &facts.Packages[i])
		coupling := graph.Couplings[i]

		abstractness := metrics.Abstractness(counts.Interfaces, counts.Total)
		instability := metrics.Instability(coupling.Afferent, coupling.Efferent)
		results[i] = Result{
			Abstractness: abstractness,
			Instability:  instability,
			Distance:     metrics.Distance(abstractness, instability),
		}
	}

	return results
}
