package application

import (
	"context"

	architecture "github.com/mostafakhairy0305-dot/go-modularity/internal/features/architecture/application"
	archdomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/architecture/domain"
	cohesion "github.com/mostafakhairy0305-dot/go-modularity/internal/features/cohesion/application"
	complexity "github.com/mostafakhairy0305-dot/go-modularity/internal/features/complexity/application"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/projectanalysis/ports/inbound"
	reusability "github.com/mostafakhairy0305-dot/go-modularity/internal/features/reusability/application"
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/application"
	tfdomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	tfoutbound "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/ports/outbound"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/workerpool"
)

// runWorkers is a seam so tests can force workerpool.Run failures.
var runWorkers = workerpool.Run

// Pipeline implements the inbound Analyzer port.
type Pipeline struct {
	facts typefacts.Collector
}

// NewPipeline returns a pipeline backed by the given fact collector.
func NewPipeline(facts typefacts.Collector) *Pipeline {
	return &Pipeline{facts: facts}
}

var _ inbound.Analyzer = (*Pipeline)(nil)

// Analyze runs the full pipeline for one request.
func (p *Pipeline) Analyze(ctx context.Context, opts inbound.Options) (inbound.Result, error) {
	display := nameSet(opts.SelectedMetrics)
	compute := nameSet(metrics.Closure(opts.SelectedMetrics))

	// Reusability weights are validated up front so a bad configuration
	// fails before any loading happens.
	reusabilityCalculator, err := newReusabilityCalculator(compute, opts.Weights)
	if err != nil {
		return inbound.Result{}, err
	}

	facts, err := p.facts.Collect(ctx, collectOptions(opts))
	if err != nil {
		return inbound.Result{}, err
	}

	return assembleResult(ctx, &facts, reusabilityCalculator, display, compute, opts)
}

// assembleResult computes the architecture metrics and every package's results
// in parallel, honouring cancellation, and folds them into the final report.
func assembleResult(
	ctx context.Context,
	facts *tfdomain.ProjectFacts,
	reusabilityCalculator reusability.Calculator,
	display, compute map[string]bool,
	opts inbound.Options,
) (inbound.Result, error) {
	if err := ctx.Err(); err != nil {
		return inbound.Result{}, err
	}

	// The dependency graph is cheap and feeds the structural Ca/Ce facts,
	// so it is built regardless of the selected metrics.
	graph := archdomain.BuildDependencyGraph(facts, archdomain.Scope(opts.DependencyScope))
	archResults := computeArchitecture(facts, graph, compute)

	packageResults := make([]inbound.PackageResult, len(facts.Packages))
	workers := workerpool.Workers(opts.Workers, len(facts.Packages))

	err := runWorkers(ctx, workers, len(facts.Packages), func(i int) error {
		packageResults[i] = analyzePackage(
			facts,
			i,
			graph.Coupling(i),
			archResults,
			reusabilityCalculator,
			display,
			compute,
			opts,
		)

		return nil
	})
	if err != nil {
		return inbound.Result{}, err
	}

	return inbound.Result{ModulePath: facts.ModulePath, Packages: packageResults}, nil
}

// newReusabilityCalculator builds the reusability calculator when the compute
// set needs it; a nil calculator disables per-type reusability and CBO.
func newReusabilityCalculator(
	compute map[string]bool,
	weights metrics.ReusabilityWeights,
) (reusability.Calculator, error) {
	if !compute[metrics.MetricReusability] && !compute[metrics.MetricCBO] {
		return nil, nil
	}

	return reusability.NewService(weights)
}

// collectOptions maps the analysis request onto the fact-source options.
func collectOptions(opts inbound.Options) tfoutbound.FactOptions {
	return tfoutbound.FactOptions{
		Directory:        opts.Directory,
		Patterns:         opts.Patterns,
		IncludeTests:     opts.IncludeTests,
		IncludeGenerated: opts.IncludeGenerated,
		BuildTags:        opts.BuildTags,
		Workers:          opts.Workers,
		ContinueOnError:  opts.ContinueOnError,
	}
}

// computeArchitecture runs the package-level metrics once over the whole
// dependency graph; nil when no package metric is in the compute set.
func computeArchitecture(
	facts *tfdomain.ProjectFacts,
	graph archdomain.CouplingGraph,
	compute map[string]bool,
) []architecture.Result {
	if !compute[metrics.MetricAbstractness] && !compute[metrics.MetricInstability] &&
		!compute[metrics.MetricDistance] {
		return nil
	}

	return architecture.ComputeForPackages(facts, graph)
}

