package analyzer

import (
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/projectanalysis/application"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/projectanalysis/ports/inbound"
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/application"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/infrastructure/goloader"
)

// NewAnalyzer returns the production analyzer: go/packages fact extraction
// feeding the metric pipeline.
func NewAnalyzer() inbound.Analyzer {
	return application.NewPipeline(typefacts.NewService(goloader.New()))
}
