package gomodularity

import (
	"fmt"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/metrics"
)

// DependencyScope selects which import edges count toward package coupling.
type DependencyScope string

const (
	// DependencyScopeProject counts only imports of other analyzed packages.
	DependencyScopeProject DependencyScope = "project"
	// DependencyScopeModule counts imports of packages in the same module.
	DependencyScopeModule DependencyScope = "module"
	// DependencyScopeAll counts every import, including external modules and
	// the standard library.
	DependencyScopeAll DependencyScope = "all"
)

// FieldUsageMode selects how method→field usage is resolved.
type FieldUsageMode string

const (
	// FieldUsageDirect counts only fields a method body accesses directly.
	FieldUsageDirect FieldUsageMode = "direct"
	// FieldUsageTransitive additionally propagates field usage through calls
	// to sibling methods of the same type, to a fixpoint.
	FieldUsageTransitive FieldUsageMode = "transitive"
)

// MetricName identifies a selectable metric.
type MetricName string

// The selectable metrics.
const (
	MetricAMC          MetricName = metrics.MetricAMC
	MetricLCOM1        MetricName = metrics.MetricLCOM1
	MetricLCOM96b      MetricName = metrics.MetricLCOM96b
	MetricCAMC         MetricName = metrics.MetricCAMC
	MetricTCC          MetricName = metrics.MetricTCC
	MetricCBO          MetricName = metrics.MetricCBO
	MetricReusability  MetricName = metrics.MetricReusability
	MetricAbstractness MetricName = metrics.MetricAbstractness
	MetricInstability  MetricName = metrics.MetricInstability
	MetricDistance     MetricName = metrics.MetricDistance
)

// ReusabilityWeights aliases the metrics package's weight type so callers can
// configure weights without importing the metrics package directly.
type ReusabilityWeights = metrics.ReusabilityWeights

// AllMetrics returns every metric name in the fixed rendering order.
func AllMetrics() []MetricName {
	ordered := append(metrics.TypeMetricOrder(), metrics.PackageMetricOrder()...)
	out := make([]MetricName, len(ordered))
	for i, name := range ordered {
		out[i] = MetricName(name)
	}
	return out
}

// DefaultMetrics returns the default display set: every metric except CBO,
// which is a reusability input and is only reported when selected explicitly.
func DefaultMetrics() []MetricName {
	return []MetricName{
		MetricAMC, MetricLCOM1, MetricLCOM96b, MetricCAMC, MetricTCC,
		MetricReusability,
		MetricAbstractness, MetricInstability, MetricDistance,
	}
}

// MetricClosure expands a display set to the full compute set by taking the
// transitive closure over metric dependencies. Selecting a downstream metric
// therefore never produces empty inputs; a metric computed only as a
// dependency is not rendered unless also selected.
func MetricClosure(selected []MetricName) []MetricName {
	names := make([]string, len(selected))
	for i, m := range selected {
		names[i] = string(m)
	}
	closure := metrics.Closure(names)
	out := make([]MetricName, len(closure))
	for i, name := range closure {
		out[i] = MetricName(name)
	}
	return out
}

// Config controls an analysis run. The zero value is usable: defaults are
// applied by Analyze (pattern "./...", module dependency scope, direct field
// usage, the default metric set, and the default reusability weights).
type Config struct {
	// Directory is the working directory for package loading. Empty means the
	// process working directory.
	Directory string
	// Patterns are the package patterns to analyze. Empty means ["./..."].
	Patterns []string
	// IncludeTests also analyzes test files and test packages.
	IncludeTests bool
	// IncludeGenerated also analyzes files carrying the standard
	// "Code generated … DO NOT EDIT." marker.
	IncludeGenerated bool
	// BuildTags are extra build tags for package loading.
	BuildTags []string
	// Workers bounds analysis concurrency. Zero or negative means
	// min(GOMAXPROCS, packageCount).
	Workers int
	// DependencyScope selects the import edges counted by package coupling
	// metrics. Empty means DependencyScopeModule.
	DependencyScope DependencyScope
	// FieldUsageMode selects direct or transitive field-usage resolution.
	// Empty means FieldUsageDirect.
	FieldUsageMode FieldUsageMode
	// SelectedMetrics is the display set. Metrics required as inputs are
	// computed automatically but only rendered when selected. Empty means
	// DefaultMetrics().
	SelectedMetrics []MetricName
	// ContinueOnError proceeds past packages that fail to load or type-check.
	ContinueOnError bool
	// ReusabilityWeights overrides the reusability component weights. The
	// zero value means the defaults (0.35, 0.25, 0.25, 0.15).
	ReusabilityWeights ReusabilityWeights
}

// withDefaults returns a copy of the config with every empty knob replaced by
// its documented default.
func (c Config) withDefaults() Config {
	if len(c.Patterns) == 0 {
		c.Patterns = []string{"./..."}
	}
	if c.DependencyScope == "" {
		c.DependencyScope = DependencyScopeModule
	}
	if c.FieldUsageMode == "" {
		c.FieldUsageMode = FieldUsageDirect
	}
	if len(c.SelectedMetrics) == 0 {
		c.SelectedMetrics = DefaultMetrics()
	}
	if (c.ReusabilityWeights == ReusabilityWeights{}) {
		c.ReusabilityWeights = metrics.DefaultReusabilityWeights()
	}
	return c
}

// validate checks a defaults-applied config.
func (c Config) validate() error {
	switch c.DependencyScope {
	case DependencyScopeProject, DependencyScopeModule, DependencyScopeAll:
	default:
		return fmt.Errorf("invalid dependency scope %q (want project, module, or all)", c.DependencyScope)
	}
	switch c.FieldUsageMode {
	case FieldUsageDirect, FieldUsageTransitive:
	default:
		return fmt.Errorf("invalid field usage mode %q (want direct or transitive)", c.FieldUsageMode)
	}
	known := make(map[MetricName]bool)
	for _, m := range AllMetrics() {
		known[m] = true
	}
	for _, m := range c.SelectedMetrics {
		if !known[m] {
			return fmt.Errorf("unknown metric %q", m)
		}
	}
	for _, p := range c.Patterns {
		if p == "" {
			return fmt.Errorf("empty package pattern")
		}
	}
	if err := c.ReusabilityWeights.Validate(); err != nil {
		return err
	}
	return nil
}
