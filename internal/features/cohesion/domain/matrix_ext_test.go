package domain_test

import (
	"testing"

	cohesion "github.com/mostafakhairy0305-dot/go-modularity/internal/features/cohesion/domain"
	typefacts "github.com/mostafakhairy0305-dot/go-modularity/internal/features/typefacts/domain"
	"github.com/mostafakhairy0305-dot/go-modularity/internal/shared/bitset"
)

// Black-box: field-set derivation and pair counting from the exported API.
func TestFieldSetsAndPairs(t *testing.T) {
	t.Parallel()

	fs := func(indices ...int) bitset.FieldSet {
		s := bitset.NewFieldSet(2)
		for _, i := range indices {
			s.Set(i)
		}

		return s
	}
	tf := &typefacts.TypeFacts{
		Fields: []typefacts.FieldFacts{{Name: "a"}, {Name: "b"}},
		Methods: []typefacts.MethodFacts{
			{Name: "M1", FieldsUsed: fs(0)},
			{Name: "M2", FieldsUsed: fs(0)}, // shares field 0
		},
	}

	sets := cohesion.EffectiveFieldSets(tf, false)

	pairs := cohesion.CountPairs(sets, len(tf.Fields))
	if pairs.Sharing != 1 || pairs.NonSharing != 0 {
		t.Fatalf("pairs = %+v, want one sharing pair", pairs)
	}

	if got := cohesion.TotalFieldAccesses(sets); got != 2 {
		t.Errorf("total accesses = %d, want 2", got)
	}

	oneCells, distinct := cohesion.ParamMatrix([]typefacts.MethodFacts{
		{Name: "M1", ParamTypeKeys: []string{"int", "string"}},
		{Name: "M2", ParamTypeKeys: []string{"int"}},
	})
	if oneCells != 3 || distinct != 2 {
		t.Errorf("ParamMatrix = (%d,%d), want (3,2)", oneCells, distinct)
	}
}
