package analyzer

import (
	"fmt"

	gomodularity "github.com/mostafakhairy0305-dot/go-modularity"
)

// Settings configures the modularity policy analyzer. Fields map to the same
// knobs as the CLI and the gomodularity.Config facade. JSON tags match the
// keys accepted under linters.settings.custom.gomodularity.settings.
type Settings struct {
	// Config is an explicit policy file path. Empty means discover
	// .modularity.yml under Directory, else use the recommended defaults.
	Config string `json:"config"`
	// Directory is the working directory for package loading and policy
	// discovery. Empty means the process working directory.
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
}

func (s Settings) withDefaults() Settings {
	if len(s.Patterns) == 0 {
		s.Patterns = []string{"./..."}
	}

	if s.DependencyScope == "" {
		s.DependencyScope = string(gomodularity.DependencyScopeModule)
	}

	if s.FieldUsage == "" {
		s.FieldUsage = string(gomodularity.FieldUsageDirect)
	}

	return s
}

func (s Settings) validate() error {
	switch gomodularity.DependencyScope(s.DependencyScope) {
	case gomodularity.DependencyScopeProject,
		gomodularity.DependencyScopeModule,
		gomodularity.DependencyScopeAll:
	default:
		return fmt.Errorf(
			"invalid dependency-scope %q (want project, module, or all)",
			s.DependencyScope,
		)
	}

	switch gomodularity.FieldUsageMode(s.FieldUsage) {
	case gomodularity.FieldUsageDirect, gomodularity.FieldUsageTransitive:
	default:
		return fmt.Errorf("invalid field-usage %q (want direct or transitive)", s.FieldUsage)
	}

	return nil
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
