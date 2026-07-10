package analyzer

import (
	"testing"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/projectanalysis/ports/inbound"
)

// White-box: the composition root wires up an analyzer satisfying the port.
func TestNewAnalyzerImplementsPort(t *testing.T) {
	t.Parallel()
	var a inbound.Analyzer = NewAnalyzer()
	if a == nil {
		t.Fatal("NewAnalyzer returned nil")
	}
}
