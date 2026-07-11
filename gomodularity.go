package gomodularity

import (
	"context"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/projectanalysis/ports/inbound"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/infrastructure/analyzer"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/version"
)

// Analyze validates the configuration, runs the analysis pipeline once over
// the configured patterns, and returns a deterministic report. The context
// cancels package loading and metric computation.
func Analyze(ctx context.Context, config Config) (Report, error) {
	cfg := config.withDefaults()
	if err := cfg.validate(); err != nil {
		return Report{}, err
	}

	selected := make([]string, len(cfg.SelectedMetrics))
	for i, m := range cfg.SelectedMetrics {
		selected[i] = string(m)
	}

	result, err := analyzer.NewAnalyzer().Analyze(ctx, inbound.Options{
		Directory:            cfg.Directory,
		Patterns:             cfg.Patterns,
		IncludeTests:         cfg.IncludeTests,
		IncludeGenerated:     cfg.IncludeGenerated,
		BuildTags:            cfg.BuildTags,
		Workers:              cfg.Workers,
		DependencyScope:      string(cfg.DependencyScope),
		FieldUsageTransitive: cfg.FieldUsageMode == FieldUsageTransitive,
		SelectedMetrics:      selected,
		ContinueOnError:      cfg.ContinueOnError,
		Weights:              cfg.ReusabilityWeights,
	})
	if err != nil {
		return Report{}, err
	}

	report := Report{
		SchemaVersion: SchemaVersion,
		Tool:          ToolInfo{Name: ToolName, Version: version.Version},
		Module:        result.ModulePath,
		Packages:      make([]PackageReport, len(result.Packages)),
	}
	for i, pkg := range result.Packages {
		out := PackageReport{
			Path:     pkg.Path,
			Afferent: pkg.Afferent,
			Efferent: pkg.Efferent,
			Funcs:    pkg.Funcs,
			Metrics:  pkg.Metrics,
			Types:    make([]TypeReport, len(pkg.Types)),
		}
		for j, t := range pkg.Types {
			out.Types[j] = TypeReport{
				Name:    t.Name,
				Fields:  t.Fields,
				Methods: t.Methods,
				Metrics: t.Metrics,
			}
		}

		report.Packages[i] = out
	}

	return report, nil
}
