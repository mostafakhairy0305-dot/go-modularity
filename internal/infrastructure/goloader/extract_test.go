package goloader

import (
	"go/ast"
	"testing"

	"github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
)

// TestCountBranchForLoops guards that every for statement contributes to the
// cyclomatic branch count, including a conditionless "for {}", matching the
// gocyclo convention that a for statement is always a decision point.
func TestCountBranchForLoops(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		node ast.Node
	}{
		{"conditionless for {}", &ast.ForStmt{}},
		{"conditional for", &ast.ForStmt{Cond: &ast.Ident{Name: "ok"}}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var branches domain.BranchStats
			countBranch(tt.node, &branches)

			if branches.Fors != 1 {
				t.Errorf("countBranch(%s): Fors = %d, want 1", tt.name, branches.Fors)
			}
		})
	}
}
