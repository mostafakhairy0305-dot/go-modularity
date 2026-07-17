package analyzer

import (
	"cmp"
	"fmt"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity/gomodularity"
)

// Settings configures the modularity policy analyzer. Analysis fields map to
// the gomodularity.Config facade. Policy fields use the same package/type
// min/max shape as CLI thresholds, but are decoded directly from
// golangci-lint's linters.settings.custom.gomodularity.settings block.
type Settings struct {
	// Directory is the working directory for package loading. Empty means the
	// process working directory.
	Directory string `json:"directory"`
	// Patterns are the package patterns to analyze. Empty means ["./..."].
	Patterns []string `json:"patterns"`
	// Tests includes test files and test packages.
	Tests bool `json:"tests"`
	// Generated includes files with the standard generated-code marker.
	Generated bool `json:"generated"`
	// DependencyScope is "project", "module", or "all". Empty means "module".
	DependencyScope string `json:"dependency-scope"`
	// FieldUsage is "direct" or "transitive". Empty means "direct".
	FieldUsage string `json:"field-usage"`
	// Workers bounds analysis concurrency. Zero selects the facade default.
	Workers int `json:"workers"`
	// ContinueOnError skips packages that fail to load or type-check.
	ContinueOnError bool `json:"continue-on-error"`
	// BuildTags are extra build tags for package loading.
	BuildTags []string `json:"build-tags"`
	// Package configures package-level structural and metric limits. Nil,
	// together with nil Type and Metrics, selects the recommended defaults.
	Package *PackageSettings `json:"package"`
	// Type configures type-level structural and metric limits.
	Type *TypeSettings `json:"type"`
	// Metrics configures legacy/global metric limits. Prefer the scoped metric
	// maps under Package and Type for new configurations.
	Metrics map[string]LimitSettings `json:"metrics"`
}

func (s Settings) withDefaults() Settings {
	if len(s.Patterns) == 0 {
		s.Patterns = []string{"./..."}
	}

	s.DependencyScope = cmp.Or(s.DependencyScope, string(gomodularity.DependencyScopeModule))
	s.FieldUsage = cmp.Or(s.FieldUsage, string(gomodularity.FieldUsageDirect))

	return s
}

func (s Settings) validate() error {
	if err := validateDependencyScope(s.DependencyScope); err != nil {
		return err
	}

	if err := validateFieldUsage(s.FieldUsage); err != nil {
		return err
	}

	_, err := s.policy()

	return err
}

func validateDependencyScope(value string) error {
	switch gomodularity.DependencyScope(value) {
	case gomodularity.DependencyScopeProject,
		gomodularity.DependencyScopeModule,
		gomodularity.DependencyScopeAll:
		return nil
	default:
		return fmt.Errorf(
			"invalid dependency-scope %q (want project, module, or all)",
			value,
		)
	}
}

func validateFieldUsage(value string) error {
	switch gomodularity.FieldUsageMode(value) {
	case gomodularity.FieldUsageDirect, gomodularity.FieldUsageTransitive:
		return nil
	default:
		return fmt.Errorf("invalid field-usage %q (want direct or transitive)", value)
	}
}

func (s Settings) toConfig(selected []gomodularity.MetricName) gomodularity.Config {
	return gomodularity.Config{
		Directory:        s.Directory,
		Patterns:         append([]string(nil), s.Patterns...),
		IncludeTests:     s.Tests,
		IncludeGenerated: s.Generated,
		BuildTags:        append([]string(nil), s.BuildTags...),
		Workers:          s.Workers,
		DependencyScope:  gomodularity.DependencyScope(s.DependencyScope),
		FieldUsageMode:   gomodularity.FieldUsageMode(s.FieldUsage),
		SelectedMetrics:  selected,
		ContinueOnError:  s.ContinueOnError,
	}
}