// analyzePackage computes one package's display metrics and those of its
// types. It writes only into its own result value, so package workers never
// share mutable state.
func analyzePackage(
	facts *tfdomain.ProjectFacts,
	pkgID int,
	coupling archdomain.Coupling,
	archResults []architecture.Result,
	reusabilityCalculator reusability.Calculator,
	display, compute map[string]bool,
	opts inbound.Options,
) inbound.PackageResult {
	pkg := &facts.Packages[pkgID]

	result := inbound.PackageResult{
		Path:            pkg.Path,
		Afferent:        coupling.Afferent,
		Efferent:        coupling.Efferent,
		ExportedFuncs:   pkg.ExportedFuncCount,
		UnexportedFuncs: pkg.UnexportedFuncCount,
	}
	if archResults != nil {
		result.Metrics = packageMetrics(archResults[pkgID], display)
	}

	needComplexity := compute[metrics.MetricAMC]
	needCohesion := compute[metrics.MetricLCOM1] || compute[metrics.MetricLCOM96b] ||
		compute[metrics.MetricTCC] || compute[metrics.MetricCAMC]

	result.Types = make([]inbound.TypeResult, 0, len(pkg.TypeIDs))
	for _, typeID := range pkg.TypeIDs {
		result.Types = append(result.Types, analyzeType(
			&facts.Types[typeID],
			reusabilityCalculator,
			display,
			needComplexity,
			needCohesion,
			opts,
		))
	}

	return result
}

// packageMetrics selects the displayed package-level metrics in the fixed
// metric order.
func packageMetrics(arch architecture.Result, display map[string]bool) []metrics.MetricResult {
	var out []metrics.MetricResult

	for _, name := range metrics.PackageMetricOrder() {
		if !display[name] {
			continue
		}

		switch name {
		case metrics.MetricAbstractness:
			out = append(out, arch.Abstractness)
		case metrics.MetricInstability:
			out = append(out, arch.Instability)
		case metrics.MetricDistance:
			out = append(out, arch.Distance)
		}
	}

	return out
}

// analyzeType computes one type's displayed metrics in the fixed metric
// order.
func analyzeType(
	t *tfdomain.TypeFacts,
	reusabilityCalculator reusability.Calculator,
	display map[string]bool,
	needComplexity, needCohesion bool,
	opts inbound.Options,
) inbound.TypeResult {
	var complexityResult complexity.Result
	if needComplexity {
		complexityResult = complexity.ComputeForType(t)
	}

	var cohesionResult cohesion.Result
	if needCohesion {
		cohesionResult = cohesion.ComputeForType(t, opts.FieldUsageTransitive)
	}

	var reusabilityResult reusability.Result
	if reusabilityCalculator != nil {
		reusabilityResult = reusabilityCalculator.ComputeForType(
			t, complexityResult.AMC, cohesionResult.LCOM96b,
		)
	}

	typeResult := inbound.TypeResult{
		Name:    t.Name,
		Fields:  len(t.Fields),
		Methods: len(t.Methods),
	}

	for _, name := range metrics.TypeMetricOrder() {
		if !display[name] {
			continue
		}

		switch name {
		case metrics.MetricAMC:
			typeResult.Metrics = append(typeResult.Metrics, complexityResult.AMC)
		case metrics.MetricLCOM1:
			typeResult.Metrics = append(typeResult.Metrics, cohesionResult.LCOM1)
		case metrics.MetricLCOM96b:
			typeResult.Metrics = append(typeResult.Metrics, cohesionResult.LCOM96b)
		case metrics.MetricCAMC:
			typeResult.Metrics = append(typeResult.Metrics, cohesionResult.CAMC)
		case metrics.MetricTCC:
			typeResult.Metrics = append(typeResult.Metrics, cohesionResult.TCC)
		case metrics.MetricCBO:
			typeResult.Metrics = append(typeResult.Metrics, reusabilityResult.CBO)
		case metrics.MetricReusability:
			typeResult.Metrics = append(typeResult.Metrics, reusabilityResult.Reusability)
		}
	}

	return typeResult
}

func nameSet(names []string) map[string]bool {
	set := make(map[string]bool, len(names))
	for _, name := range names {
		set[name] = true
	}

	return set
}
