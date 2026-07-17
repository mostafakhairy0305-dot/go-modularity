package analyzer

import (
	"context"
	"fmt"
	"go/token"
	"sync"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
	policydomain "github.com/mostafakhairy0305-dot/go-modularity/internal/features/policy/domain"
	"golang.org/x/tools/go/analysis"
)

// Name is the analyzer and golangci-lint custom linter name.
const Name = "gomodularity"

// Doc is the short documentation shown by go/analysis tooling.
const Doc = `enforce Go modularity policy thresholds

Reports policy violations for type- and package-level modularity metrics
(AMC, LCOM*, CAMC, TCC, CBO, reusability, abstractness, instability, distance,
and structural budgets). Configuration matches go-modularity -check /
.modularity.yml.`

// New returns a go/analysis Analyzer that loads the module once, evaluates the
// modularity policy, and emits diagnostics for the package under analysis.
func New(settings Settings) (*analysis.Analyzer, error) {
	s := settings.withDefaults()
	if err := s.validate(); err != nil {
		return nil, err
	}

	r := newRunner(s)

	return &analysis.Analyzer{
		Name: Name,
		Doc:  Doc,
		Run:  r.run,
	}, nil
}

// reportAnalyzer is the facade behavior needed by the go/analysis runner.
// Keeping it narrow allows the runner to be exercised without loading a real
// module.
type reportAnalyzer interface {
	// Analyze evaluates one modularity configuration.
	Analyze(context.Context, gomodularity.Config) (gomodularity.Report, error)
}

// analyzeFunc adapts a function to reportAnalyzer.
type analyzeFunc func(context.Context, gomodularity.Config) (gomodularity.Report, error)

// Analyze delegates to the adapted function.
func (f analyzeFunc) Analyze(
	ctx context.Context,
	cfg gomodularity.Config,
) (gomodularity.Report, error) {
	return f(ctx, cfg)
}

// runner caches a whole-module analysis so each package pass only filters and
// reports diagnostics. Coupling metrics require a multi-package load.
type runner struct {
	settings Settings
	analyzer reportAnalyzer

	once  sync.Once
	byPkg map[string][]policydomain.Violation
	err   error
}

func newRunner(settings Settings) *runner {
	return &runner{settings: settings, analyzer: analyzeFunc(gomodularity.Analyze)}
}

func (r *runner) run(pass *analysis.Pass) (any, error) {
	r.once.Do(r.load)

	if r.err != nil {
		return nil, r.err
	}

	path := pass.Pkg.Path()
	for _, v := range r.byPkg[path] {
		pass.Report(analysis.Diagnostic{
			Pos:      violationPos(pass, v),
			Category: Name,
			Message:  formatViolation(v),
		})
	}

	return nil, nil
}

func (r *runner) load() {
	policy, err := resolvePolicy(r.settings.Config, r.settings.Directory)
	if err != nil {
		r.err = fmt.Errorf("gomodularity policy: %w", err)

		return
	}

	report, err := r.analyzer.Analyze(
		context.Background(),
		r.settings.toConfig(gatedMetrics(policy)),
	)
	if err != nil {
		r.err = fmt.Errorf("gomodularity analyze: %w", err)

		return
	}

	r.byPkg = groupByPackage(policydomain.Evaluate(report, policy))
}

func groupByPackage(violations []policydomain.Violation) map[string][]policydomain.Violation {
	byPkg := make(map[string][]policydomain.Violation, len(violations))
	for _, v := range violations {
		byPkg[v.Package] = append(byPkg[v.Package], v)
	}

	return byPkg
}

func violationPos(pass *analysis.Pass, v policydomain.Violation) token.Pos {
	if v.Type != "" {
		return typePos(pass, v.Type)
	}

	return packagePos(pass)
}
